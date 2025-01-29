package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/structpb"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"
)

type tableRepo struct {
	db *psqlpool.Pool
}

func NewTableRepo(db *psqlpool.Pool) storage.TableRepoI {
	return &tableRepo{
		db: db,
	}
}

func (t *tableRepo) Create(ctx context.Context, req *nb.CreateTableRequest) (resp *nb.CreateTableResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.Create")
	defer dbSpan.Finish()

	var conn = psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	jsonAttr, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to marshal attributes")
	}

	query := `INSERT INTO "table" (
		id, "slug", "label", "icon",
		"description", "show_in_menu",
		"subtitle_field_slug", "is_cached",
		"with_increment_id", "soft_delete",
		"digit_number", "is_changed_by_host", "attributes", is_login_table
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`

	data, err := helper.ChangeHostname([]byte(`{}`))
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to change hostname")
	}

	var tableId = uuid.NewString()

	_, err = tx.Exec(ctx, query,
		tableId, req.Slug, req.Label, req.Icon, req.Description,
		req.ShowInMenu, req.SubtitleFieldSlug, req.IsCached,
		req.GetIncrementId().GetWithIncrementId(), req.SoftDelete,
		req.GetIncrementId().GetDigitNumber(), data, jsonAttr, req.IsLoginTable,
	)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert table")
	}

	var (
		fieldId       = uuid.NewString()
		folderGroupId = uuid.NewString()
	)

	query = `INSERT INTO "field" ( "table_id", "slug", "label", "default", "type", "index", id) 
			 VALUES ($1, 'guid', 'ID', 'uuid_generate_v4()', 'UUID', true, $2), ($1, 'folder_id', 'Folder Id', NULL, 'UUID', NULL, $3)`

	_, err = tx.Exec(ctx, query, tableId, fieldId, folderGroupId)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert field")
	}

	query = `CREATE TABLE IF NOT EXISTS "` + req.Slug + `" (
		guid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		folder_id UUID REFERENCES "folder_group"("id") ON DELETE SET NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMP
	)`

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to create table")
	}

	query = `INSERT INTO "layout" (
		id, "table_id", "order", "label", "icon", "type", "is_default", "attributes", "is_visible_section", "is_modal" ) 
		VALUES ($1, $2, 1, 'Layout', '', 'PopupLayout', true, $3, false, true)`

	_, err = tx.Exec(ctx, query, req.LayoutId, tableId, []byte(`{}`))
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert layout")
	}

	var tabId = uuid.NewString()

	query = `INSERT INTO "tab" ("id", "order", "label", "icon", "type", "layout_id", "table_slug") 
			 VALUES ($1, 1, 'Tab', '', 'section', $2, $3)`

	_, err = tx.Exec(ctx, query, tabId, req.LayoutId, req.Slug)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert tab")
	}

	query = `INSERT INTO "section" ("id", "order", "column", "label", "icon", "table_id", "tab_id") 
			 VALUES ($1, 1, 'SINGLE', 'Info', '', $2, $3)`

	_, err = tx.Exec(ctx, query, uuid.NewString(), tableId, tabId)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert section")
	}

	var viewID = uuid.NewString()

	query = `INSERT INTO "view" ("id", "table_slug", "type" )
			 VALUES ($1, $2, $3)`

	_, err = tx.Exec(ctx, query, viewID, req.Slug, "TABLE")
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert view")
	}

	var roleIds = []string{}
	query = `SELECT guid FROM role`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to select role")
	}
	defer rows.Close()

	for rows.Next() {
		var id string

		err = rows.Scan(&id)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to scan role")
		}

		roleIds = append(roleIds, id)
	}

	query = `INSERT INTO view_permission (guid, view_id, role_id, "view", "edit", "delete") 
			VALUES ($1, $2, $3, $4, $5, $6)`

	recordPermission := `INSERT INTO record_permission (
		guid, role_id, table_slug, is_have_condition, delete, write, update,
		read, pdf_action, add_field, language_btn, view_create, automation,
		settings, share_modal, add_filter, field_filter, fix_column, tab_group,
		columns, "group", excel_menu, search_button) 
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`

	for _, id := range roleIds {
		_, err = tx.Exec(ctx, query,
			uuid.NewString(), viewID, id, true, true, true,
		)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert view permission")
		}

		_, err = tx.Exec(ctx, recordPermission,
			uuid.NewString(), id, req.Slug, true, "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes",
			"Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes", "Yes",
		)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert record permission")
		}
	}

	if req.IsLoginTable {
		attributesMap, err := helper.ConvertStructToMap(req.Attributes)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "convert attributes struct to map")
		}

		attributesAuthInfo, ok := attributesMap["auth_info"].(map[string]any)
		if !ok {
			return &nb.CreateTableResponse{}, errors.New("auth_info does not exist")
		}

		loginStrategy, ok := attributesAuthInfo["login_strategy"].([]any)
		if !ok {
			return &nb.CreateTableResponse{}, errors.New("login_strategy does not exist")
		}

		var authInfo = map[string]any{
			"client_type_id": "client_type_id",
			"login_strategy": loginStrategy,
			"role_id":        "role_id",
		}

		for _, value := range loginStrategy {
			strategy := cast.ToString(value)
			switch cast.ToString(strategy) {
			case "phone":
				phoneAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Phone", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert to struct phone field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Phone",
					Slug:       "phone",
					Type:       "INTERNATION_PHONE",
					TableId:    tableId,
					TableSlug:  req.Slug,
					ShowLabel:  true,
					Required:   false,
					Attributes: phoneAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when upsert phone field")
				}
				authInfo["phone"] = "phone"
			case "login":
				loginAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Login", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert to struct login field attributes")
				}

				passwordAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Password", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert to struct password field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Login",
					Slug:       "login",
					Type:       "SINGLE_LINE",
					TableId:    tableId,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: loginAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when upsert login field")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Password",
					Slug:       "password",
					Type:       "PASSWORD",
					TableId:    tableId,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: passwordAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when upsert password field")
				}

				authInfo["login"] = "login"
				authInfo["password"] = "password"
			case "email":
				emailAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Email", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert to struct email field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Email",
					Slug:       "email",
					Type:       "EMAIL",
					TableId:    tableId,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: emailAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.CreateTableResponse{}, errors.Wrap(err, "when upsert email field")
				}

				authInfo["email"] = "email"
			default:
				return &nb.CreateTableResponse{}, errors.New("Unknown strategy: " + cast.ToString(strategy))
			}
		}

		var clientTypeRelationCount, roleRelationCount int32
		query = `SELECT COUNT(id) 
			FROM "relation" 
			WHERE table_from = $1 AND field_from = 'client_type_id' AND table_to = 'client_type' AND field_to = 'id'`

		err = tx.QueryRow(ctx, query, req.Slug).Scan(&clientTypeRelationCount)
		if err != nil && err != pgx.ErrNoRows {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "when get count client type relaion")
		}

		if clientTypeRelationCount == 0 {
			clientTypeAttributes, err := helper.ConvertMapToStruct(map[string]any{
				"label_en":              "Client Type",
				"label_to_en":           req.Label,
				"table_editable":        false,
				"enable_multi_language": false,
			})
			if err != nil {
				return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert ")
			}

			_, err = helper.CreateRelationWithTx(ctx, &models.CreateRelationRequest{
				Tx:                tx,
				Id:                uuid.NewString(),
				TableFrom:         req.Slug,
				TableTo:           config.CLIENT_TYPE,
				Type:              config.MANY2ONE,
				ViewFields:        []string{"04d0889a-b9ba-4f5c-8473-c8447aab350d"},
				RelationTableSlug: config.CLIENT_TYPE,
				Attributes:        clientTypeAttributes,
				AutoFilters:       []*nb.AutoFilter{{FieldTo: "", FieldFrom: ""}},
				RelationFieldId:   uuid.NewString(),
			})
			if err != nil {
				return &nb.CreateTableResponse{}, errors.Wrap(err, "when create relation")
			}
		}

		query = `SELECT COUNT(id) 
				FROM "relation" 
				WHERE table_from = $1 AND field_from = 'role_id' AND table_to = 'role' AND field_to = 'id'`

		err = tx.QueryRow(ctx, query, req.Slug).Scan(&roleRelationCount)
		if err != nil && err != pgx.ErrNoRows {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "when get count client type relaion")
		}

		if roleRelationCount == 0 {
			roleAttributes, err := helper.ConvertMapToStruct(map[string]any{
				"label_en":              "Role",
				"label_to_en":           req.Label,
				"table_editable":        false,
				"enable_multi_language": false,
			})
			if err != nil {
				return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert role attributes")
			}

			_, err = helper.CreateRelationWithTx(ctx, &models.CreateRelationRequest{
				Tx:                tx,
				Id:                uuid.NewString(),
				TableFrom:         req.Slug,
				TableTo:           config.ROLE,
				Type:              config.MANY2ONE,
				ViewFields:        []string{"c12adfef-2991-4c6a-9dff-b4ab8810f0df"},
				RelationTableSlug: config.ROLE,
				Attributes:        roleAttributes,
				AutoFilters:       []*nb.AutoFilter{{FieldTo: "client_type_id", FieldFrom: "client_type_id"}},
				RelationFieldId:   uuid.NewString(),
			})
			if err != nil {
				return &nb.CreateTableResponse{}, errors.Wrap(err, "when create relation")
			}
		}

		req.Attributes, err = helper.ConvertMapToStruct(map[string]any{
			"auth_info": authInfo,
			"label":     req.Label,
			"label_en":  req.Label,
		})
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "when convert to struct auth_info")
		}

		query = `UPDATE "table" SET attributes = $2 WHERE id = $1`

		_, err = tx.Exec(ctx, query, tableId, req.Attributes)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "when update table attributes")
		}

		query = `INSERT INTO "field" (id, table_id, slug, label, type, is_visible, is_system, attributes) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    		ON CONFLICT (table_id, slug)
    		DO NOTHING`

		_, err = tx.Exec(ctx, query,
			uuid.NewString(), tableId, "user_id_auth",
			"User ID Auth", "UUID", false, true,
			`{"label_en":"UserIdAuth","label":"UserIdAuth","defaultValue":""}`,
		)
		if err != nil {
			return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to insert field")
		}

		query = `ALTER TABLE IF EXISTS "` + req.GetSlug() + `" ADD COLUMN IF NOT EXISTS ` + ` user_id_auth` + ` UUID`
		_, _ = tx.Exec(ctx, query)
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.CreateTableResponse{}, errors.Wrap(err, "failed to commit transaction")
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

	resp.Fields = append(resp.Fields,
		&nb.Field{
			Id:      fieldId,
			TableId: tableId,
			Slug:    "guid",
			Label:   "ID",
			Default: "uuid_generate_v4()",
			Type:    "UUID",
			Index:   "true",
		}, &nb.Field{
			Id:      folderGroupId,
			TableId: tableId,
			Slug:    "folder_id",
			Label:   "Folder Id",
			Type:    "UUID",
		},
	)

	return resp, nil
}

func (t *tableRepo) GetByID(ctx context.Context, req *nb.TablePrimaryKey) (*nb.Table, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.GetByID")
	defer dbSpan.Finish()

	var (
		conn          = psqlpool.Get(req.GetProjectId())
		filter string = "id = $1"
		resp          = &nb.Table{IncrementId: &nb.IncrementID{}}
	)

	_, err := uuid.Parse(req.Id)
	if err != nil {
		filter = "slug = $1"
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
		"order_by",
		"digit_number",
		"attributes",
		is_login_table
	FROM "table" WHERE ` + filter

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
		&resp.OrderBy,
		&resp.IncrementId.DigitNumber,
		&attrData,
		&resp.IsLoginTable,
	)

	if err == pgx.ErrNoRows {
		return resp, nil
	} else if err != nil {
		return resp, errors.Wrap(err, "table get by id scan")
	}

	resp.Exists = true
	var attrDataStruct *structpb.Struct
	if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
		return &nb.Table{}, errors.Wrap(err, "unmarchal structpb table get by id")
	}

	resp.Attributes = attrDataStruct

	return resp, nil
}

func (t *tableRepo) GetAll(ctx context.Context, req *nb.GetAllTablesRequest) (resp *nb.GetAllTablesResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.GetAll")
	defer dbSpan.Finish()

	var (
		conn   = psqlpool.Get(req.GetProjectId())
		params = make(map[string]any)
		query  = `SELECT 
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
			"order_by",
			"digit_number",
			"attributes",
			is_login_table
		FROM "table" WHERE (is_system = false OR (slug = 'role' OR slug = 'client_type' OR slug = 'person')) `
	)

	resp = &nb.GetAllTablesResponse{}

	if req.Search != "" {
		query += ` AND label ~* :label `
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
			&table.OrderBy,
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.Update")
	defer dbSpan.Finish()

	var (
		conn          = psqlpool.Get(req.GetProjectId())
		oldAttributes []byte
		isLoginTable  sql.NullBool
	)

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Table{}, errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	query := `SELECT is_login_table, attributes FROM "table" WHERE id = $1`

	err = tx.QueryRow(ctx, query, req.Id).Scan(&isLoginTable, &oldAttributes)
	if err != nil {
		return &nb.Table{}, errors.Wrap(err, "when get is_login_table")
	}

	query = `UPDATE "table" SET 
		"label" = $2,
		"icon" = $3,
		"description" = $4,
		"show_in_menu" = $5,
		"subtitle_field_slug" = $6,
		"is_cached" = $7,
		"with_increment_id" = $8,
		"soft_delete" = $9,
		"order_by" = $10,
		"digit_number" = $11,
		"attributes" = $12,
		is_login_table = $13
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
		req.OrderBy,
		req.IncrementId.DigitNumber,
		req.Attributes,
		req.IsLoginTable,
	)
	if err != nil {
		return &nb.Table{}, errors.Wrap(err, "failed to update table")
	}

	query = `SELECT guid FROM "role" `

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return &nb.Table{}, errors.Wrap(err, "failed to select role")
	}
	defer rows.Close()

	query = `SELECT COUNT(*) FROM "record_permission" WHERE table_slug = $1 AND role_id = $2`

	var (
		guids       = []string{}
		createQuery = `INSERT INTO "record_permission" (table_slug, role_id, read, update, write, delete, is_have_condition) 
						VALUES ($1, $2, 'Yes', 'Yes', 'Yes', 'Yes', false)`
	)

	for rows.Next() {
		var guid string

		if err = rows.Scan(&guid); err != nil {
			return &nb.Table{}, errors.Wrap(err, "failed to scan role")
		}

		guids = append(guids, guid)
	}

	for _, guid := range guids {
		var count = 0

		if err = tx.QueryRow(ctx, query, req.Slug, guid).Scan(&count); err != nil {
			return &nb.Table{}, errors.Wrap(err, "failed to select count")
		}

		if count == 0 {
			_, err = tx.Exec(ctx, createQuery, req.Slug, guid)
			if err != nil {
				return &nb.Table{}, errors.Wrap(err, "failed to insert record permission")
			}
		}
	}

	loginStrategyMap := helper.GetLoginStrategyMap(ctx, oldAttributes)
	if req.IsLoginTable {
		attributesMap, err := helper.ConvertStructToMap(req.Attributes)
		if err != nil {
			return &nb.Table{}, errors.Wrap(err, "convert attributes struct to map")
		}

		attributesAuthInfo, ok := attributesMap["auth_info"].(map[string]any)
		if !ok {
			return &nb.Table{}, errors.New("auth_info does not exist")
		}

		loginStrategy, ok := attributesAuthInfo["login_strategy"].([]any)
		if !ok {
			return &nb.Table{}, errors.New("login_strategy does not exist")
		}

		var authInfo = map[string]any{
			"role_id":        "role_id",
			"client_type_id": "client_type_id",
			"login_strategy": loginStrategy,
		}

		for _, value := range loginStrategy {
			strategy := cast.ToString(value)
			oldVal, exist := loginStrategyMap[strategy]

			switch cast.ToString(strategy) {
			case "phone":
				if exist {
					authInfo["phone"] = oldVal
					continue
				}

				phoneAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Phone", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when convert to struct phone field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Phone",
					Slug:       "phone",
					Type:       "INTERNATION_PHONE",
					TableId:    req.Id,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: phoneAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when upsert phone field")
				}
				authInfo["phone"] = "phone"
			case "login":
				if exist {
					authInfo["login"] = oldVal
					authInfo["password"] = "password"
					continue
				}

				loginAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Login", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when convert to struct login field attributes")
				}

				passwordAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Password", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when convert to struct password field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Login",
					Slug:       "login",
					Type:       "SINGLE_LINE",
					TableId:    req.Id,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: loginAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when upsert login field")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Password",
					Slug:       "password",
					Type:       "PASSWORD",
					TableId:    req.Id,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: passwordAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when upsert password field")
				}

				authInfo["login"] = "login"
				authInfo["password"] = "password"
			case "email":
				if exist {
					authInfo["email"] = oldVal
					continue
				}

				emailAttributes, err := helper.ConvertMapToStruct(map[string]any{
					"attributes": map[string]any{"fields": map[string]any{"label_en": map[string]any{"stringValue": "Email", "kind": "stringValue"}}},
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when convert to struct email field attributes")
				}

				err = helper.UpsertLoginTableField(ctx, models.Field{
					Tx:         tx,
					Label:      "Email",
					Slug:       "email",
					Type:       "EMAIL",
					TableId:    req.Id,
					TableSlug:  req.Slug,
					Required:   false,
					ShowLabel:  true,
					Attributes: emailAttributes,
					Default:    "",
					Index:      "string",
				})
				if err != nil {
					return &nb.Table{}, errors.Wrap(err, "when upsert email field")
				}

				authInfo["email"] = "email"
			default:
				return &nb.Table{}, errors.New("Unknown strategy: " + cast.ToString(strategy))
			}
		}

		var clientTypeRelationCount, roleRelationCount int32
		query = `SELECT COUNT(id) 
			FROM "relation" 
			WHERE table_from = $1 AND field_from = 'client_type_id' AND table_to = 'client_type' AND field_to = 'id'`

		err = tx.QueryRow(ctx, query, req.Slug).Scan(&clientTypeRelationCount)
		if err != nil && err != pgx.ErrNoRows {
			return &nb.Table{}, errors.Wrap(err, "when get count client type relaion")
		}

		if clientTypeRelationCount == 0 {
			clientTypeAttributes, err := helper.ConvertMapToStruct(map[string]any{
				"label_en":              "Client Type",
				"label_to_en":           req.Label,
				"table_editable":        false,
				"enable_multi_language": false,
			})
			if err != nil {
				return &nb.Table{}, errors.Wrap(err, "when convert ")
			}

			_, err = helper.CreateRelationWithTx(ctx, &models.CreateRelationRequest{
				Tx:                tx,
				Id:                uuid.NewString(),
				TableFrom:         req.Slug,
				TableTo:           config.CLIENT_TYPE,
				Type:              config.MANY2ONE,
				ViewFields:        []string{"04d0889a-b9ba-4f5c-8473-c8447aab350d"},
				RelationTableSlug: config.CLIENT_TYPE,
				Attributes:        clientTypeAttributes,
				AutoFilters:       []*nb.AutoFilter{{FieldTo: "", FieldFrom: ""}},
				RelationFieldId:   uuid.NewString(),
			})
			if err != nil {
				return &nb.Table{}, errors.Wrap(err, "when create relation")
			}
		}

		query = `SELECT COUNT(id) FROM "relation" 
				WHERE table_from = $1 AND field_from = 'role_id' AND table_to = 'role' AND field_to = 'id'`

		err = tx.QueryRow(ctx, query, req.Slug).Scan(&roleRelationCount)
		if err != nil && err != pgx.ErrNoRows {
			return &nb.Table{}, errors.Wrap(err, "when get count client type relaion")
		}

		if roleRelationCount == 0 {
			roleAttributes, err := helper.ConvertMapToStruct(map[string]any{
				"label_en":              "Role",
				"label_to_en":           req.Label,
				"table_editable":        false,
				"enable_multi_language": false,
			})
			if err != nil {
				return &nb.Table{}, errors.Wrap(err, "when convert role attributes")
			}

			_, err = helper.CreateRelationWithTx(ctx, &models.CreateRelationRequest{
				Tx:                tx,
				Id:                uuid.NewString(),
				TableFrom:         req.Slug,
				TableTo:           config.ROLE,
				Type:              config.MANY2ONE,
				ViewFields:        []string{"c12adfef-2991-4c6a-9dff-b4ab8810f0df"},
				RelationTableSlug: config.ROLE,
				Attributes:        roleAttributes,
				AutoFilters:       []*nb.AutoFilter{{FieldTo: "client_type_id", FieldFrom: "client_type_id"}},
				RelationFieldId:   uuid.NewString(),
			})
			if err != nil {
				return &nb.Table{}, errors.Wrap(err, "when create relation")
			}
		}

		req.Attributes, err = helper.ConvertMapToStruct(map[string]any{"auth_info": authInfo, "label": req.Label, "label_en": req.Label})
		if err != nil {
			return &nb.Table{}, errors.Wrap(err, "when convert to struct auth_info")
		}

		query = `UPDATE "table" SET attributes = $2 WHERE id = $1`

		_, err = tx.Exec(ctx, query, req.Id, req.Attributes)
		if err != nil {
			return &nb.Table{}, errors.Wrap(err, "when update table attributes")
		}

		query = `INSERT INTO "field" ( id, table_id, slug, label, type, is_visible, is_system, attributes) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    		ON CONFLICT (table_id, slug)
    		DO NOTHING`

		_, err = tx.Exec(ctx, query,
			uuid.NewString(), req.Id, "user_id_auth",
			"User ID Auth", "UUID", false, true,
			`{"label_en":"UserIdAuth","label":"UserIdAuth","defaultValue":""}`,
		)
		if err != nil {
			return &nb.Table{}, errors.Wrap(err, "failed to insert field")
		}

		query = `ALTER TABLE IF EXISTS "` + req.GetSlug() + `" ADD COLUMN IF NOT EXISTS ` + ` user_id_auth` + ` UUID`
		_, _ = tx.Exec(ctx, query)
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.Table{}, errors.Wrap(err, "failed to commit transaction")
	}

	return t.GetByID(ctx, &nb.TablePrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (t *tableRepo) Delete(ctx context.Context, req *nb.TablePrimaryKey) error {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.Delete")
	defer dbSpan.Finish()

	var conn = psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var (
		query    = `SELECT is_system FROM "table" WHERE id = $1`
		slug     string
		isSystem sql.NullBool
	)

	err = tx.QueryRow(ctx, query, req.Id).Scan(&isSystem)
	if err != nil {
		return errors.Wrap(err, "failed select from table")
	}

	if isSystem.Valid {
		if isSystem.Bool {
			return errors.New("system table can not be deleted")
		}
	}

	query = `DELETE FROM "table" WHERE id = $1 RETURNING slug`

	err = tx.QueryRow(ctx, query, req.Id).Scan(&slug)
	if err != nil {
		return errors.Wrap(err, "failed to delete table")
	}

	query = `DROP TABLE IF EXISTS ` + slug

	_, err = tx.Exec(ctx, query)
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}

	if err := tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}

func (t *tableRepo) GetTablesByLabel(ctx context.Context, req *nb.GetTablesByLabelReq) (resp *nb.GetAllTablesResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "table.GetTablesByLabel")
	defer dbSpan.Finish()

	conn := psqlpool.Get(req.GetProjectId())

	resp = &nb.GetAllTablesResponse{}

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
		"order_by",
		"digit_number",
		"attributes",
		is_login_table
	FROM "table" WHERE (is_system = false OR (slug = 'role' OR slug = 'client_type')) AND label = $1`

	query += ` ORDER BY created_at DESC `

	rows, err := conn.Query(ctx, query, req.Label)
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
			&table.OrderBy,
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

	return resp, nil
}
