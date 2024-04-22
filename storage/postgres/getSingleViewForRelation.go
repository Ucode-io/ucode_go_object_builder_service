package postgres

import (
	"context"
	"errors"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cast"
)

type relationRepo struct {
	db *pgxpool.Pool
}

func NewRelationRepo(db *pgxpool.Pool) storage.RelationRepoI {
	return &relationRepo{
		db: db,
	}
}

func (r *relationRepo) Create(ctx context.Context, req *nb.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()
	var (
		fieldFrom, fieldTo string
	)

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	switch req.Type {
	case config.MANY2DYNAMIC:
	case config.MANY2MANY:
	case config.MANY2ONE:
	case config.ONE2ONE:
	}

	query := `
		INSERT INTO "relation" (
			"id", 
			"table_from", 
			"table_to", 
			"field_from", 
			"field_to", 
			"type",
			"view_fields", 
			"relation_field_slug", 
			"dynamic_tables", 
			"editable",
			"is_user_id_default", 
			"cascadings", 
			"is_system", 
			"object_id_from_jwt",
			"cascading_tree_table_slug", 
			"cascading_tree_field_slug" 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = conn.Exec(ctx, query,
		req.Id, req.TableFrom, req.TableTo, fieldFrom, fieldTo,
		// req.Type, req.ViewFields, req.reqFieldSlug, req.DynamicTables,
		// req.Editable, req.IsUserIDDefault, req.Cascadings, req.IsSystem,
		// req.ObjectIDFromJWT, req.CascadingTreeTableSlug, req.CascadingTreeFieldSlug,
		// req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		// return fmt.Errorf("failed to insert relation: %v", err)
	}

	return resp, nil
}

func (r *relationRepo) GetByID(ctx context.Context, req *nb.RelationPrimaryKey) (resp *nb.RelationForGetAll, err error) {
	return resp, nil
}

func (r *relationRepo) GetList(ctx context.Context, req *nb.GetAllRelationsRequest) (resp *nb.GetAllRelationsResponse, err error) {
	return resp, nil
}

func (r *relationRepo) Update(ctx context.Context, req *nb.UpdateRelationRequest) (resp *nb.RelationForGetAll, err error) {
	return resp, err
}

func (r *relationRepo) Delete(ctx context.Context, req *nb.RelationPrimaryKey) (err error) {
	return nil
}

func (r *relationRepo) GetSingleViewForRelation(ctx context.Context, req models.ReqForViewRelation) (resp models.RelationForView, err error) {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return resp, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return resp, err
	}
	defer conn.Close()

	var tableId string

	table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Slug: req.TableSlug})
	if err != nil {
		return resp, err
	}
	if table != nil {
		cast.ToString(table["id"])
		table["id"] = tableId
	}

	query := `
		SELECT 
			r.id AS relation_id,
			t.slug AS table_slug,
			t.id AS table_id
		FROM 
			relation r
		LEFT JOIN 
			table t ON t.id = r.table_id
		LEFT JOIN 
			fields f ON f.relation_id = r.id
		WHERE 
			r.id = $1
		ORDER BY 
			r.created_at DESC
		LIMIT 1
	`

	rows, err := conn.Query(ctx, query, req.Id)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		resp = models.RelationForView{}
		if err := rows.Scan(&resp.Id, &resp.TableFrom, &tableId); err != nil {
			return resp, err
		}
	}
	if resp.Id == "" {
		return resp, errors.New("no data found")
	}
	if err := rows.Err(); err != nil {
		return resp, err
	}

	tableFrom, err := helper.TableVer(ctx, helper.TableVerReq{
		Conn: conn,
		Slug: req.TableSlug,
	})
	if err != nil {
		return resp, err
	}
	if resp.Type == config.MANY2DYNAMIC {
		for _, dynamicTable := range resp.DynamicTables {

			if dynamicTable.TableSlug == req.TableSlug || cast.ToString(table["slug"]) == req.TableSlug {
				tableTo, err := helper.TableVer(ctx, helper.TableVerReq{
					Conn: conn,
					Slug: dynamicTable.TableSlug,
				})
				if err != nil {
					return resp, err
				}
				view := &nb.View{}
				err = conn.QueryRow(ctx, `
					SELECT * FROM view
					WHERE relation_id = $1
				`, req.Id).Scan(
					&view.RelationId,
					&view.TableSlug,
				)
				if err != nil {
					return resp, err
				}

				viewFieldsInDynamicTable := []nb.Field{}
				for _, fieldID := range dynamicTable.ViewFields {

					query := `SELECT id, slug, table_id , attributes,  FROM field WHERE id = $1`
					view_field := &nb.Field{}
					err = conn.QueryRow(ctx, query, fieldID).Scan(
						&view_field.Id,
						&view_field.Slug,
						&view_field.TableId,
						&view_field.Attributes,
					)

					if view_field != nil {
						viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, nb.Field{
							Id:         view_field.Id,
							Slug:       view_field.Slug,
							TableId:    view_field.TableId,
							Attributes: view_field.Attributes,
						})
					}
					responseRelation := map[string]interface{}{
						"id":                        resp.Id,
						"table_from":                tableFrom,
						"table_to":                  tableTo,
						"type":                      resp.Type,
						"view_fields":               viewFieldsInDynamicTable,
						"editable":                  resp.Editable,
						"dynamic_tables":            resp.DynamicTables,
						"relation_field_slug":       resp.RelationFieldSlug,
						"auto_filters":              resp.AutoFilters,
						"is_user_id_default":        resp.IsUserIdDefault,
						"cascadings":                resp.Cascadings,
						"object_id_from_jwt":        resp.ObjectIdFromJwt,
						"cascading_tree_table_slug": resp.CascadingTreeTableSlug,
						"cascading_tree_field_slug": resp.CascadingTreeFieldSlug,
					}

					if view != nil {
						responseRelation["title"] = view.Name
						responseRelation["columns"] = view.Columns
						responseRelation["quick_filters"] = view.QuickFilters
						responseRelation["group_fields"] = view.GroupFields
						responseRelation["is_editable"] = view.IsEditable
						responseRelation["relation_table_slug"] = view.RelationTableSlug
						responseRelation["view_type"] = view.Type
						responseRelation["summaries"] = view.Summaries
						responseRelation["relation_id"] = view.RelationId
						responseRelation["default_values"] = view.DefaultValues
						responseRelation["action_relations"] = view.ActionRelations
						responseRelation["default_limit"] = view.DefaultLimit
						responseRelation["multiple_insert"] = view.MultipleInsert
						responseRelation["multiple_insert_field"] = view.MultipleInsertField
						responseRelation["updated_fields"] = view.UpdatedFields
						responseRelation["attributes"] = view.Attributes
					}
				}
			}
		}
	}
	tableTo, err := helper.TableVer(ctx, helper.TableVerReq{
		Conn: conn,
		Slug: req.TableSlug,
	})
	if err != nil {
		return resp, err
	}
	query = `
		SELECT relation_id, table_slug FROM view
		WHERE relation_id = $1
	`
	view := &nb.View{}
	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&view.RelationId,
		&view.TableSlug,
	)
	if err != nil {
		return resp, err
	}
	if err != nil {
		return resp, err
	}
	responseRelation := map[string]interface{}{
		"id":                        resp.Id,
		"table_from":                tableFrom,
		"table_to":                  tableTo,
		"type":                      resp.Type,
		"view_fields":               resp.ViewFields,
		"editable":                  resp.Editable,
		"dynamic_tables":            resp.DynamicTables,
		"relation_field_slug":       resp.RelationFieldSlug,
		"auto_filters":              resp.AutoFilters,
		"is_user_id_default":        resp.IsUserIdDefault,
		"cascadings":                resp.Cascadings,
		"object_id_from_jwt":        resp.ObjectIdFromJwt,
		"cascading_tree_table_slug": resp.CascadingTreeTableSlug,
		"cascading_tree_field_slug": resp.CascadingTreeFieldSlug,
	}
	if view != nil {
		responseRelation["title"] = view.Name
		responseRelation["columns"] = view.Columns
		responseRelation["quick_filters"] = view.QuickFilters
		responseRelation["group_fields"] = view.GroupFields
		responseRelation["is_editable"] = view.IsEditable
		responseRelation["relation_table_slug"] = view.RelationTableSlug
		responseRelation["view_type"] = view.Type
		responseRelation["summaries"] = view.Summaries
		responseRelation["relation_id"] = view.RelationId
		responseRelation["default_values"] = view.DefaultValues
		responseRelation["action_relations"] = view.ActionRelations
		responseRelation["default_limit"] = view.DefaultLimit
		responseRelation["multiple_insert"] = view.MultipleInsert
		responseRelation["multiple_insert_field"] = view.MultipleInsertField
		responseRelation["updated_fields"] = view.UpdatedFields
		responseRelation["attributes"] = view.Attributes
	}
	relationTabWithPermission, err := helper.AddPermissionToTab(ctx, responseRelation, conn, req.RoleId, req.TableSlug, req.ProjectId)
	if err != nil {
		return resp, err
	}

	resp.Id = cast.ToString(relationTabWithPermission["id"])
	return resp, nil

}
