package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/opentracing/opentracing-go"

	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"
)

// GetProjectTablesSchema returns column info for all user-created tables (non-system) in the project DB.
func (r *aiChatRepo) GetProjectTablesSchema(ctx context.Context, resourceEnvId string) ([]storage.DBTableSchema, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetProjectTablesSchema")
	defer span.Finish()

	conn, err := psqlpool.Get(resourceEnvId)
	if err != nil {
		return nil, err
	}

	// Get user table slugs from the "table" metadata table (excludes system tables)
	slugQuery := `
		SELECT slug FROM "table" 
		WHERE deleted_at IS NULL 
		AND (is_system = false OR slug IN ('role', 'client_type'))
	`
	slugRows, err := conn.Query(ctx, slugQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query table slugs: %w", err)
	}
	defer slugRows.Close()

	var tableSlugs []string
	for slugRows.Next() {
		var slug string
		if err := slugRows.Scan(&slug); err != nil {
			return nil, fmt.Errorf("failed to scan table slug: %w", err)
		}
		tableSlugs = append(tableSlugs, slug)
	}
	if err := slugRows.Err(); err != nil {
		return nil, fmt.Errorf("slug rows error: %w", err)
	}

	if len(tableSlugs) == 0 {
		return nil, nil
	}

	// Query information_schema for these tables
	query := `
		SELECT c.table_name, c.column_name, c.data_type, c.is_nullable
		FROM information_schema.columns c
		JOIN information_schema.tables t 
		  ON c.table_name = t.table_name AND c.table_schema = t.table_schema
		WHERE c.table_schema = 'public'
		  AND t.table_type = 'BASE TABLE'
		  AND c.table_name = ANY($1)
		ORDER BY c.table_name, c.ordinal_position
	`

	rows, err := conn.Query(ctx, query, tableSlugs)
	if err != nil {
		return nil, fmt.Errorf("failed to query information_schema: %w", err)
	}
	defer rows.Close()

	tableMap := make(map[string]*storage.DBTableSchema)
	for rows.Next() {
		var tableName, colName, dataType, isNullable string
		if err := rows.Scan(&tableName, &colName, &dataType, &isNullable); err != nil {
			return nil, fmt.Errorf("failed to scan column info: %w", err)
		}

		tbl, ok := tableMap[tableName]
		if !ok {
			tbl = &storage.DBTableSchema{TableName: tableName}
			tableMap[tableName] = tbl
		}

		tbl.Columns = append(tbl.Columns, storage.DBColumn{
			ColumnName: colName,
			DataType:   dataType,
			IsNullable: isNullable,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	result := make([]storage.DBTableSchema, 0, len(tableMap))
	for _, tbl := range tableMap {
		result = append(result, *tbl)
	}

	// Sort by table name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].TableName < result[j].TableName
	})

	return result, nil
}

// ExecuteCrudOperation executes a dynamic CRUD operation using parameterized queries.
// CRITICAL: Always uses $N placeholders, never string interpolation for values.
func (r *aiChatRepo) ExecuteCrudOperation(ctx context.Context, resourceEnvId string, op storage.CrudOperationReq) (*storage.CrudOperationResult, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.ExecuteCrudOperation")
	defer span.Finish()

	conn, err := psqlpool.Get(resourceEnvId)
	if err != nil {
		return nil, err
	}

	// Validate table name exists in metadata (prevent injection via table name)
	var tableExists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM "table" WHERE slug = $1 AND deleted_at IS NULL)`
	if err := conn.QueryRow(ctx, checkQuery, op.Table).Scan(&tableExists); err != nil {
		return nil, fmt.Errorf("failed to validate table: %w", err)
	}
	if !tableExists {
		return nil, fmt.Errorf("table %q not found", op.Table)
	}

	var data, where map[string]any
	if op.DataJSON != "" {
		if err := json.Unmarshal([]byte(op.DataJSON), &data); err != nil {
			return nil, fmt.Errorf("invalid data_json: %w", err)
		}
	}
	if op.WhereJSON != "" {
		if err := json.Unmarshal([]byte(op.WhereJSON), &where); err != nil {
			return nil, fmt.Errorf("invalid where_json: %w", err)
		}
	}

	switch op.Operation {
	case "insert":
		return r.execInsert(ctx, conn, op.Table, data)
	case "update":
		return r.execUpdate(ctx, conn, op.Table, data, where)
	case "delete":
		return r.execDelete(ctx, conn, op.Table, where)
	case "select":
		return r.execSelect(ctx, conn, op.Table, where)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", op.Operation)
	}
}

func (r *aiChatRepo) execInsert(ctx context.Context, conn *psqlpool.Pool, table string, data map[string]any) (*storage.CrudOperationResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for insert")
	}

	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	args := make([]any, 0, len(data))

	i := 1
	for col, val := range data {
		cols = append(cols, fmt.Sprintf(`"%s"`, col))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	res, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}

	return &storage.CrudOperationResult{
		RowsAffected: int32(res.RowsAffected()),
	}, nil
}

func (r *aiChatRepo) execUpdate(ctx context.Context, conn *psqlpool.Pool, table string, data, where map[string]any) (*storage.CrudOperationResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}
	if len(where) == 0 {
		return nil, fmt.Errorf("no where conditions provided for update (safety)")
	}

	setClauses := make([]string, 0, len(data))
	args := make([]any, 0, len(data)+len(where))
	i := 1

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, col, i))
		args = append(args, val)
		i++
	}

	whereClauses := make([]string, 0, len(where))
	for col, val := range where {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, col, i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf(`UPDATE "%s" SET %s WHERE %s`,
		table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "),
	)

	res, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("update failed: %w", err)
	}

	return &storage.CrudOperationResult{
		RowsAffected: int32(res.RowsAffected()),
	}, nil
}

func (r *aiChatRepo) execDelete(ctx context.Context, conn *psqlpool.Pool, table string, where map[string]any) (*storage.CrudOperationResult, error) {
	if len(where) == 0 {
		return nil, fmt.Errorf("no where conditions provided for delete (safety)")
	}

	whereClauses := make([]string, 0, len(where))
	args := make([]any, 0, len(where))
	i := 1

	for col, val := range where {
		whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, col, i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf(`DELETE FROM "%s" WHERE %s`, table, strings.Join(whereClauses, " AND "))

	res, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("delete failed: %w", err)
	}

	return &storage.CrudOperationResult{
		RowsAffected: int32(res.RowsAffected()),
	}, nil
}

func (r *aiChatRepo) execSelect(ctx context.Context, conn *psqlpool.Pool, table string, where map[string]any) (*storage.CrudOperationResult, error) {
	args := make([]any, 0, len(where))
	var whereClause string

	if len(where) > 0 {
		whereClauses := make([]string, 0, len(where))
		i := 1
		for col, val := range where {
			whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, col, i))
			args = append(args, val)
			i++
		}
		whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query := fmt.Sprintf(`SELECT * FROM "%s"%s LIMIT 50`, table, whereClause)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	defer rows.Close()

	fieldDescs := rows.FieldDescriptions()
	colNames := make([]string, len(fieldDescs))
	for i, fd := range fieldDescs {
		colNames[i] = string(fd.Name)
	}

	var results []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("failed to read row values: %w", err)
		}

		row := make(map[string]any, len(colNames))
		for i, col := range colNames {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	resultJSON, err := json.Marshal(results)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal results: %w", err)
	}

	return &storage.CrudOperationResult{
		ResultJSON:   string(resultJSON),
		RowsAffected: int32(len(results)),
	}, nil
}
