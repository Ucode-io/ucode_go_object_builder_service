package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/pkg/errors"
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
	var (
		conn                                  = psqlpool.Get(req.GetProjectId())
		body, data                            []byte
		fields                                = []SectionFields{}
		tableSlug, layoutId, tabId, sectionId string
		sectionCount                          int32
		ids                                   []string
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error creating transaction")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	req.Slug = strings.ToLower(req.GetSlug())

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
		"is_search"
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
	)`

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error marshaling attributes")
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
		true,
	)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error inserting field")
	}

	query = `SELECT is_changed_by_host, slug FROM "table" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error getting table")
	}

	query = `ALTER TABLE "` + tableSlug + `" ADD COLUMN ` + req.Slug + " " + helper.GetDataType(req.Type)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error adding column")
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error changing hostname")
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $1
	WHERE id = $2
	`

	_, err = tx.Exec(ctx, query, data, req.TableId)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error updating table")
	}

	query = `SELECT guid FROM "role"`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error getting roles")
	}
	defer rows.Close()

	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error scanning role")
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
			return &nb.Field{}, errors.Wrap(err, "error inserting field permission")
		}
	}

	query = `SELECT id FROM "layout" WHERE table_id = $1`
	err = tx.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, errors.Wrap(err, "error getting layout")
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	query = `SELECT id FROM "tab" WHERE "layout_id" = $1 and type = 'section'`
	err = tx.QueryRow(ctx, query, layoutId).Scan(&tabId)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, errors.Wrap(err, "error getting tab")
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	query = `SELECT id, fields FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
	err = tx.QueryRow(ctx, query, tabId).Scan(&sectionId, &body)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error getting section")
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
	err = tx.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
	if err != nil && err != pgx.ErrNoRows {
		return &nb.Field{}, errors.Wrap(err, "error getting section count")
	} else if err == pgx.ErrNoRows {
		return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	}

	if err := json.Unmarshal(body, &fields); err != nil {
		return &nb.Field{}, errors.Wrap(err, "error unmarshaling section")
	}

	if len(fields) < 3 {
		query := `UPDATE "section" SET fields = $2 WHERE id = $1`

		fields = append(fields, SectionFields{
			Id:    req.Id,
			Order: len(fields) + 1,
		})

		reqBody, err := json.Marshal(fields)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error marshaling fields")
		}

		_, err = tx.Exec(ctx, query, sectionId, reqBody)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error updating section")
		}
	} else {
		query = `INSERT INTO "section" ("order", "column", label, table_id, tab_id, fields) VALUES ($1, $2, $3, $4, $5, $6)`

		sectionId = uuid.NewString()

		fields := []SectionFields{{Id: req.Id, Order: 1}}

		reqBody, err := json.Marshal(fields)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error marshaling fields")
		}

		_, err = tx.Exec(ctx, query, sectionCount+1, "SINGLE", "Info", req.TableId, tabId, reqBody)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error inserting section")
		}
	}

	if req.Type == "INCREMENT_ID" {
		query = `INSERT INTO "incrementseqs" (field_slug, table_slug) VALUES ($1, $2)`

		_, err = tx.Exec(ctx, query, req.Slug, tableSlug)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error inserting incrementseq")
		}

	}

	err = tx.Commit(ctx)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error committing transaction")
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

// DONE
func (f *fieldRepo) GetByID(ctx context.Context, req *nb.FieldPrimaryKey) (resp *nb.Field, err error) {
	var (
		conn           = psqlpool.Get(req.GetProjectId())
		attributes     = []byte{}
		relationIdNull sql.NullString
	)

	resp = &nb.Field{}
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
		return &nb.Field{}, errors.Wrap(err, "error getting field")
	}

	resp.RelationId = relationIdNull.String

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return &nb.Field{}, errors.Wrap(err, "error unmarshaling attributes")
	}

	return resp, nil
}

func (f *fieldRepo) GetAll(ctx context.Context, req *nb.GetAllFieldsRequest) (resp *nb.GetAllFieldsResponse, err error) {
	conn := psqlpool.Get(req.GetProjectId())
	resp = &nb.GetAllFieldsResponse{}

	getTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Id: req.TableId, Slug: req.TableSlug})
	if err != nil {
		return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error getting table")
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
		return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error getting fields")
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
			return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error scanning fields")
		}

		if attributes.Valid {
			err := json.Unmarshal([]byte(attributes.String), &field.Attributes)
			if err != nil {
				return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error unmarshaling attributes")
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
		return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error getting count")
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
			return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error getting relation")
		}
		defer rows.Close()

		for rows.Next() {
			var tableFrom string

			err = rows.Scan(&tableFrom)
			if err != nil {
				return &nb.GetAllFieldsResponse{}, errors.Wrap(err, "error scanning table from")
			}
		}
	}

	return resp, nil
}

func (f *fieldRepo) GetAllForItems(ctx context.Context, req *nb.GetAllFieldsForItemsRequest) (resp *nb.AllFields, err error) {
	// Skipped ...

	return &nb.AllFields{}, nil
}

func (f *fieldRepo) Update(ctx context.Context, req *nb.Field) (resp *nb.Field, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error creating transaction")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	resp, err = f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error getting field")
	}

	if resp.IsSystem {
		return &nb.Field{}, fmt.Errorf("error you can't update this field its system field")
	}

	tableSlug := ""

	query := `SELECT slug FROM "table" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.TableId).Scan(&tableSlug)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error getting table slug")
	}

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.Field{}, errors.Wrap(err, "error marshaling attributes")
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
	WHERE id = $1`

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
		return &nb.Field{}, errors.Wrap(err, "error updating field")
	}

	if resp.Type != req.Type {
		query = fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, tableSlug, resp.Slug)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error dropping column")
		}

		fieldType := helper.GetDataType(req.Type)

		query = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN %s %s`, tableSlug, req.Slug, fieldType)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error adding column")
		}
	}

	if resp.Slug != req.Slug {
		query = fmt.Sprintf(`ALTER TABLE "%s"
		RENAME COLUMN %s TO %s;`, tableSlug, resp.Slug, req.Slug)

		_, err = tx.Exec(ctx, query)
		if err != nil {
			return &nb.Field{}, errors.Wrap(err, "error renaming column")
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return &nb.Field{}, errors.Wrap(err, "error committing transaction")
	}

	return f.GetByID(ctx, &nb.FieldPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (f *fieldRepo) UpdateSearch(ctx context.Context, req *nb.SearchUpdateRequest) error {
	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating transaction")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `UPDATE "field" SET is_search = $1 WHERE id = $2`

	for _, val := range req.Fields {
		_, err = tx.Exec(ctx, query, val.IsSearch, val.Id)
		if err != nil {
			return errors.Wrap(err, "error updating search")
		}
	}

	query = `SELECT is_changed_by_host FROM "table" where slug = $1`

	var (
		data = []byte{}
	)

	err = tx.QueryRow(ctx, query, req.TableSlug).Scan(&data)
	if err != nil {
		return errors.Wrap(err, "error getting table")
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return errors.Wrap(err, "error changing hostname")
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $2
	WHERE slug = $1
	`

	_, err = tx.Exec(ctx, query, req.TableSlug, data)
	if err != nil {
		return errors.Wrap(err, "error updating table")
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "error committing transaction")
	}

	return nil
}

func (f *fieldRepo) Delete(ctx context.Context, req *nb.FieldPrimaryKey) error {
	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "error creating transaction")
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var (
		isSystem, isExists   bool
		tableId, viewId      string
		columns, newColumns  []string
		tableSlug, fieldSlug string
		data                 = []byte{}
	)

	query := `SELECT is_system, table_id, slug FROM "field" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.Id).Scan(&isSystem, &tableId, &fieldSlug)
	if err != nil {
		return errors.Wrap(err, "error getting field")
	}

	if isSystem {
		return errors.Wrap(err, "error you can't delete this field its system field")
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	err = tx.QueryRow(ctx, query, tableId).Scan(&data, &tableSlug)
	if err != nil {
		return errors.Wrap(err, "error getting table")
	}

	query = `DELETE FROM "field_permission" WHERE field_id = $1`

	_, err = tx.Exec(ctx, query, req.Id)
	if err != nil && err != pgx.ErrNoRows {
		return errors.Wrap(err, "error deleting field permission")
	}

	query = `SELECT id, columns FROM "view" WHERE table_slug = $1 `

	err = tx.QueryRow(ctx, query, tableSlug).Scan(&viewId, &columns)
	if err != nil {
		return errors.Wrap(err, "error getting view")
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

			_, err = tx.Exec(ctx, query, pq.Array(newColumns), viewId)
			if err != nil {
				return errors.Wrap(err, "error updating view")
			}
		}
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return errors.Wrap(err, "error changing hostname")
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $2
	WHERE id = $1
	`

	_, err = tx.Exec(ctx, query, tableId, data)
	if err != nil {
		return errors.Wrap(err, "error updating table")
	}

	query = `SELECT id, fields FROM "section" WHERE table_id = $1`

	sectionRows, err := tx.Query(ctx, query, tableId)
	if err != nil {
		return errors.Wrap(err, "error getting section")
	}
	defer sectionRows.Close()

	sections := []SectionBody{}

	for sectionRows.Next() {
		var (
			id         string
			fields     = []byte{}
			fieldsBody []map[string]interface{}
		)

		err = sectionRows.Scan(&id, &fields)
		if err != nil {
			return errors.Wrap(err, "error scanning section")
		}

		if err := json.Unmarshal(fields, &fieldsBody); err != nil {
			return errors.Wrap(err, "error unmarshaling fields")
		}

		sections = append(sections, SectionBody{Id: id, Fields: fieldsBody})
	}

	query = `UPDATE "section" SET fields = $2 WHERE id = $1`

	for _, section := range sections {
		var (
			isExists  = false
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
				return errors.Wrap(err, "error marshaling fields")
			}

			_, err = tx.Exec(ctx, query, section.Id, fieldsBody)
			if err != nil {
				return errors.Wrap(err, "error updating section")
			}
		}
	}

	query = `DELETE FROM "field" WHERE id = $1`

	_, err = tx.Exec(ctx, query, req.Id)
	if err != nil {
		return errors.Wrap(err, "error deleting field")
	}

	query = `
		UPDATE "relation"
			SET view_fields = array_remove(view_fields, $1)
		WHERE $1 = ANY(view_fields);`
	_, err = tx.Exec(ctx, query, req.Id)
	if err != nil {
		return errors.Wrap(err, "error deleting relation view fields")
	}

	query = fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, tableSlug, fieldSlug)

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return errors.Wrap(err, "error dropping column")
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "error committing transaction")
	}

	return nil
}

func (f *fieldRepo) FieldsWithPermissions(ctx context.Context, req *nb.FieldsWithRelationRequest) (resp *nb.FieldsWithRelationsResponse, err error) {
	resp = &nb.FieldsWithRelationsResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	getTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Slug: req.TableSlug})
	if err != nil {
		return &nb.FieldsWithRelationsResponse{}, err
	}

	query := `SELECT slug, label FROM "field" WHERE table_id = $1 AND type <> 'LOOKUP'`
	rows, err := conn.Query(ctx, query, cast.ToString(getTable["id"]))
	if err != nil {
		return &nb.FieldsWithRelationsResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			slug  string
			label string
		)
		err = rows.Scan(
			&slug,
			&label,
		)
		if err != nil {
			return &nb.FieldsWithRelationsResponse{}, err
		}

		resp.Fields = append(resp.Fields, &nb.FieldNew{Slug: slug, Label: label})
	}

	query = `SELECT table_to, table_from, field_from FROM "relation" WHERE (table_to = $1 OR table_from = $1) AND type = 'Many2One'`
	rows, err = conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.FieldsWithRelationsResponse{}, err
	}
	for rows.Next() {
		var (
			tableTo   string
			tableFrom string
			fieldFrom string
			tableSlug string
		)
		err = rows.Scan(
			&tableTo,
			&tableFrom,
			&fieldFrom,
		)
		if err != nil {
			return &nb.FieldsWithRelationsResponse{}, err
		}

		relation := &nb.RelationNew{}
		if tableFrom == req.TableSlug {
			relation = &nb.RelationNew{
				Slug:  fieldFrom + "_data",
				Label: tableTo,
			}
			tableSlug = tableTo
		} else {
			relation = &nb.RelationNew{
				Slug:  tableFrom,
				Label: tableFrom,
			}
			tableSlug = tableFrom
		}

		getTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Slug: tableSlug})
		if err != nil {
			return &nb.FieldsWithRelationsResponse{}, err
		}
		query := `SELECT slug, label FROM "field" WHERE table_id = $1 AND type <> 'LOOKUP'`
		rows, err := conn.Query(ctx, query, cast.ToString(getTable["id"]))
		if err != nil {
			return &nb.FieldsWithRelationsResponse{}, err
		}
		defer rows.Close()

		for rows.Next() {
			var (
				slug  string
				label string
			)
			err = rows.Scan(
				&slug,
				&label,
			)
			if err != nil {
				return &nb.FieldsWithRelationsResponse{}, err
			}

			relation.Fields = append(relation.Fields, &nb.FieldNew{Slug: slug, Label: label})
		}

		query = `SELECT table_to, table_from, field_from FROM "relation" WHERE (table_to = $1 OR table_from = $1) AND type = 'Many2One'`
		rows, err = conn.Query(ctx, query, tableSlug)
		if err != nil {
			return &nb.FieldsWithRelationsResponse{}, err
		}
		for rows.Next() {
			var (
				tableTo    string
				tableFrom  string
				fieldFrom  string
				tableSlug2 string
			)
			err = rows.Scan(
				&tableTo,
				&tableFrom,
				&fieldFrom,
			)
			if err != nil {
				return &nb.FieldsWithRelationsResponse{}, err
			}
			relation2 := &nb.RelationNew{}
			if tableFrom == tableSlug {
				relation2 = &nb.RelationNew{
					Slug:  fieldFrom + "_data",
					Label: tableTo,
				}
				tableSlug2 = tableTo
			} else {
				relation2 = &nb.RelationNew{
					Slug:  tableFrom,
					Label: tableFrom,
				}
				tableSlug2 = tableFrom
			}

			getTable, err := helper.GetTableByIdSlug(ctx, helper.GetTableByIdSlugReq{Conn: conn, Slug: tableSlug2})
			if err != nil {
				return &nb.FieldsWithRelationsResponse{}, err
			}
			query := `SELECT slug, label FROM "field" WHERE table_id = $1 AND type <> 'LOOKUP'`
			rows, err := conn.Query(ctx, query, cast.ToString(getTable["id"]))
			if err != nil {
				return &nb.FieldsWithRelationsResponse{}, err
			}
			defer rows.Close()

			for rows.Next() {
				var (
					slug  string
					label string
				)
				err = rows.Scan(
					&slug,
					&label,
				)
				if err != nil {
					return &nb.FieldsWithRelationsResponse{}, err
				}

				relation2.Fields = append(relation2.Fields, &nb.FieldNew{Slug: slug, Label: label})
			}
			relation.Relations = append(relation.Relations, relation2)
		}

		resp.Relations = append(resp.Relations, relation)
	}

	return resp, nil
}

type SectionBody struct {
	Id     string
	Fields []map[string]interface{}
}
