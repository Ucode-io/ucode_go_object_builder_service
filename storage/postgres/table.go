package postgres

import (
	"context"
	"encoding/json"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/structpb"

	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
)

type tableRepo struct {
	db *pgxpool.Pool
}

func NewTableRepo(db *pgxpool.Pool) storage.TableRepoI {
	return &tableRepo{
		db: db,
	}
}

func (t *tableRepo) Create(ctx context.Context, req *nb.CreateTableRequest) (resp *nb.CreateTableResponse, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CreateTableResponse{}, err
	}

	jsonAttr, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.CreateTableResponse{}, err
	}

	query := `INSERT INTO "table" (
		id,
		"slug",
		"label",
		"icon",
		"description",
		"show_in_menu",
		"subtitle_field_slug",
		"is_cached",
		"with_increment_id",
		"soft_delete",
		"digit_number",
		"is_changed_by_host",
		"attributes"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	data, err := helper.ChangeHostname([]byte(`{}`))
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	tableId := uuid.NewString()

	_, err = tx.Exec(ctx, query,
		tableId,
		req.Slug,
		req.Label,
		req.Icon,
		req.Description,
		req.ShowInMenu,
		req.SubtitleFieldSlug,
		req.IsCached,
		req.IncrementId.WithIncrementId,
		req.SoftDelete,
		req.IncrementId.DigitNumber,
		data,
		jsonAttr,
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	fieldId := uuid.NewString()

	query = `INSERT INTO "field" (
		"table_id",
		"slug",
		"label",
		"default",
		"type",
		"index",
		id
	) VALUES ($1, 'guid', 'ID', 'uuid_generate_v4()', 'UUID', true, $2)`

	_, err = tx.Exec(ctx, query, tableId, fieldId)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	query = `CREATE TABLE IF NOT EXISTS ` + req.Slug + ` (
		guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP
	)`

	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	query = `INSERT INTO "layout" (
		id, 
		"table_id",
		"order",
		"label",
		"icon",
		"type",
		"is_default",
		"attributes",
		"is_visible_section",
		"is_modal"
	) VALUES ($1, $2, 1, 'Layout', '', 'PopupLayout', true, $3, false, true)`

	_, err = tx.Exec(ctx, query, req.LayoutId, tableId, []byte(`{}`))
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	tabId := uuid.NewString()

	query = `INSERT INTO "tab" (
		"id",
		"order",
		"label",
		"icon",
		"type",
		"layout_id",
		"table_slug"
	) VALUES ($1, 1, 'Tab', '', 'section', $2, $3)`

	_, err = tx.Exec(ctx, query, tabId, req.LayoutId, req.Slug)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	query = `INSERT INTO "section" (
		"id",
		"order",
		"column",
		"label",
		"icon",
		"table_id",
		"tab_id"
	) VALUES ($1, 1, 'SINGLE', 'Info', '', $2, $3)`

	_, err = tx.Exec(ctx, query, uuid.NewString(), tableId, tabId)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	viewID := uuid.NewString()

	query = `INSERT INTO "view" (
		"id",
		"table_slug",
		"type"
	)
	VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, query,
		viewID,
		req.Slug,
		"TABLE",
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}

	roleIds := []string{}

	query = `SELECT guid FROM role`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.CreateTableResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		id := ""

		err = rows.Scan(&id)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.CreateTableResponse{}, err
		}

		roleIds = append(roleIds, id)
	}

	query = `INSERT INTO view_permission (
		guid,
		view_id, 
		role_id, 
		"view", 
		"edit", 
		"delete"
	) VALUES ($1, $2, $3, $4, $5, $6)`

	for _, id := range roleIds {

		_, err = tx.Exec(ctx, query,
			uuid.NewString(),
			viewID,
			id,
			true,
			true,
			true,
		)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.CreateTableResponse{}, err
		}
	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		return &nb.CreateTableResponse{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.CreateTableResponse{}, err
	}

	resp = &nb.CreateTableResponse{
		Id:                tableId,
		Label:             req.Label,
		Slug:              req.Slug,
		ShowInMenu:        req.ShowInMenu,
		Icon:              req.Icon,
		SubtitleFieldSlug: req.SubtitleFieldSlug,
		IsCached:          req.IsCached,
		DefaultEditable:   req.DefaultEditable,
		SoftDelete:        req.SoftDelete,
	}

	resp.Fields = append(resp.Fields, &nb.Field{
		Id:      fieldId,
		TableId: tableId,
		Slug:    "guid",
		Label:   "ID",
		Default: "uuid_generate_v4()",
		Type:    "UUID",
		Index:   "true",
	})

	return resp, nil
}

func (t *tableRepo) GetByID(ctx context.Context, req *nb.TablePrimaryKey) (resp *nb.Table, err error) {

	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	resp = &nb.Table{
		IncrementId: &nb.IncrementID{},
	}

	query := `SELECT 
		id,
		"slug",
		"label",
		"icon",
		"description",
		"show_in_menu",
		"subtitle_field_slug",
		"is_cached",
		"with_increment_id",
		"soft_delete",
		"digit_number",
		"attributes",
		is_login_table
	FROM "table" WHERE id = $1`

	var attrData []byte

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.Slug,
		&resp.Label,
		&resp.Icon,
		&resp.Description,
		&resp.ShowInMenu,
		&resp.SubtitleFieldSlug,
		&resp.IsCached,
		&resp.IncrementId.WithIncrementId,
		&resp.SoftDelete,
		&resp.IncrementId.DigitNumber,
		&attrData,
		&resp.IsLoginTable,
	)
	if err != nil {
		return &nb.Table{}, err
	}

	var attrDataStruct *structpb.Struct
	if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
		return &nb.Table{}, err
	}

	resp.Attributes = attrDataStruct

	return resp, nil
}

func (t *tableRepo) GetAll(ctx context.Context, req *nb.GetAllTablesRequest) (resp *nb.GetAllTablesResponse, err error) {
	conn := psqlpool.Get(req.GetProjectId())

	resp = &nb.GetAllTablesResponse{}

	params := make(map[string]interface{})

	query := `SELECT 
		id,
		"slug",
		"label",
		"icon",
		"description",
		"show_in_menu",
		"subtitle_field_slug",
		"is_changed",
		"with_increment_id",
		"soft_delete",
		"digit_number",
		"attributes",
		is_login_table
	FROM "table" WHERE 1=1`

	if req.Search != "" {
		query += ` label ~* :label `
		params["label"] = req.Search
	}

	if req.IsLoginTable {
		query += ` AND is_login_table = true `
	}

	query += ` ORDER BY created_at DESC `

	if req.Limit != 0 && req.Limit > 0 {
		query += ` LIMIT :limit `
		params["limit"] = req.Limit
	}

	if req.Offset >= 0 {
		query += ` OFFSET :offset `
		params["offset"] = req.Offset
	}

	query, args := helper.ReplaceQueryParams(query, params)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.GetAllTablesResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		table := &nb.Table{
			IncrementId: &nb.IncrementID{},
		}

		var attrData []byte

		err := rows.Scan(
			&table.Id,
			&table.Slug,
			&table.Label,
			&table.Icon,
			&table.Description,
			&table.ShowInMenu,
			&table.SubtitleFieldSlug,
			&table.IsCached,
			&table.IncrementId.WithIncrementId,
			&table.SoftDelete,
			&table.IncrementId.DigitNumber,
			&attrData,
			&table.IsLoginTable,
		)
		if err != nil {
			return &nb.GetAllTablesResponse{}, err
		}

		var attrDataStruct *structpb.Struct
		if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
			return &nb.GetAllTablesResponse{}, err
		}

		table.Attributes = attrDataStruct

		resp.Tables = append(resp.Tables, table)
	}

	query = `SELECT COUNT(*) FROM "table" `

	err = conn.QueryRow(ctx, query).Scan(&resp.Count)
	if err != nil {
		return &nb.GetAllTablesResponse{}, err
	}

	return resp, nil
}

func (t *tableRepo) Update(ctx context.Context, req *nb.UpdateTableRequest) (resp *nb.Table, err error) {

	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Table{}, err
	}

	// Think about it...
	// table, err := t.GetByID(ctx, &nb.TablePrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	// if err != nil {
	// 	return &nb.Table{}, err
	// }

	query := `UPDATE "table" SET 
		"label" = $2,
		"icon" = $3,
		"description" = $4,
		"show_in_menu" = $5,
		"subtitle_field_slug" = $6,
		"is_cached" = $7,
		"with_increment_id" = $8,
		"soft_delete" = $9,
		"digit_number" = $10,
		"attributes" = $11,
		is_login_table = $12
	WHERE id = $1`

	_, err = tx.Exec(ctx, query, req.Id,
		req.Label,
		req.Icon,
		req.Description,
		req.ShowInMenu,
		req.SubtitleFieldSlug,
		req.IsCached,
		req.IncrementId.WithIncrementId,
		req.SoftDelete,
		req.IncrementId.DigitNumber,
		req.Attributes,
		req.IsLoginTable,
	)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Table{}, err
	}

	query = `SELECT guid FROM "role" `

	rows, err := tx.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Table{}, err
	}
	defer rows.Close()

	query = `SELECT COUNT(*) FROM "record_permission" WHERE table_slug = $1 AND role_id = $2`
	createQuery := `INSERT INTO "record_permission" (
		table_slug,
		role_id,
		read,
		update,
		write,
		delete,
		is_have_condition
	) VALUES ($1, $2, 'Yes', 'Yes', 'Yes', 'Yes', false)`

	guids := []string{}

	for rows.Next() {

		var (
			guid = ""
		)

		err = rows.Scan(&guid)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Table{}, err
		}

		guids = append(guids, guid)
	}

	for _, guid := range guids {

		count := 0

		err = tx.QueryRow(ctx, query, req.Slug, guid).Scan(&count)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Table{}, err
		}

		if count == 0 {
			_, err = tx.Exec(ctx, createQuery, req.Slug, guid)
			if err != nil {
				tx.Rollback(ctx)
				return &nb.Table{}, err
			}
		}
	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		return &nb.Table{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.Table{}, err
	}

	return t.GetByID(ctx, &nb.TablePrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (t *tableRepo) Delete(ctx context.Context, req *nb.TablePrimaryKey) error {

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	slug := ""

	query := `DELETE FROM "table" WHERE id = $1 RETURNING slug`

	err = tx.QueryRow(ctx, query, req.Id).Scan(&slug)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `DROP TABLE IF EXISTS ` + slug

	_, err = tx.Exec(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	query = `DISCARD PLANS;`

	_, err = conn.Exec(ctx, query)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
