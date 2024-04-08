package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/spf13/cast"
)

type fieldRepo struct {
	db *pgxpool.Pool
}

func NewFieldRepo(db *pgxpool.Pool) storage.FieldRepoI {
	return &fieldRepo{
		db: db,
	}
}

// DONE
func (f *fieldRepo) Create(ctx context.Context, req *nb.CreateFieldRequest) (resp *nb.Field, err error) {

	conn := psqlpool.Get(req.ProjectId)
	fieldId := uuid.NewString()

	if req.Type == "AUTOFILL" {
		autoFillTableSlug := req.AutofillTable

		if strings.Contains(req.AutofillTable, "#") {
			autoFillTableSlug = strings.Split(req.AutofillTable, "#")[0]
		}

		autoFill, err := helper.GetTableByIdSlug(ctx, conn, "", autoFillTableSlug)
		if err != nil {
			return &nb.Field{}, err
		}

		var autoFillFieldSlug = ""

		if strings.Contains(req.AutofillField, ".") {
			splitedTable := strings.Split(strings.Split(req.AutofillField, ".")[0], "_")
			tableSlug := ""
			for i := 0; i < len(splitedTable)-2; i++ {
				tableSlug = tableSlug + "_" + splitedTable[i]
			}
			tableSlug = tableSlug[1:]
			autoFill, err = helper.GetTableByIdSlug(ctx, conn, "", tableSlug)
			if err != nil {
				return &nb.Field{}, err
			}

			autoFillFieldSlug = strings.Split(req.AutofillField, ".")[1]
		} else {
			autoFillFieldSlug = req.AutofillField
		}

		autoFillField, err := helper.GetFieldBySlug(ctx, conn, autoFillFieldSlug, cast.ToString(autoFill["id"]))
		if err != nil && err != pgx.ErrNoRows {
			return &nb.Field{}, err
		}

		attributes, _ := autoFillField["attributes"].([]byte)

		if err := json.Unmarshal(attributes, &req.Attributes); err != nil {
			return &nb.Field{}, err
		}
		req.Type = cast.ToString(autoFillField["type"])
	}

	query := `INSERT INTO "field" (
		id,
		"table_id",
		"required",
		"slug",
		"label",
		"default",
		"type",
		"index",
		"attributes",
		"is_visible",
		autofill_field,
		autofill_table,
		"commit_id",
		"unique",
		"automatic",
		relation_id
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
	)`

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.Field{}, err
	}

	_, err = conn.Exec(ctx, query,
		fieldId,
		req.TableId,
		false,
		req.Slug,
		req.Label,
		req.Default,
		req.Type,
		req.Index,
		attributes,
		req.IsVisible,
		req.AutofillField,
		req.AutofillTable,
		req.CommitId,
		req.Unique,
		req.Automatic,
		req.RelationId,
	)
	if err != nil {
		return &nb.Field{}, err
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	var (
		data          = []byte{}
		tableSlug     string
		layoutId      string
		tabId         string
		sectionId     string
		sectionCount  int32
		sectionFields int32
	)

	err = conn.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
	if err != nil {
		return &nb.Field{}, err
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return &nb.Field{}, err
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $1
	WHERE id = $2
	`

	_, err = conn.Exec(ctx, query, data, req.TableId)
	if err != nil {
		return &nb.Field{}, err
	}

	query = `SELECT guid FROM "role"`

	row, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.Field{}, err
	}

	query = `INSERT INTO "field_permission" (
		"edit_permission",
		"view_permission",
		"table_slug",
		"field_id",
		"label",
		role_id
	) VALUES (true, true, $1, $2, $3, $4)`

	for row.Next() {
		id := ""

		err := row.Scan(&id)
		if err != nil {
			return &nb.Field{}, err
		}

		_, err = conn.Exec(ctx, query, tableSlug, fieldId, req.Label, id)
		if err != nil {
			return &nb.Field{}, err
		}
	}

	query = `SELECT id FROM "layout" WHERE table_id = $1`
	err = conn.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, err
	}

	query = `SELECT id FROM "tab" WHERE section_id = $1 and type = 'section'`
	err = conn.QueryRow(ctx, query, layoutId).Scan(&tabId)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, err
	}

	query = `SELECT id FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
	err = conn.QueryRow(ctx, query, tabId).Scan(&sectionId)
	if err != nil {
		return &nb.Field{}, err
	}

	queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
	err = conn.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, err
	}

	query = `SELECT COUNT(*) FROM "section_fields" WHERE section_id = $1`
	err = conn.QueryRow(ctx, queryCount, tabId).Scan(&sectionFields)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, err
	}

	if sectionFields < 3 {
		query := `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

		_, err = conn.Exec(ctx, query, fieldId, sectionFields+1, req.Label, sectionId)
		if err != nil {
			return &nb.Field{}, err
		}
	} else {
		query = `INSERT INTO "section" (id, order, column, label, table_id, tab_id) VALUES ($1, $2, $3, $4, $5, $6)`

		sectionId = uuid.NewString()

		_, err = conn.Exec(ctx, query, sectionId, sectionCount+1, "SINGLE", "Info", req.TableId, tabId)
		if err != nil {
			return &nb.Field{}, err
		}

		query = `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

		_, err = conn.Exec(ctx, query, fieldId, 1, req.Label, sectionId)
		if err != nil {
			return &nb.Field{}, err
		}
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

// DONE
func (f *fieldRepo) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error) {

	conn := psqlpool.Get(req.ProjectId)

	resp = &nb.Field{}

	attributes := []byte{}

	query := `SELECT 
		id,
		"table_id",
		"required",
		"slug",
		"label",
		"default",
		"type",
		"index",
		"attributes",
		"is_visible",
		autofill_field,
		autofill_table,
		"commit_id",
		"unique",
		"automatic",
		relation_id
	FROM "field" WHERE id = $1`

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.TableId,
		&resp.Required,
		&resp.Slug,
		&resp.Label,
		&resp.Default,
		&resp.Type,
		&resp.Index,
		&attributes,
		&resp.IsVisible,
		&resp.AutofillField,
		&resp.AutofillTable,
		&resp.CommitId,
		&resp.Unique,
		&resp.Automatic,
		&resp.RelationId,
	)
	if err != nil {
		return &nb.Field{}, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return &nb.Field{}, err
	}

	return resp, nil
}

func (f *fieldRepo) GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error) {
	conn := psqlpool.Get(req.ProjectId)

	getTable, err := helper.GetTableByIdSlug(ctx, conn, req.TableId, req.TableSlug)
	if err != nil {
		return &nb.GetAllFieldsResponse{}, err
	}

	req.TableId = cast.ToString(getTable["id"])
	req.TableSlug = cast.ToString(getTable["slug"])

	query := `SELECT 
		id,
		"table_id",
		"required",
		"slug",
		"label",
		"default",
		"type",
		"index",
		"attributes",
		"is_visible",
		autofill_field,
		autofill_table,
		"commit_id",
		"unique",
		"automatic",
		relation_id
	FROM "field" WHERE name ~* $1 and table_id = $2 LIMIT $3 OFFSET $4`

	rows, err := conn.Query(ctx, query, req.Search, req.TableId, req.Limit, req.Offset)
	if err != nil {
		return &nb.GetAllFieldsResponse{}, err
	}

	for rows.Next() {
		var (
			field      = &nb.Field{}
			attributes = []byte{}
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.Required,
			&field.Slug,
			&field.Label,
			&field.Default,
			&field.Type,
			&field.Index,
			&attributes,
			&field.IsVisible,
			&field.AutofillField,
			&field.AutofillTable,
			&field.CommitId,
			&field.Unique,
			&field.Automatic,
			&field.RelationId,
		)
		if err != nil {
			return &nb.GetAllFieldsResponse{}, err
		}

		resp.Fields = append(resp.Fields, field)
	}

	query = `SELECT COUNT(*) FROM "field" WHERE name ~* $1 and table_id = $2`

	err = conn.QueryRow(ctx, query, req.Search, req.TableId).Scan(&resp.Count)
	if err != nil {
		return &nb.GetAllFieldsResponse{}, err
	}

	if req.WithManyRelation {
		query = `
		SELECT table_from FROM "relation" WHERE 
			(table_to = $1 AND type = 'Many2One')
			OR
			(table_to = $1 AND type = 'Many2Many')
		`

		rows, err := conn.Query(ctx, query, req.TableSlug)
		if err != nil {
			return &nb.GetAllFieldsResponse{}, err
		}

		defer rows.Close()

		query = `SELECT id, slug, type, attributes FROM "field" WHERE table_id = $1`

		// queryR := `SELECT view_fields FROM "relation" WHERE table_from = $1 AND table_to = $2`

		for rows.Next() {
			tableFrom := ""

			err = rows.Scan(&tableFrom)
			if err != nil {
				return &nb.GetAllFieldsResponse{}, err
			}

			relationTable, err := helper.GetTableByIdSlug(ctx, conn, "", tableFrom)
			if err != nil {
				return &nb.GetAllFieldsResponse{}, err
			}

			fieldRows, err := conn.Query(ctx, query, cast.ToString(relationTable["id"]))
			if err != nil {
				return &nb.GetAllFieldsResponse{}, err
			}

			for fieldRows.Next() {
				var (
					id, slug, ftype string
					attributes      []byte
					// viewFields      []string
				)

				err := fieldRows.Scan(
					&id,
					&slug,
					&ftype,
					&attributes,
				)
				if err != nil {
					return &nb.GetAllFieldsResponse{}, err
				}

				// ! skipped

				// if ftype == "LOOKUP" {
				// 	view_fildes := []map[string]interface{}{}

				// 	err = conn.QueryRow(ctx, queryR, cast.ToString(relationTable["slug"]), slug[:len(slug)-3]).Scan(
				// 		pq.Array(&viewFields),
				// 	)
				// 	if err != nil && err != pgx.ErrNoRows {
				// 		return &nb.GetAllFieldsResponse{}, err
				// 	}

				// 	for _, view_field := range viewFields {
				// 		field, err := f.GetByID(ctx, &nb.FieldPrimaryKey{Id: view_field, ProjectId: req.ProjectId})
				// 		if err != nil {
				// 			return &nb.GetAllFieldsResponse{}, err
				// 		}

				// 	}
				// }
			}

		}

	}

	return resp, nil
}

func (f *fieldRepo) GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error) {
	// conn := psqlpool.Get(req.ProjectId)

	// Skipped ...

	return &nb.AllFields{}, nil
}

func (f *fieldRepo) Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error) {
	conn := psqlpool.Get(req.ProjectId)

	resp, err = f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	if err != nil {
		return &nb.Field{}, err
	}

	if resp.IsSystem {
		return &nb.Field{}, fmt.Errorf("error you can't update this field its system field")
	}

	if req.Type == "AUTOFILL" && req.AutofillField != "" && req.AutofillTable != "" {
		// var autoFillTableSlug = req.AutofillTable

		// if strings.Contains(req.AutofillTable, "#") {
		// 	// autoFillTableSlug = strings.Split(req.AutofillTable, "#")[0]
		// }

		var (
			autoFillFieldSlug = ""
			attributes        = []byte{}
			autoFieldtype     = ""
		)

		// there should be code this table version

		if strings.Contains(req.AutofillField, ".") {
			var (
				splitedAutofillField = strings.Split(req.AutofillField, ".")
				splitedTable         = strings.Split(splitedAutofillField[0], "_")
				tableSlug            = ""
			)
			for i := 0; i < len(splitedTable)-2; i++ {
				tableSlug = tableSlug + "_" + splitedTable[i]
			}

			// tableSlug = tableSlug[1:]
			autoFillFieldSlug = splitedAutofillField[1]
		} else {
			autoFillFieldSlug = req.AutofillField
		}

		query := `SELECT type, attributes FROM "field" WHERE slug = $1 and table_id = $2`

		err = conn.QueryRow(ctx, query, autoFillFieldSlug, "").Scan(
			&autoFieldtype,
			&attributes,
		)
		if err != nil && err != pgx.ErrNoRows {
			return &nb.Field{}, err
		}

		if autoFieldtype != "" {
			req.Type = autoFieldtype
			if err := json.Unmarshal(attributes, &req.Attributes); err != nil {
				return &nb.Field{}, err
			}
		}
	}

	attributes := json.Marshaler(req.Attributes)

	query := `UPDATE "field" SET
		"required" = $2,
		"slug" = $3,
		"label" = $4,
		"default" = $5,
		"type" = $6,
		"index" = $7,
		"attributes" = $8,
		"is_visible" = $9,
		autofill_field = $10,
		autofill_table = $11,
		"commit_id" = $12,
		"unique" = $13,
		"automatic" = $14,
		relation_id = $15,
		"table_id" = $16,
	WHERE id = $1
	`

	_, err = conn.Exec(ctx, query, req.Id,
		req.Required,
		req.Slug,
		req.Label,
		req.Default,
		req.Type,
		req.Index,
		attributes,
		req.IsVisible,
		req.AutofillField,
		req.AutofillTable,
		req.CommitId,
		req.Unique,
		req.Automatic,
		req.RelationId,
		req.TableId,
	)
	if err != nil {
		return &nb.Field{}, err
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (f *fieldRepo) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error {
	conn := psqlpool.Get(req.ProjectId)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	query := `UPDATE "field" SET is_search = $1 WHERE id = $2`

	for _, val := range req.Fields {
		_, err = tx.Exec(ctx, query, val.IsSearch, val.Id)
		if err != nil {
			tx.Rollback(ctx)
			return err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	query = `SELECT is_changed_by_host FROM "table" where slug = $1`

	var (
		data = []byte{}
	)

	err = conn.QueryRow(ctx, query, req.TableSlug).Scan(&data)
	if err != nil {
		return err
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return err
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $2
	WHERE slug = $1
	`

	_, err = conn.Exec(ctx, query, req.TableSlug, data)
	if err != nil {
		return err
	}

	return nil
}

func (f *fieldRepo) Delete(ctx context.Context, req *nb.FieldPrimaryKey) error {

	conn := psqlpool.Get(req.ProjectId)

	var (
		isSystem            bool
		tableId             string
		viewId              string
		columns, newColumns []string
		isExists            bool
		tableSlug           string
	)

	query := `SELECT is_system, table_id FROM "field" WHERE id = $1`

	err := conn.QueryRow(ctx, query, req.Id).Scan(&isSystem, &tableId)
	if err != nil {
		return err
	}

	if isSystem {
		return fmt.Errorf("you can't delete! this filed is system field ")
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	var (
		data = []byte{}
	)

	err = conn.QueryRow(ctx, query, tableId).Scan(&data, &tableSlug)
	if err != nil {
		return err
	}

	query = `DELETE FROM "field_permission" WHERE field_id = $1`

	_, err = conn.Exec(ctx, query, req.Id)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	query = `SELECT id, columns FROM "view" WHERE table_slug = $1 `

	err = conn.QueryRow(ctx, query, tableSlug).Scan(&viewId, pq.Array(&columns))
	if err != nil {
		return err
	}

	if len(columns) > 0 {
		for _, c := range columns {
			if c == req.Id {
				isExists = true
				continue
			} else {
				newColumns = append(newColumns, c)
			}
		}

		if isExists {
			query = `UPDATE "field_permission" SET columns = $1`

			_, err = conn.Exec(ctx, query, pq.Array(newColumns))
			if err != nil {
				return err
			}
		}
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return err
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $2
	WHERE id = $1
	`

	_, err = conn.Exec(ctx, query, tableId, data)
	if err != nil {
		return err
	}

	return nil
}
