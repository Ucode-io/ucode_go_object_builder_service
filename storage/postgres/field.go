package postgres

import (
	"context"
	"database/sql"
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

	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Field{}, err
	}

	req.Slug = strings.ToLower(req.GetSlug())

	// ! FIELD_TYPE AUTOFILL DOESN'T USE IN NEW VERSION

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
		"unique",
		"automatic"
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
	)`

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	_, err = tx.Exec(ctx, query,
		req.GetId(),
		req.GetTableId(),
		false,
		req.GetSlug(),
		req.GetLabel(),
		req.GetDefault(),
		req.GetType(),
		req.GetIndex(),
		attributes,
		req.GetIsVisible(),
		req.GetAutofillField(),
		req.GetAutofillTable(),
		req.GetUnique(),
		req.GetAutomatic(),
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	query = `SELECT is_changed_by_host, slug FROM "table" WHERE id = $1`

	var (
		data         = []byte{}
		tableSlug    string
		layoutId     string
		tabId        string
		sectionId    string
		sectionCount int32
	)

	err = tx.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	query = `ALTER TABLE ` + tableSlug + ` ADD COLUMN ` + req.Slug + " " + helper.GetDataType(req.Type)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
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

	_, err = tx.Exec(ctx, query, data, req.TableId)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	query = `SELECT guid FROM "role"`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}
	defer rows.Close()

	ids := []string{}

	for rows.Next() {
		id := ""

		err := rows.Scan(&id)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}

		ids = append(ids, id)
	}

	query = `INSERT INTO "field_permission" (
		"edit_permission",
		"view_permission",
		"table_slug",
		"field_id",
		"label",
		role_id
	) VALUES (true, true, $1, $2, $3, $4)`

	for _, id := range ids {

		_, err = tx.Exec(ctx, query, tableSlug, req.Id, req.Label, id)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	}

	query = `SELECT id FROM "layout" WHERE table_id = $1`
	err = tx.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
	if err != nil && err != pgx.ErrNoRows {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	query = `SELECT id FROM "tab" WHERE "layout_id" = $1 and type = 'section'`
	err = tx.QueryRow(ctx, query, layoutId).Scan(&tabId)
	if err != nil && err != pgx.ErrNoRows {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	var (
		body   = []byte{}
		fields = []SectionFields{}
	)

	query = `SELECT id, fields FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
	err = tx.QueryRow(ctx, query, tabId).Scan(&sectionId, &body)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
	err = tx.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
	if err != nil && err != pgx.ErrNoRows {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	if err := json.Unmarshal(body, &fields); err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	if len(fields) < 3 {

		query := `UPDATE "section" SET fields = $2 WHERE id = $1`

		fields = append(fields, SectionFields{
			Id:    req.Id,
			Order: len(fields) + 1,
		})

		reqBody, err := json.Marshal(fields)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}

		_, err = tx.Exec(ctx, query, sectionId, reqBody)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	} else {
		query = `INSERT INTO "section" ("order", "column", label, table_id, tab_id, fields) VALUES ($1, $2, $3, $4, $5, $6)`

		sectionId = uuid.NewString()

		fields := []SectionFields{
			{
				Id:    req.Id,
				Order: 1,
			},
		}

		reqBody, err := json.Marshal(fields)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}

		_, err = tx.Exec(ctx, query, sectionCount+1, "SINGLE", "Info", req.TableId, tabId, reqBody)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	}

	if req.Type == "INCREMENT_ID" {
		query = `INSERT INTO "incrementseqs" (field_slug, table_slug) VALUES ($1, $2)`

		_, err = tx.Exec(ctx, query, req.Slug, tableSlug)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}

	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	// return &nb.Field{}, nil
	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

// DONE
func (f *fieldRepo) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error) {

	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	resp = &nb.Field{}

	var (
		attributes     = []byte{}
		relationIdNull sql.NullString
	)
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
		&resp.Unique,
		&resp.Automatic,
		&relationIdNull,
	)
	if err != nil {
		return &nb.Field{}, err
	}

	resp.RelationId = relationIdNull.String

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return &nb.Field{}, err
	}

	return resp, nil
}

func (f *fieldRepo) GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	resp = &nb.GetAllFieldsResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	getTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Id: req.TableId, Slug: req.TableSlug})
	if err != nil {
		return &nb.GetAllFieldsResponse{}, err
	}

	req.TableId = cast.ToString(getTable["id"])
	req.TableSlug = cast.ToString(getTable["slug"])

	query := `SELECT 
		"id",
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
		"unique",
		"automatic",
		relation_id
	FROM "field" WHERE table_id = $1 LIMIT $2 OFFSET $3`

	rows, err := conn.Query(ctx, query, req.TableId, req.Limit, req.Offset)
	if err != nil {
		return &nb.GetAllFieldsResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			field             = nb.Field{}
			attributes        sql.NullString
			autoFillFieldNull sql.NullString
			autoFillTableNull sql.NullString
			relationIdNull    sql.NullString
			defaultStr, index sql.NullString
		)

		err = rows.Scan(
			&field.Id,
			&field.TableId,
			&field.Required,
			&field.Slug,
			&field.Label,
			&defaultStr,
			&field.Type,
			&index,
			&attributes,
			&field.IsVisible,
			&autoFillFieldNull,
			&autoFillTableNull,
			&field.Unique,
			&field.Automatic,
			&relationIdNull,
		)
		if err != nil {
			return &nb.GetAllFieldsResponse{}, err
		}

		if attributes.Valid {
			err := json.Unmarshal([]byte(attributes.String), &field.Attributes)
			if err != nil {
				return &nb.GetAllFieldsResponse{}, err
			}
		}

		field.AutofillField = autoFillFieldNull.String
		field.AutofillTable = autoFillTableNull.String
		field.RelationField = relationIdNull.String
		field.Default = defaultStr.String
		field.Index = index.String

		resp.Fields = append(resp.Fields, &field)
	}

	query = `SELECT COUNT(*) FROM "field" WHERE table_id = $1`

	err = conn.QueryRow(ctx, query, req.TableId).Scan(&resp.Count)
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

			// relationTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Id: "", Slug: tableFrom})
			// if err != nil {
			// 	return &nb.GetAllFieldsResponse{}, err
			// }

			// fieldRows, err := conn.Query(ctx, query, cast.ToString(relationTable["id"]))
			// if err != nil {
			// 	return &nb.GetAllFieldsResponse{}, err
			// }

			// for fieldRows.Next() {
			// 	var (
			// 		id, slug, ftype string
			// 		attributes      []byte
			// 		// viewFields      []string
			// 	)

			// 	err := fieldRows.Scan(
			// 		&id,
			// 		&slug,
			// 		&ftype,
			// 		&attributes,
			// 	)
			// 	if err != nil {
			// 		return &nb.GetAllFieldsResponse{}, err
			// 	}

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

	// }

	return resp, nil
}

func (f *fieldRepo) GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error) {
	// conn := psqlpool.Get(req.ProjectId)

	// Skipped ...

	return &nb.AllFields{}, nil
}

func (f *fieldRepo) Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	resp, err = f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	if resp.IsSystem {
		tx.Rollback(ctx)
		return &nb.Field{}, fmt.Errorf("error you can't update this field its system field")
	}

	// ! FIELD_TYPE AUTOFILL DOESN'T USE IN NEW VERSION

	// if req.Type == "AUTOFILL" && req.AutofillField != "" && req.AutofillTable != "" {
	// var autoFillTableSlug = req.AutofillTable

	// if strings.Contains(req.AutofillTable, "#") {
	// 	// autoFillTableSlug = strings.Split(req.AutofillTable, "#")[0]
	// }

	// 	var (
	// 		autoFillFieldSlug string
	// 		attributes        = []byte{}
	// 		autoFieldtype     string
	// 	)

	// 	// there should be code this table version

	// 	if strings.Contains(req.AutofillField, ".") {
	// 		var (
	// 			splitedAutofillField = strings.Split(req.AutofillField, ".")
	// 			splitedTable         = strings.Split(splitedAutofillField[0], "_")
	// 			tableSlug            = ""
	// 		)
	// 		for i := 0; i < len(splitedTable)-2; i++ {
	// 			tableSlug = tableSlug + "_" + splitedTable[i]
	// 		}

	// 		// tableSlug = tableSlug[1:]
	// 		autoFillFieldSlug = splitedAutofillField[1]
	// 	} else {
	// 		autoFillFieldSlug = req.AutofillField
	// 	}

	// 	query := `SELECT type, attributes FROM "field" WHERE slug = $1 and table_id = $2`

	// 	err = conn.QueryRow(ctx, query, autoFillFieldSlug, "").Scan(
	// 		&autoFieldtype,
	// 		&attributes,
	// 	)
	// 	if err != nil && err != pgx.ErrNoRows {
	// 		return &nb.Field{}, err
	// 	}

	// 	if autoFieldtype != "" {
	// 		req.Type = autoFieldtype
	// 		if err := json.Unmarshal(attributes, &req.Attributes); err != nil {
	// 			return &nb.Field{}, err
	// 		}
	// 	}
	// }

	tableSlug := ""

	query := `SELECT slug FROM "table" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.TableId).Scan(&tableSlug)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	query = `UPDATE "field" SET
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
		"unique" = $12,
		"automatic" = $13
	WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, req.Id,
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
		req.Unique,
		req.Automatic,
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	if resp.Type != req.Type {

		// * QUERY -> Try to change columns type if error it changes value to new type's default value
		// * OLD: VARCHAR - "John Doe" NEW: FLOAT - 0.0

		fieldType := helper.GetDataType(req.Type)
		regExp := helper.GetRegExp(fieldType)
		defaultValue := helper.GetDefault(fieldType)

		query = fmt.Sprintf(`ALTER TABLE %s 
		ALTER COLUMN %s TYPE %s
		USING CASE WHEN %s ~ '%s' THEN %s::%s ELSE %v END;`, tableSlug, req.Slug, fieldType, req.Slug, regExp, req.Slug, fieldType, defaultValue)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	}

	if resp.Slug != req.Slug {
		query = fmt.Sprintf(`ALTER TABLE %s
		RENAME COLUMN %s TO %s;`, tableSlug, resp.Slug, req.Slug)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.Field{}, err
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (f *fieldRepo) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	query := `UPDATE "field" SET is_search = $1 WHERE id = $2`

	for _, val := range req.Fields {
		_, err = tx.Exec(ctx, query, val.IsSearch, val.Id)
		if err != nil {
			return err
		}
	}

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

	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	var (
		isSystem            bool
		tableId             string
		viewId              string
		columns, newColumns []string
		isExists            bool
		tableSlug           string
		fieldSlug           string
	)

	query := `SELECT is_system, table_id, slug FROM "field" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.Id).Scan(&isSystem, &tableId, &fieldSlug)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if isSystem {
		return fmt.Errorf("you can't delete! this filed is system field ")
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	var (
		data = []byte{}
	)

	err = tx.QueryRow(ctx, query, tableId).Scan(&data, &tableSlug)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `DELETE FROM "field_permission" WHERE field_id = $1`

	_, err = conn.Exec(ctx, query, req.Id)
	if err != nil && err != pgx.ErrNoRows {
		tx.Rollback(ctx)
		return err
	}

	query = `SELECT id, columns FROM "view" WHERE table_slug = $1 `

	err = tx.QueryRow(ctx, query, tableSlug).Scan(&viewId, &columns)
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
			query = `UPDATE "view" SET columns = $1 WHERE id = $2`

			_, err = conn.Exec(ctx, query, pq.Array(newColumns), viewId)
			if err != nil {
				tx.Rollback(ctx)
				return err
			}
		}
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $2
	WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, tableId, data)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `SELECT id, fields FROM "section" WHERE table_id = $1`

	sectionRows, err := tx.Query(ctx, query, tableId)
	if err != nil {
		return err
	}
	defer sectionRows.Close()

	sections := []SectionBody{}

	for sectionRows.Next() {
		var (
			fields     = []byte{}
			id         string
			fieldsBody []map[string]interface{}
		)

		err = sectionRows.Scan(
			&id,
			&fields,
		)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(fields, &fieldsBody); err != nil {
			return err
		}

		sections = append(sections, SectionBody{Id: id, Fields: fieldsBody})
	}

	query = `UPDATE "section" SET fields = $2 WHERE id = $1`

	for _, section := range sections {
		var (
			isExists = false

			newFields []map[string]interface{}
		)

		for i, field := range section.Fields {
			if cast.ToString(field["id"]) != req.Id {
				field["order"] = i + 1
				newFields = append(newFields, field)
			} else {
				isExists = true
			}
		}

		if isExists {

			fieldsBody, err := json.Marshal(newFields)
			if err != nil {
				return err
			}

			_, err = tx.Exec(ctx, query, section.Id, fieldsBody)
			if err != nil {
				return err
			}
		}
	}

	query = `DELETE FROM "field" WHERE id = $1`

	_, err = tx.Exec(ctx, query, req.Id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = fmt.Sprintf(`ALTER TABLE %s DROP COLUMN %s`, tableSlug, fieldSlug)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

type SectionBody struct {
	Id     string
	Fields []map[string]interface{}
}
