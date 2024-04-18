package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type viewRepo struct {
	db *pgxpool.Pool
}

func NewViewRepo(db *pgxpool.Pool) storage.ViewRepoI {
	return &viewRepo{
		db: db,
	}
}

func (v viewRepo) Create(ctx context.Context, req *nb.CreateViewRequest) (resp *nb.View, err error) {
	resp = &nb.View{}
	viewID := uuid.New().String()

	attributes, err := protojson.Marshal(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshaling attributes: %v", err)
	}

	_, err = v.db.Exec(ctx, `
        INSERT INTO "view" (
			"id",
			"table_slug",
			"type",
			"group_fields",
			"view_fields",
			"main_field",
			"disable_dates",
			"quick_filters",
			"users",
			"name",
			"columns",
			"calendar_from_slug",
			"calendar_to_slug",
			"time_interval",
			"multiple_insert",
			"status_field_slug",
			"is_editable",
			"relation_table_slug",
			"relation_id",
			"multiple_insert_field",
			"updated_fields",
			"app_id",
			"table_label",
			"default_limit",
			"default_editable",
			"order",
			"name_uz",
			"name_en",
			"attributes"
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
    `, viewID,
		req.TableSlug,
		req.Type,
		req.GroupFields,
		req.ViewFields,
		req.MainField,
		req.DisableDates,
		req.QuickFilters,
		req.Users,
		req.Name,
		req.Columns,
		req.CalendarFromSlug,
		req.CalendarToSlug,
		req.TimeInterval,
		req.MultipleInsert,
		req.StatusFieldSlug,
		req.IsEditable,
		req.RelationTableSlug,
		req.RelationId,
		req.MultipleInsertField,
		req.UpdatedFields,
		req.AppId,
		req.TableLabel,
		req.DefaultLimit,
		req.DefaultEditable,
		req.Order,
		req.NameUz,
		req.NameEn,
		attributes,
	)
	if err != nil {
		return nil, err
	}

	var data = []byte(`{}`)
	data, err = helper.ChangeHostname(data)
	if err != nil {
		return nil, err
	}

	_, err = v.db.Exec(ctx, `
        UPDATE "table" SET is_changed = true, is_changed_by_host = $2
        WHERE "slug" = $1
    `, req.TableSlug,
		data)

	if err != nil {
		return nil, err
	}

	roles, err := v.db.Query(ctx, `
        SELECT guid FROM role
    `)
	if err != nil {
		return nil, err
	}
	defer roles.Close()

	for roles.Next() {
		var roleID string
		if err := roles.Scan(&roleID); err != nil {
			return nil, err
		}

		_, err := v.db.Exec(ctx, `
            INSERT INTO view_permission (guid, view_id, role_id, "view", "edit", "delete")
            VALUES ($1, $2, $3, true, true, true)
        `, uuid.New().String(), viewID, roleID)
		if err != nil {
			return nil, err
		}
	}

	return v.GetSingle(ctx, &nb.ViewPrimaryKey{Id: viewID, ProjectId: req.ProjectId})
}

func (v viewRepo) GetList(ctx context.Context, req *nb.GetAllViewsRequest) (resp *nb.GetAllViewsResponse, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	fmt.Println("Table slug==", req.TableSlug)

	resp = &nb.GetAllViewsResponse{}
	query := `
        SELECT 
			COUNT(*) OVER() AS count,
			"id",
			"table_slug",
			"type",
			"name",
			"main_field",
			"disable_dates",
			"columns",
			"quick_filters",
			"users",
			"view_fields",
			"group_fields",
			"calendar_from_slug",
			"calendar_to_slug",
			"time_interval",
			"multiple_insert",
			"status_field_slug",
			"is_editable",
			"relation_table_slug",
			"relation_id",
			"multiple_insert_field",
			"updated_fields",
			"app_id",
			"table_label",
			"default_limit",
			"attributes",
			"default_editable",
			"navigate",
			"function_path",
			"order",
			"name_uz",
			"name_en"
	        FROM view
        WHERE table_slug = $1
        ORDER BY "order" ASC
    `

	rows, err := v.db.Query(ctx, query, req.TableSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		TableSlug           sql.NullString
		Type                sql.NullString
		Name                sql.NullString
		MainField           sql.NullString
		CalendarFromSlug    sql.NullString
		CalendarToSlug      sql.NullString
		TimeInterval        sql.NullInt32
		StatusFieldSlug     sql.NullString
		RelationTableSlug   sql.NullString
		RelationId          sql.NullString
		MultipleInsertField sql.NullString
		AppId               sql.NullString
		TableLabel          sql.NullString
		DefaultLimit        sql.NullString
		FunctionPath        sql.NullString
		Order               sql.NullInt32
		NameUz              sql.NullString
		NameEn              sql.NullString
	)
	for rows.Next() {
		row := &nb.View{}
		attributes := []byte{}

		err = rows.Scan(
			&resp.Count,
			&row.Id,
			&TableSlug,
			&Type,
			&Name,
			&MainField,
			&row.DisableDates,
			&row.Columns,
			&row.QuickFilters,
			&row.Users,
			&row.ViewFields,
			&row.GroupFields,
			&CalendarFromSlug,
			&CalendarToSlug,
			&TimeInterval,
			&row.MultipleInsert,
			&StatusFieldSlug,
			&row.IsEditable,
			&RelationTableSlug,
			&RelationId,
			&MultipleInsertField,
			&row.UpdatedFields,
			&AppId,
			&TableLabel,
			&DefaultLimit,
			&row.Attributes,
			&row.DefaultEditable,
			&row.Navigate,
			&FunctionPath,
			&Order,
			&NameUz,
			&NameEn,
		)
		if err != nil {
			return nil, err
		}

		resp.Views = append(resp.Views, &nb.View{
			Id:                  row.Id,
			TableSlug:           TableSlug.String,
			Type:                Type.String,
			Name:                Name.String,
			MainField:           MainField.String,
			DisableDates:        row.DisableDates,
			Columns:             row.Columns,
			QuickFilters:        row.QuickFilters,
			Users:               row.Users,
			ViewFields:          row.ViewFields,
			GroupFields:         row.GroupFields,
			CalendarFromSlug:    CalendarFromSlug.String,
			CalendarToSlug:      CalendarToSlug.String,
			TimeInterval:        TimeInterval.Int32,
			MultipleInsert:      row.MultipleInsert,
			StatusFieldSlug:     StatusFieldSlug.String,
			IsEditable:          row.IsEditable,
			RelationTableSlug:   RelationTableSlug.String,
			RelationId:          RelationId.String,
			MultipleInsertField: MultipleInsertField.String,
			UpdatedFields:       row.UpdatedFields,
			AppId:               AppId.String,
			TableLabel:          TableLabel.String,
			DefaultLimit:        DefaultLimit.String,
			DefaultEditable:     row.DefaultEditable,
			Navigate:            row.Navigate,
			FunctionPath:        FunctionPath.String,
			Order:               Order.Int32,
			NameUz:              NameUz.String,
			NameEn:              NameEn.String,
			Attributes:          row.Attributes,
		})

		if len(attributes) > 0 {
			err = json.Unmarshal(attributes, &row.Attributes)
			if err != nil {
				return nil, err
			}
		}

		permissionsQuery := `
			SELECT role_id
			FROM view_permission
			WHERE view_id = $1 AND role_id = $2
`
		rows, err := v.db.Query(ctx, permissionsQuery, row.Id, req.RoleId)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		if rows.Next() {
			attributesMap := row.Attributes.AsMap()
			attributesMap["view_permissions"] = true
			row.Attributes, _ = structpb.NewStruct(attributesMap)
		} else {
			attributesMap := row.Attributes.AsMap()
			attributesMap["view_permissions"] = false
			row.Attributes, _ = structpb.NewStruct(attributesMap)
		}

		encodedAttributes, err := json.Marshal(row.Attributes)
		if err != nil {
			return nil, err
		}

		var attributesMap map[string]interface{}
		err = json.Unmarshal(encodedAttributes, &attributesMap)
		if err != nil {
			return nil, err
		}

		structAttributes, err := structpb.NewStruct(attributesMap)
		if err != nil {
			return nil, err
		}

		row.Attributes = structAttributes

	}
	return

}
func (v *viewRepo) GetSingle(ctx context.Context, req *nb.ViewPrimaryKey) (resp *nb.View, err error) {
	resp = &nb.View{}
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	query := `
		SELECT 
			"id",
			"table_slug",
			"type",
			"name",
			"main_field",
			"disable_dates",
			"columns",
			"quick_filters",
			"users",
			"view_fields",
			"group_fields",
			"calendar_from_slug",
			"calendar_to_slug",
			"time_interval",
			"multiple_insert",
			"status_field_slug",
			"is_editable",
			"relation_table_slug",
			"relation_id",
			"multiple_insert_field",
			"updated_fields",
			"app_id",
			"table_label",
			"default_limit",
			"default_editable",
			"navigate",
			"function_path",
			"order",
			"name_uz",
			"name_en",
			"attributes"
			FROM "view" WHERE id = $1`

	var (
		attributes          []byte
		TableSlug           sql.NullString
		Type                sql.NullString
		Name                sql.NullString
		MainField           sql.NullString
		CalendarFromSlug    sql.NullString
		CalendarToSlug      sql.NullString
		TimeInterval        sql.NullInt32
		StatusFieldSlug     sql.NullString
		RelationTableSlug   sql.NullString
		RelationId          sql.NullString
		MultipleInsertField sql.NullString
		AppId               sql.NullString
		TableLabel          sql.NullString
		DefaultLimit        sql.NullString
		FunctionPath        sql.NullString
		Order               sql.NullInt32
		NameUz              sql.NullString
		NameEn              sql.NullString
	)

	err = v.db.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&TableSlug,
		&Type,
		&Name,
		&MainField,
		&resp.DisableDates,
		&resp.Columns,
		&resp.QuickFilters,
		&resp.Users,
		&resp.ViewFields,
		&resp.GroupFields,
		&CalendarFromSlug,
		&CalendarToSlug,
		&TimeInterval,
		&resp.MultipleInsert,
		&StatusFieldSlug,
		&resp.IsEditable,
		&RelationTableSlug,
		&RelationId,
		&MultipleInsertField,
		&resp.UpdatedFields,
		&AppId,
		&TableLabel,
		&DefaultLimit,
		&resp.DefaultEditable,
		&resp.Navigate,
		&FunctionPath,
		&Order,
		&NameUz,
		&NameEn,
		&attributes,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return nil, err
	}

	resp = &nb.View{
		Id:                  resp.Id,
		TableSlug:           TableSlug.String,
		Type:                Type.String,
		Name:                Name.String,
		MainField:           MainField.String,
		DisableDates:        resp.DisableDates,
		Columns:             resp.Columns,
		QuickFilters:        resp.QuickFilters,
		Users:               resp.Users,
		ViewFields:          resp.ViewFields,
		GroupFields:         resp.GroupFields,
		CalendarFromSlug:    CalendarFromSlug.String,
		CalendarToSlug:      CalendarToSlug.String,
		TimeInterval:        TimeInterval.Int32,
		MultipleInsert:      resp.MultipleInsert,
		StatusFieldSlug:     StatusFieldSlug.String,
		IsEditable:          resp.IsEditable,
		RelationTableSlug:   RelationTableSlug.String,
		RelationId:          RelationId.String,
		MultipleInsertField: MultipleInsertField.String,
		UpdatedFields:       resp.UpdatedFields,
		AppId:               AppId.String,
		TableLabel:          TableLabel.String,
		DefaultLimit:        DefaultLimit.String,
		DefaultEditable:     resp.DefaultEditable,
		Navigate:            resp.Navigate,
		FunctionPath:        FunctionPath.String,
		Order:               Order.Int32,
		NameUz:              NameUz.String,
		NameEn:              NameEn.String,
		Attributes:          resp.Attributes,
	}

	return resp, nil
}

func (v viewRepo) Update(ctx context.Context, req *nb.View) (resp *nb.View, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return nil, err
	}

	_, err = v.db.Exec(ctx, `
		UPDATE view
		SET
			table_slug = $2,
			type = $3,
			group_fields = $4,
			view_fields = $5,
			main_field = $6,
			disable_dates = $7,
			quick_filters = $8,
			users = $9,
			name = $10,
			columns = $11,
			calendar_from_slug = $12,
			calendar_to_slug = $13,
			time_interval = $14,
			multiple_insert = $15,
			status_field_slug = $16,
			is_editable = $17,
			relation_table_slug = $18,
			relation_id = $19,
			multiple_insert_field = $20,
			updated_fields = $21,
			app_id = $22,
			table_label = $23,
			default_limit = $24,
			attributes = $25,
			default_editable = $26,
			"order" = $27,
			name_uz = $28,
			name_en = $29
		WHERE id = $1
	`, req.Id,
		req.TableSlug,
		req.Type,
		req.GroupFields,
		req.ViewFields,
		req.MainField,
		req.DisableDates,
		req.QuickFilters,
		req.Users,
		req.Name,
		req.Columns,
		req.CalendarFromSlug,
		req.CalendarToSlug,
		req.TimeInterval,
		req.MultipleInsert,
		req.StatusFieldSlug,
		req.IsEditable,
		req.RelationTableSlug,
		req.RelationId,
		req.MultipleInsertField,
		req.UpdatedFields,
		req.AppId,
		req.TableLabel,
		req.DefaultLimit,
		attributes,
		req.DefaultEditable,
		req.Order,
		req.NameUz,
		req.NameEn,
	)
	if err != nil {
		return nil, err
	}

	var data = []byte(`{}`)
	data, err = helper.ChangeHostname(data)
	if err != nil {
		return nil, err
	}
	_, err = v.db.Exec(ctx, `
	UPDATE "table" 
	SET 
		is_changed = true,
		is_changed_by_host = $2, 
		updated_at = NOW()
	WHERE 
		slug = $1
	`, req.TableSlug, data)
	if err != nil {
		return nil, err
	}

	return v.GetSingle(ctx, &nb.ViewPrimaryKey{Id: req.Id})
}

func (v *viewRepo) Delete(ctx context.Context, req *nb.ViewPrimaryKey) error {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	var data = []byte(`{}`)
	data, err := helper.ChangeHostname(data)
	if err != nil {
		return err
	}

	var (
		filter    string
		condition interface{}
	)
	if req.Id != "" {
		filter = "id"
		condition = req.Id
	} else if req.TableSlug != "" {
		filter = "table_slug"
		condition = req.TableSlug
	}

	query := `
		SELECT
			id,
			table_slug
		FROM view
		WHERE ` + filter + ` = $1
	`

	var (
		id        sql.NullString
		tableSlug sql.NullString
	)

	row := v.db.QueryRow(ctx, query, condition)
	err = row.Scan(&id, &tableSlug)
	if err != nil {
		return err
	}

	_, err = v.db.Exec(ctx, `
		UPDATE "table"
		SET is_changed = true, is_changed_by_host = $2
		WHERE "slug" = $1
	`, tableSlug.String, data)
	if err != nil {
		return err
	}

	_, err = v.db.Exec(ctx, fmt.Sprintf("DELETE FROM view WHERE %v = $1", filter), condition)
	if err != nil {
		return err
	}

	return nil
}
