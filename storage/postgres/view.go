package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
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
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	resp = &nb.View{}
	viewId := uuid.New().String()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	fmt.Println(req.Attributes)

	attributes, err := protojson.Marshal(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshaling attributes: %v", err)
	}

	_, err = v.db.Exec(ctx, `
        INSERT INTO view (
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
			"attributes",
			"default_editable",
			"order",
			"name_uz",
			"name_en"
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
    `, viewId,
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
		fmt.Println("--------------------------")
		return nil, err
	}

	_, err = v.db.Exec(ctx, `
        UPDATE "table" SET is_changed = true, is_changed_by_host = jsonb_set(is_changed_by_host, '{`+hostname+`}', 'true')
        WHERE "slug" = $1
    `, req.TableSlug)

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
		err = roles.Scan(&roleID)
		if err != nil {
			return nil, err
		}

		_, err := v.db.Exec(ctx, `
            INSERT INTO view_permission (guid, view_id, role_id, "view", "edit", "delete")
            VALUES ($1, $2, $3, true, true, true)
        `, uuid.New().String(), viewId, roleID)
		if err != nil {
			return nil, err
		}
	}

	return //v.GetSingle(ctx, &nb.ViewPrimaryKey{Id: viewId, ProjectId: req.ProjectId})
}

func (v viewRepo) GetList(ctx context.Context, req *nb.GetAllViewsRequest) (resp *nb.GetAllViewsResponse, err error) {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

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

	for rows.Next() {
		row := &nb.View{}
		attributes := []byte{}

		err = rows.Scan(
			&resp.Count,
			&row.Id,
			&row.TableSlug,
			&row.Type,
			&row.Name,
			&row.MainField,
			&row.DisableDates,
			&row.Columns,
			&row.QuickFilters,
			&row.Users,
			&row.ViewFields,
			&row.GroupFields,
			&row.CalendarFromSlug,
			&row.CalendarToSlug,
			&row.TimeInterval,
			&row.MultipleInsert,
			&row.StatusFieldSlug,
			&row.IsEditable,
			&row.RelationTableSlug,
			&row.RelationId,
			&row.MultipleInsertField,
			&row.UpdatedFields,
			&row.AppId,
			&row.TableLabel,
			&row.DefaultLimit,
			&row.Attributes,
			&row.DefaultEditable,
			&row.Navigate,
			&row.FunctionPath,
			&row.Order,
			&row.NameUz,
			&row.NameEn,
		)
		if err != nil {
			return nil, err
		}

		resp.Views = append(resp.Views, row)

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
	return resp, nil

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
			"attributes",
			"default_editable",
			"navigate",
			"function_path",
			"order",
			"name_uz",
			"name_en"
			FROM "view" WHERE id = $1`

	var attributes []byte

	err = v.db.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.TableSlug,
		&resp.Type,
		&resp.Name,
		&resp.MainField,
		&resp.DisableDates,
		&resp.Columns,
		&resp.QuickFilters,
		&resp.Users,
		&resp.ViewFields,
		&resp.GroupFields,
		&resp.CalendarFromSlug,
		&resp.CalendarToSlug,
		&resp.TimeInterval,
		&resp.MultipleInsert,
		&resp.StatusFieldSlug,
		&resp.IsEditable,
		&resp.RelationTableSlug,
		&resp.RelationId,
		&resp.MultipleInsertField,
		&resp.UpdatedFields,
		&resp.AppId,
		&resp.TableLabel,
		&resp.DefaultLimit,
		&attributes,
		&resp.DefaultEditable,
		&resp.Navigate,
		&resp.FunctionPath,
		&resp.Order,
		&resp.NameUz,
		&resp.NameEn,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return nil, err
	}

	return resp, nil
}

func (v *viewRepo) Delete(ctx context.Context, req *nb.ViewPrimaryKey) error {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	var filter string
	if req.Id != "" {
		filter = " id = $1"
	} else if req.TableSlug != "" {
		filter = " table_slug = $1"
	}

	query := `
		SELECT 
			id,
			table_slug
		FROM view
		WHERE ` + filter + `
	`
	row := v.db.QueryRow(ctx, query, req.Id)
	view := &nb.View{}
	err := row.Scan(&view.Id, &view.TableSlug)
	if err != nil {
		return err
	}

	hostname, _ := os.Hostname()

	_, err = v.db.Exec(ctx, `
		UPDATE table
		SET is_changed = true, is_changed_by_host = $2
		WHERE "slug" = $1
	`, view.TableSlug, hostname)
	if err != nil {
		return err
	}

	_, err = v.db.Exec(ctx, `DELETE FROM view WHERE `+filter, req.Id)
	if err != nil {
		return err
	}

	return nil
}

func (v viewRepo) Update(ctx context.Context, req *nb.View) (resp *nb.View, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

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
		req.DefaultLimit, attributes,
		req.DefaultEditable,
		req.Order,
		req.NameUz,
		req.NameEn,
	)
	if err != nil {
		return nil, err
	}

	_, err = v.db.Exec(ctx, `
	UPDATE table SET is_changed = true, is_changed_by_host = jsonb_set(is_changed_by_host, '{`+hostname+`}', 'true')
	WHERE slug = $1
	`, req.TableSlug)
	if err != nil {
		return nil, err
	}

	resp, err = v.GetSingle(ctx, &nb.ViewPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
	if err != nil {
		return nil, err
	}

	return resp, nil
}
