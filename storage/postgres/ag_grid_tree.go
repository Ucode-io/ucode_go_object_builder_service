package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cast"
)

func (o *objectBuilderRepo) AgGridTree(ctx context.Context, req *nb.CommonMessage) (*nb.CommonMessage, error) {
	conn, err := psqlpool.Get(req.GetProjectId())
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to get database connection")
	}

	fields, filterValue, err := parseAndValidateRequest(req)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to parse request")
	}

	lookupFields, relatedTables, err := getLookupFields(ctx, conn, req.TableSlug)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to get lookup fields")
	}

	results, err := buildAndExecuteQuery(ctx, conn, req.TableSlug, fields, lookupFields, relatedTables, filterValue)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to build and execute query")
	}

	response := map[string]any{"response": results}
	respData, err := helper.ConvertMapToStruct(response)
	if err != nil {
		return nil, helper.HandleDatabaseError(err, o.logger, "AgGridTree: Failed to convert response")
	}

	return &nb.CommonMessage{
		TableSlug: req.TableSlug,
		ProjectId: req.ProjectId,
		Data:      respData,
	}, nil
}

func parseAndValidateRequest(req *nb.CommonMessage) ([]string, string, error) {
	var params map[string]any

	paramBody, err := json.Marshal(req.Data)
	if err != nil {
		return nil, "", err
	}
	if err := json.Unmarshal(paramBody, &params); err != nil {
		return nil, "", err
	}

	fields, ok := params["fields"].([]any)
	if !ok {
		return nil, "", fmt.Errorf("fields not found or invalid")
	}

	childField := req.TableSlug + "_id"
	var filterValue string
	if filterValues, ok := params[childField].([]any); ok && len(filterValues) > 0 {
		if filterValues[0] == nil {
			filterValue = ""
		} else {
			val := cast.ToString(filterValues[0])
			filterValue = val
		}
	}

	return cast.ToStringSlice(fields), filterValue, nil
}

func getLookupFields(ctx context.Context, conn *psqlpool.Pool, tableSlug string) ([]string, []string, error) {
	const fieldQuery = `
		SELECT f.slug, f.type 
		FROM field f 
		JOIN "table" t ON t.id = f.table_id 
		WHERE t.slug = $1 AND type IN ('LOOKUP', 'LOOKUPS')`

	rows, err := conn.Query(ctx, fieldQuery, tableSlug)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var lookupFields, relatedTables []string

	for rows.Next() {
		var slug, ftype string
		if err := rows.Scan(&slug, &ftype); err != nil {
			return nil, nil, err
		}

		if strings.Contains(slug, "_id") {
			lookupFields = append(lookupFields, slug)
			tableName := strings.TrimSuffix(slug, "_id")
			relatedTables = append(relatedTables, tableName)
		}
	}

	return lookupFields, relatedTables, nil
}

func buildAndExecuteQuery(ctx context.Context, conn *psqlpool.Pool, tableSlug string, fields []string, lookupFields, relatedTables []string, filterValue string) ([]map[string]any, error) {
	query := buildRecursiveQuery(tableSlug, fields, lookupFields, relatedTables, filterValue)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}

	var results []map[string]any

	for rows.Next() {
		row, err := processRow(rows, columns)
		if err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func buildRecursiveQuery(tableSlug string, fields, lookupFields, relatedTables []string, filterValue string) string {
	childField := tableSlug + "_id"
	baseSelect := buildSelectClause("parent_node", fields, lookupFields, relatedTables)
	recursiveSelect := buildSelectClause("child_node", fields, lookupFields, relatedTables)

	var lookupDataSelects strings.Builder
	for _, field := range lookupFields {
		lookupDataSelects.WriteString(fmt.Sprintf(", h.%s_data", field))
	}

	var baseFilterCondition string
	if filterValue == "" {
		baseFilterCondition = fmt.Sprintf("parent_node.%s IS NULL", childField)
	} else {
		baseFilterCondition = fmt.Sprintf("parent_node.%s = '%s'", childField, filterValue)
	}

	query := fmt.Sprintf(`
        WITH RECURSIVE hierarchy AS (
            -- Base case: select nodes matching our filter
            SELECT 
                %s,
                ARRAY[parent_node.guid] AS path,
                EXISTS (
                    SELECT 1 FROM %s child 
                    WHERE child.%s = parent_node.guid
                ) AS has_child,
                parent_node.%s AS original_parent_id
            FROM %s parent_node
            WHERE %s

            UNION ALL

            -- Recursive case: select all descendants
            SELECT 
                %s,
                parent.path || child_node.guid AS path,
                EXISTS (
                    SELECT 1 FROM %s grandchild 
                    WHERE grandchild.%s = child_node.guid
                ) AS has_child,
                parent.original_parent_id
            FROM %s child_node
            INNER JOIN hierarchy parent ON child_node.%s = parent.guid
        )
        SELECT %s, h.path, h.has_child%s FROM hierarchy h
        WHERE h.%s %s`,
		baseSelect,
		tableSlug,
		childField,
		childField,
		tableSlug,
		baseFilterCondition,
		recursiveSelect,
		tableSlug,
		childField,
		tableSlug,
		childField,
		joinColumnsWithPrefix(fields, "h."),
		lookupDataSelects.String(),
		childField,
		ifElse(filterValue == "", "IS NULL", fmt.Sprintf("= '%s'", filterValue)),
	)

	return query
}

func ifElse(condition bool, trueVal, falseVal string) string {
	if condition {
		return trueVal
	}
	return falseVal
}

func buildSelectClause(prefix string, fields []string, lookupFields, relatedTables []string) string {
	var sb strings.Builder
	sb.WriteString(joinColumnsWithPrefix(fields, prefix+"."))

	for i, field := range lookupFields {
		relationAlias := fmt.Sprintf("related_table%d", i+1)
		sb.WriteString(fmt.Sprintf(`, 
            (SELECT row_to_json(%s) 
             FROM "%s" %s 
             WHERE %s.guid = %s.%s
            ) AS %s_data`,
			relationAlias,
			relatedTables[i],
			relationAlias,
			relationAlias,
			prefix,
			field,
			field))
	}

	return sb.String()
}

func processRow(rows pgx.Rows, columns []string) (map[string]any, error) {
	values, err := rows.Values()
	if err != nil {
		return nil, err
	}

	rowMap := make(map[string]any)
	for i, colName := range columns {
		val := values[i]
		switch {
		case val == nil:
			rowMap[colName] = nil
		case (colName == config.GUID || strings.Contains(colName, config.ID)) && !strings.Contains(colName, "_id_data"):
			if arr, ok := val.([16]byte); ok {
				rowMap[colName] = helper.ConvertGuid(arr)
			}
		case colName == config.PATH:
			if arr, ok := val.([]any); ok {
				rowMap[colName] = convertGuidPath(arr)
			}
		case colName == "has_child":
			if b, ok := val.(bool); ok {
				rowMap[colName] = b
			}
		default:
			rowMap[colName] = val
		}
	}
	return rowMap, nil
}

func convertGuidPath(path []any) []string {
	var guidList []string
	for _, guid := range path {
		if guidArr, ok := guid.([16]byte); ok {
			guidList = append(guidList, helper.ConvertGuid(guidArr))
		}
	}
	return guidList
}

func joinColumnsWithPrefix(columns []string, prefix string) string {
	prefixed := make([]string, len(columns))
	for i, col := range columns {
		prefixed[i] = prefix + col
	}
	return strings.Join(prefixed, ", ")
}
