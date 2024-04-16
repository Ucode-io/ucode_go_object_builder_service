package postgres

import (
	"context"
	"encoding/json"
	"fmt"
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

	fieldId := uuid.NewString()

	tx, err := f.db.Begin(ctx)
	if err != nil {
		return &nb.Field{}, err
	}

	// ! FIELD_TYPE AUTOFILL DOESN'T USE IN NEW VERSION

	// if req.Type == "AUTOFILL" {
	// 	autoFillTableSlug := req.AutofillTable

	// 	if strings.Contains(req.AutofillTable, "#") {
	// 		autoFillTableSlug = strings.Split(req.AutofillTable, "#")[0]
	// 	}

	// 	autoFill, err := helper.GetTableByIdSlug(ctx, conn, "", autoFillTableSlug)
	// 	if err != nil {
	// 		return &nb.Field{}, err
	// 	}

	// 	var autoFillFieldSlug string

	// 	if strings.Contains(req.AutofillField, ".") {
	// 		splitedTable := strings.Split(strings.Split(req.AutofillField, ".")[0], "_")
	// 		tableSlug := ""
	// 		for i := 0; i < len(splitedTable)-2; i++ {
	// 			tableSlug = tableSlug + "_" + splitedTable[i]
	// 		}
	// 		tableSlug = tableSlug[1:]
	// 		autoFill, err = helper.GetTableByIdSlug(ctx, conn, "", tableSlug)
	// 		if err != nil {
	// 			return &nb.Field{}, err
	// 		}

	// 		autoFillFieldSlug = strings.Split(req.AutofillField, ".")[1]
	// 	} else {
	// 		autoFillFieldSlug = req.AutofillField
	// 	}

	// 	autoFillField, err := helper.GetFieldBySlug(ctx, conn, autoFillFieldSlug, cast.ToString(autoFill["id"]))
	// 	if err != nil && err != pgx.ErrNoRows {
	// 		return &nb.Field{}, err
	// 	}

	// 	attributes, _ := autoFillField["attributes"].([]byte)

	// 	if err := json.Unmarshal(attributes, &req.Attributes); err != nil {
	// 		return &nb.Field{}, err
	// 	}
	// 	req.Type = cast.ToString(autoFillField["type"])
	// }

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
		"automatic",
		relation_id
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	)`

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	_, err = tx.Exec(ctx, query,
		fieldId,
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
		req.GetRelationId(),
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Field{}, err
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	var (
		data      = []byte{}
		tableSlug string
		// layoutId      string
		// tabId         string
		// sectionId     string
		// sectionCount  int32
		// sectionFields int32
	)

	err = f.db.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
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

	row, err := f.db.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
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
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}

		_, err = tx.Exec(ctx, query, tableSlug, fieldId, req.Label, id)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Field{}, err
		}
	}

	// ?? WHEN WE ADD NEED TO CHECK FIELD WITH LAYOUT

	// query = `SELECT id FROM "layout" WHERE table_id = $1`
	// err = f.db.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
	// if err != nil && err != pgx.ErrNoRows {
	// 	tx.Rollback(ctx)
	// 	return &nb.Field{}, err
	// }

	// query = `SELECT id FROM "tab" WHERE "layout_id" = $1 and type = 'section'`
	// err = f.db.QueryRow(ctx, query, layoutId).Scan(&tabId)
	// if err != nil && err != pgx.ErrNoRows {
	// 	fmt.Println("HELLO OKOOOOKOKOOKKOK")
	// 	tx.Rollback(ctx)
	// 	return &nb.Field{}, err
	// }

	// query = `SELECT id FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
	// err = f.db.QueryRow(ctx, query, tabId).Scan(&sectionId)
	// if err != nil {
	// 	tx.Rollback(ctx)
	// 	return &nb.Field{}, err
	// }

	// queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
	// err = f.db.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
	// if err != nil && err != pgx.ErrNoRows {
	// 	tx.Rollback(ctx)
	// 	return &nb.Field{}, err
	// }

	// query = `SELECT COUNT(*) FROM "section_fields" WHERE section_id = $1`
	// err = f.db.QueryRow(ctx, query, tabId).Scan(&sectionFields)
	// if err != nil && err != pgx.ErrNoRows {
	// 	tx.Rollback(ctx)
	// 	return &nb.Field{}, err
	// }

	// if sectionFields < 3 {
	// 	query := `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

	// 	_, err = tx.Exec(ctx, query, fieldId, sectionFields+1, req.Label, sectionId)
	// 	if err != nil {
	// 		tx.Rollback(ctx)
	// 		return &nb.Field{}, err
	// 	}
	// } else {
	// 	query = `INSERT INTO "section" (id, order, column, label, table_id, tab_id) VALUES ($1, $2, $3, $4, $5, $6)`

	// 	sectionId = uuid.NewString()

	// 	_, err = tx.Exec(ctx, query, sectionId, sectionCount+1, "SINGLE", "Info", req.TableId, tabId)
	// 	if err != nil {
	// 		tx.Rollback(ctx)
	// 		return &nb.Field{}, err
	// 	}

	// 	query = `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

	// 	_, err = tx.Exec(ctx, query, fieldId, 1, req.Label, sectionId)
	// 	if err != nil {
	// 		tx.Rollback(ctx)
	// 		return &nb.Field{}, err
	// 	}
	// }

	if err := tx.Commit(ctx); err != nil {
		return &nb.Field{}, err
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

// DONE
func (f *fieldRepo) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error) {

	// conn := psqlpool.Get(req.ProjectId)

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
		"unique",
		"automatic",
		relation_id
	FROM "field" WHERE id = $1`

	err = f.db.QueryRow(ctx, query, req.Id).Scan(
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

	err = conn.QueryRow(ctx, query, req.TableId).Scan(&tableSlug)
	if err != nil {
		return &nb.Field{}, err
	}

	attributes := json.Marshaler(req.Attributes)

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
		"automatic" = $13,
		relation_id = $14
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
		req.Unique,
		req.Automatic,
		req.RelationId,
	)
	if err != nil {
		return &nb.Field{}, err
	}

	if resp.Type != req.Type {

		// * QUERY -> Try to change columns type if error it changes value to new type's default value
		// * OLD: VARCHAR - "John Doe" NEW: FLOAT - 0.0

		fieldType := helper.GetDataType(req.Type)
		regExp := helper.GetRegExp(fieldType)
		defaultValue := helper.GetDefault(fieldType)

		query = `ALTER TABLE $1
		ALTER COLUMN $2 TYPE $3
		USING CASE WHEN $2 ~ $4 THEN $2::$3 ELSE $5 END;`

		_, err = conn.Exec(ctx, query, tableSlug, req.Slug, fieldType, regExp, defaultValue)
		if err != nil {
			return &nb.Field{}, err
		}
	}

	if resp.Slug != req.Slug {
		query = `ALTER TABLE $1
		RENAME COLUMN $2 TO $3;`

		_, err = conn.Exec(ctx, query, tableSlug, resp.Slug, req.Slug)
		if err != nil {
			return &nb.Field{}, err
		}
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (f *fieldRepo) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error {
	conn := psqlpool.Get(req.ProjectId)

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

	conn := psqlpool.Get(req.ProjectId)

	tx, err := f.db.Begin(ctx)
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

	err = tx.QueryRow(ctx, query, tableSlug).Scan(&viewId, pq.Array(&columns))
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

	query = `DELETE FROM "field" WHERE id = $1`

	_, err = tx.Exec(ctx, query, req.Id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `ALTER TABLE $1 DROP COLUMN $2 `

	_, err = tx.Exec(ctx, query, tableSlug, fieldSlug)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
