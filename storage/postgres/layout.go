package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

type layoutRepo struct {
	db *pgxpool.Pool
}

func NewLayoutRepo(db *pgxpool.Pool) storage.LayoutRepoI {
	return &layoutRepo{
		db: db,
	}
}

func (l *layoutRepo) Update(ctx context.Context, req *nb.LayoutRequest) (resp *nb.LayoutResponse, err error) {

	resp = &nb.LayoutResponse{}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	var layoutId string
	if req.Id == "" {
		layoutId = uuid.New().String()
	} else {
		layoutId = req.Id
	}

	var tableSlug1 string
	result, err := helper.TableVer(ctx, helper.TableVerReq{
		Conn: conn,
		Id:   req.TableId,
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching table: %w", err)
	}

	if tableSlug, ok := result["slug"].(string); ok {
		tableSlug1 = tableSlug
	} else {
		return nil, fmt.Errorf("tableSlug not found or not a string")
	}

	query := `
        INSERT INTO "layout" (
            "id", "label", "order", "type", "icon", "is_default", 
            "is_modal", "is_visible_section",
             "table_id", "menu_id"
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        ON CONFLICT (id) DO UPDATE
        SET 
            "label" = EXCLUDED.label,
            "order" = EXCLUDED.order,
            "type" = EXCLUDED.type,
            "icon" = EXCLUDED.icon,
            "is_default" = EXCLUDED.is_default,
            "is_modal" = EXCLUDED.is_modal,
            "is_visible_section" = EXCLUDED.is_visible_section,
            "table_id" = EXCLUDED.table_id,
            "menu_id" = EXCLUDED.menu_id
    `
	_, err = tx.Exec(ctx, query,
		layoutId, req.Label, req.Order, req.Type, req.Icon,
		req.IsDefault, req.IsModal, req.IsVisibleSection,
		req.TableId, req.MenuId)
	if err != nil {
		return nil, fmt.Errorf("error inserting layout: %w", err)
	}

	if req.IsDefault {
		_, err = tx.Exec(ctx, `
            UPDATE layout
            SET is_default = false
            WHERE table_id = $1 AND id != $2
        `, req.TableId, layoutId)
		if err != nil {
			return nil, fmt.Errorf("error updating layout: %w", err)
		}
	}

	var (
		bulkWriteTab                  []string
		bulkWriteSection              []string
		mapTabs                       = make(map[string]int)
		mapSections                   = make(map[string]int)
		deletedTabIds                 []string
		deletedSectionIds             []string
		relationIds                   []string
		insertManyRelationPermissions []string
	)

	rows, err := tx.Query(ctx, "SELECT id FROM tab WHERE layout_id = $1", layoutId)

	if err != nil {
		return nil, fmt.Errorf("error fetching tabs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tabId string
		if err := rows.Scan(&tabId); err != nil {
			return nil, fmt.Errorf("error scanning tab ID: %w", err)

		}
		mapTabs[tabId] = 1
	}

	for i, tab := range req.Tabs {
		if tab.Id == "" {
			tab.Id = uuid.New().String()
		}
		if tab.Type == "relation" {
			relationIds = append(relationIds, tab.RelationId)
		}

		if _, ok := mapTabs[tab.Id]; ok {
			mapTabs[tab.Id] = 2
		}
		//attributesJSON, err := json.Marshal(tab.Attributes)
		//if err != nil {
		//	return nil, fmt.Errorf("error marshaling attributes to JSON: %w", err)
		//}

		bulkWriteTab = append(bulkWriteTab, fmt.Sprintf(`
			INSERT INTO "tab" (
				"id", "label", "layout_id",  "type",
				"order", "icon", relation_id
			) VALUES ('%s', '%s', '%s', '%s', %d, '%s', '%s')
			ON CONFLICT (id) DO UPDATE
			SET
				"label" = EXCLUDED.label,
				"layout_id" = EXCLUDED.layout_id,
				"type" = EXCLUDED.type,
				"order" = EXCLUDED.order,
				"icon" = EXCLUDED.icon,
				"relation_id" = EXCLUDED.relation_id
		`, tab.Id, tab.Label, layoutId, tab.Type,
			i, tab.Icon, tab.RelationId))

		for _, query := range bulkWriteTab {
			_, err := tx.Exec(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("error executing bulkWriteTab query: %w", err)
			}
		}
		for i, section := range tab.Sections {
			if section.Id == "" {
				section.Id = uuid.New().String()
			}

			if _, ok := mapSections[section.Id]; ok {
				mapSections[section.Id] = 2
			}

			for _, section := range tab.Sections {
				if section.Id == "" {
					section.Id = uuid.New().String()

				}
				if _, ok := mapSections[section.Id]; ok {
					mapSections[section.Id] = 2
				}
				bulkWriteSection = append(bulkWriteSection, fmt.Sprintf(`
					INSERT INTO "section" (
						"id", "tab_id", "label", "order", "icon", 
						"column",  is_summary_section
					) VALUES ('%s', '%s', '%s', %d, '%s', '%s', '%t')
					ON CONFLICT (id) DO UPDATE
					SET
						"tab_id" = EXCLUDED.tab_id,
						"label" = EXCLUDED.label,
						"order" = EXCLUDED.order,
						"icon" = EXCLUDED.icon,
						"column" = EXCLUDED.column,
						"is_summary_section" = EXCLUDED.is_summary_section
				`, section.Id, tab.Id, section.Label, i, section.Icon, section.Column, section.IsSummarySection))
			}

			for _, query := range bulkWriteSection {
				_, err := tx.Exec(ctx, query)
				if err != nil {
					tx.Rollback(ctx)
					return nil, fmt.Errorf("error executing bulkWriteSection query: %w", err)
				}
			}
		}
	}
	for key, value := range mapTabs {
		if value == 1 {
			deletedTabIds = append(deletedTabIds, key)
		}
	}

	for key, value := range mapSections {
		if value == 1 {
			deletedSectionIds = append(deletedSectionIds, key)
		}
	}

	if len(deletedTabIds) > 0 {
		_, err = tx.Exec(ctx, "DELETE FROM tab WHERE id = ANY($1)", pq.Array(deletedTabIds))
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error deleting tabs: %w", err)
		}
	}

	if len(deletedSectionIds) > 0 {
		_, err = tx.Exec(ctx, "DELETE FROM section WHERE id = ANY($1)", pq.Array(deletedSectionIds))
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error deleting sections: %w", err)

		}
	}
	rows, err = tx.Query(ctx, "SELECT guid FROM role")
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("error fetching roles: %w", err)
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		var roleGUID string
		if err := rows.Scan(&roleGUID); err != nil {
			return nil, fmt.Errorf("error scanning role GUID: %w", err)
		}
		roles = append(roles, roleGUID)
	}

	for _, role := range roles {

		for _, relationID := range relationIds {
			var exists int
			query := `
			        SELECT COUNT(*)
			        FROM view_relation_permission
			        WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
			    `

			err := tx.QueryRow(ctx, query, role, req.TableId, relationID).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking relation permission existence: %w", err)
			}

			if exists == 0 {
				insertManyRelationPermissions = append(insertManyRelationPermissions, fmt.Sprintf(`
			            INSERT INTO view_relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
			            VALUES ('%s', '%s', '%s', true, true, true, true)
			        `, role, req.TableId, relationID))
			}
		}

		if len(insertManyRelationPermissions) > 0 {
			for _, query := range insertManyRelationPermissions {
				_, err := tx.Exec(ctx, query)
				if err != nil {
					return nil, fmt.Errorf("error inserting relation permissions: %w", err)
				}
			}
		}

	}

	if len(insertManyRelationPermissions) > 0 {
		for _, role := range roles {
		
			for _, relationID := range relationIds {

				var relationPermission bool
				err := tx.QueryRow(ctx, `
					SELECT EXISTS (
						SELECT 1 
						FROM view_relation_permission 
						WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
					)`, role, tableSlug1, relationID).Scan(&relationPermission)
				if err != nil {
					tx.Rollback(ctx)
					return nil, err
				}
				if !relationPermission {
					_, err := tx.Exec(ctx, `
						INSERT INTO view_relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
						VALUES ($1, $2, $3, true, true, true, true)`, role, tableSlug1, relationID)
					if err != nil {
						tx.Rollback(ctx)
						return nil, err
					}
				}
			}
		}
	}

	if len(deletedTabIds) > 0 {
		_, err := tx.Exec(ctx, "DELETE FROM tab WHERE id = ANY($1)", pq.Array(deletedTabIds))
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error deleting tabs: %w", err)
		}
	}

	if len(deletedSectionIds) > 0 {
		_, err := tx.Exec(ctx, "DELETE FROM section WHERE id = ANY($1)", pq.Array(deletedSectionIds))
		if err != nil {
			tx.Rollback(ctx)
			return nil, err
		}
	}

	resp = &nb.LayoutResponse{
		Id:               layoutId,
		Label:            req.Label,
		Order:            req.Order,
		Type:             req.Type,
		Icon:             req.Icon,
		IsDefault:        req.IsDefault,
		IsModal:          req.IsModal,
		IsVisibleSection: req.IsVisibleSection,
		Attributes:       req.Attributes,
		TableId:          req.TableId,
		MenuId:           req.MenuId,
		Tabs: []*nb.TabResponse{
			{
				Id:         req.Tabs[0].Id,
				Label:      req.Tabs[0].Label,
				RelationId: req.Tabs[0].RelationId,
				Type:       req.Tabs[0].Type,
				Order:      req.Tabs[0].Order,
				Icon:       req.Tabs[0].Icon,
				LayoutId:   layoutId,
				Relation: &nb.RelationForSection{
					Id: req.Tabs[0].RelationId,
					TableFrom: &nb.TableForSection{
						Id:    req.TableId,
						Label: tableSlug1,
					},
				},
				Attributes: req.Tabs[0].Attributes,
				Sections: []*nb.SectionResponse{
					{
						Id:               req.Tabs[0].Sections[0].Id,
						Label:            req.Tabs[0].Sections[0].Label,
						Order:            req.Tabs[0].Sections[0].Order,
						Icon:             req.Tabs[0].Sections[0].Icon,
						Column:           req.Tabs[0].Sections[0].Column,
						IsSummarySection: req.Tabs[0].Sections[0].IsSummarySection,
						Fields: []*nb.FieldResponse{
							{
								Id:              req.Tabs[0].Sections[0].Fields[0].Id,
								Label:           req.Tabs[0].Sections[0].Fields[0].FieldName,
								Order:           req.Tabs[0].Sections[0].Fields[0].Order,
								Column:          req.Tabs[0].Sections[0].Fields[0].Column,
								RelationType:    req.Tabs[0].Sections[0].Fields[0].RelationType,
								ShowLabel:       req.Tabs[0].Sections[0].Fields[0].ShowLabel,
								Attributes:      req.Tabs[0].Sections[0].Fields[0].Attributes,
								IsVisibleLayout: req.Tabs[0].Sections[0].Fields[0].IsVisibleLayout,
								TableId:         req.TableId,
							},
						},
					},
				},
			},
		},
	}

	return resp, nil
}

func (l layoutRepo) GetSingleLayout(ctx context.Context, req *nb.GetSingleLayoutRequest) (resp *nb.LayoutResponse, err error) {
	resp = &nb.LayoutResponse{}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if req.TableId == "" {
		tableQuery := `
            SELECT
                id
            FROM "table" WHERE "slug" = $1
        `
		var tableID string
		err := conn.QueryRow(ctx, tableQuery, req.TableSlug).Scan(&tableID)
		if err != nil {
			return resp, err
		}
		req.TableId = tableID
	}

	var menu_id sql.NullString

	layoutQuery := `
        SELECT
            id,
            label,
            "order",
            "type",
            icon,
            is_default,
            is_modal,
            is_visible_section,
            attributes,
            table_id,
            menu_id
        FROM layout
        WHERE table_id = $1 AND menu_id = $2
    `

	row := conn.QueryRow(ctx, layoutQuery, req.TableId, req.MenuId)
	err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.Attributes, &resp.TableId, menu_id.String)
	if err != nil {
		if err == pgx.ErrNoRows {
			layoutQuery = `
                SELECT
                    id,
                    label,
                    "order",
                    "type",
                    icon,
                    is_default,
                    is_modal,
                    is_visible_section,
                    attributes,
                    table_id,
                    menu_id
                FROM layout
                WHERE table_id = $1 AND is_default = true
            `

			row := conn.QueryRow(ctx, layoutQuery, req.TableId)
			err = row.Scan(&resp.Id, &resp.Label,
				&resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.Attributes, &resp.TableId, &menu_id)

			if err != nil {
				return resp, err
			}
		}
	}

	table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Id: req.TableId})
	if err != nil {
		return resp, err
	}
	req.TableId = table["id"].(string)
	req.TableSlug = table["slug"].(string)

	var layouts []*nb.LayoutResponse

	query := `SELECT
			id,
			label,
			"order",
			"type",
			icon,
			is_default,
			is_modal,
			is_visible_section,
			attributes,
			table_id,
			menu_id
		FROM layout
		WHERE table_id = $1 AND menu_id = $2`

	rows, err := conn.Query(ctx, query, req.TableId, req.MenuId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		layout := &nb.LayoutResponse{}
		err := rows.Scan(&layout.Id, &layout.Label, &layout.Order, &layout.Type, &layout.Icon, &layout.IsDefault, &layout.IsModal, &layout.IsVisibleSection, &layout.Attributes, &layout.TableId, &layout.MenuId)
		if err != nil {
			return nil, err
		}
		layouts = append(layouts, layout)
	}

	if err := rows.Err(); err != nil {
		return resp, err
	}

	var layoutIDs []string
	for _, layout := range layouts {
		layoutIDs = append(layoutIDs, layout.Id)

		summaryFields := make([]*nb.FieldResponse, 0)
		if layout.SummaryFields != nil && len(layout.SummaryFields) > 0 {
			for _, fieldReq := range layout.SummaryFields {
				field := &nb.FieldResponse{}

				if strings.Contains(fieldReq.Id, "#") {
					field.Id = fieldReq.Id
					field.Label = fieldReq.Label
					field.Order = fieldReq.Order
					field.RelationType = fieldReq.RelationType
					relationID := strings.Split(fieldReq.Id, "#")[1]
					var fieldResp *nb.Field
					err := conn.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
					if err != nil {
						if err != pgx.ErrNoRows {
							return nil, err
						}
					}

					var relation *nb.RelationForGetAll
					err = conn.QueryRow(ctx, "SELECT id FROM relation WHERE id = $1", relationID).Scan(&relation.Id)
					if err != nil {
						if err == pgx.ErrNoRows {
							continue
						}
						return nil, err
					}

					var viewOfRelation *nb.View
					err = conn.QueryRow(ctx, "SELECT view_fields FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, req.TableSlug).Scan(&viewOfRelation.ViewFields)
					var viewFieldIds []*nb.Field
					if viewOfRelation.Id != "" {
						for _, id := range viewOfRelation.ViewFields {
							viewFieldIds = append(viewFieldIds, &nb.Field{Id: id})
						}
					}
					viewFieldIds = relation.ViewFields
					if viewOfRelation != nil {
						if len(viewOfRelation.ViewFields) > 0 {
							viewFieldIds = make([]*nb.Field, len(viewOfRelation.ViewFields))
							for i, id := range viewOfRelation.ViewFields {
								viewFieldIds[i] = &nb.Field{Id: id}
							}
						}
					}

					if relation.Id != "" {
						for _, fieldID := range viewFieldIds {
							var field *nb.FieldResponse
							err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
							if err != nil {
								if err == pgx.ErrNoRows {
									continue
								}
								return nil, err
							}
							var fieldAsAttribute []string

							if req.LanguageSetting != "" && field.EnableMultilanguage {
								if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
									fieldAsAttribute = append(fieldAsAttribute, field.Slug)
								} else {
									continue
								}
							} else {
								fieldAsAttribute = append(fieldAsAttribute, field.Slug)
							}

							if viewOfRelation.Id != "" {
								field.IsEditable = viewOfRelation.IsEditable
							}
						}

						tableFields := []nb.Field{}
						rows, err = conn.Query(ctx, "SELECT id, auto_fill_table, auto_fill_field FROM field WHERE table_id = $1", &tableFields, req.TableId)
						if err != nil {
							return nil, err
						}
						defer rows.Close()

						autofillFields := []map[string]interface{}{}
						for i := range tableFields {
							field := &tableFields[i]
							autoFillTable := field.AutofillTable
							splitedAutoFillTable := []string{}
							if strings.Contains(field.AutofillTable, "#") {
								splitedAutoFillTable = strings.Split(field.AutofillTable, "#")
								autoFillTable = splitedAutoFillTable[0]
							}
							if field.AutofillField != "" && autoFillTable != "" && autoFillTable == strings.Split(fieldReq.Id, "#")[0] {
								autofill := map[string]interface{}{
									"field_from": field.AutofillField,
									"field_to":   field.Slug,
									"automatic":  field.Automatic,
								}
								if fieldResp.Slug == splitedAutoFillTable[1] {
									autofillFields = append(autofillFields, autofill)
								}
							}
						}

						originalAttributes := make(map[string]interface{})
						dynamicTables := []string{}
						if relation.Type == "Many2Dynamic" {
							for _, dynamicTable := range relation.DynamicTables {
								dynamicTableInfo, err := helper.TableVer(ctx, helper.TableVerReq{Slug: dynamicTable.TableSlug})
								if err != nil {
									return nil, err
								}
								viewFieldsOfDynamicRelation := dynamicTable.ViewFields
								viewOfDynamicRelation := &nb.View{}
								err = conn.QueryRow(ctx, "SELECT id, relation_id, relation_table_slug FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, dynamicTable.TableSlug).Scan(&viewOfDynamicRelation.Id, &viewOfDynamicRelation.RelationId, &viewOfDynamicRelation.RelationTableSlug)
								if err != nil {
									if err != pgx.ErrNoRows {
										return nil, err
									}
								}
								if err != nil {
									return nil, err
								}
								if viewOfDynamicRelation != nil && len(viewOfDynamicRelation.ViewFields) > 0 {
									viewFieldsOfDynamicRelation = viewOfDynamicRelation.ViewFields
								}

								dynamicTableToAttribute := make(map[string]interface{})
								viewFieldsInDynamicTable := []string{}
								for _, fieldID := range viewFieldsOfDynamicRelation {
									field := &nb.Field{}
									err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
									if err != nil {
										if err == pgx.ErrNoRows {
											continue
										}
										return nil, err
									}
									fieldAsAttribute := []string{}
									if req.LanguageSetting != "" && field.EnableMultilanguage {
										if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
											fieldAsAttribute = append(fieldAsAttribute, field.Slug)
										} else {
											continue
										}
									} else {
										fieldAsAttribute = append(fieldAsAttribute, field.Slug)
									}
									viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, fieldAsAttribute...)
								}
								dynamicTableToAttribute["table"] = dynamicTableInfo
								dynamicTableToAttribute["viewFields"] = viewFieldsInDynamicTable

								if field != nil {
									if field.Attributes != nil {
										attributesBytes, err := field.Attributes.MarshalJSON()
										if err != nil {
											return nil, err
										}
										err = json.Unmarshal(attributesBytes, &field)
										if err != nil {
											return nil, err
										}
									}

									if req.LanguageSetting != "" && field.EnableMultilanguage {
										if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
											viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
										} else {
											continue
										}
									} else {
										viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
									}
								}
								dynamicTableToAttribute["view_fields"] = viewFieldsInDynamicTable
								dynamicTables = append(dynamicTables, fmt.Sprintf("%v", dynamicTableToAttribute))
							}

							originalAttributes = make(map[string]interface{})

							originalAttributes["autofill"] = fieldResp.AutofillField
							originalAttributes["view_fields"] = relation.ViewFields
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
						} else {
							originalAttributes["autofill"] = fieldResp.AutofillField
							originalAttributes["view_fields"] = relation.ViewFields
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
							if viewOfRelation != nil {
								for k, v := range viewOfRelation.Attributes.AsMap() {
									originalAttributes[k] = v
								}
							}

						}
						if viewOfRelation != nil {
							if len(viewOfRelation.DefaultValues) > 0 {
								originalAttributes["default_values"] = viewOfRelation.DefaultValues
							}
							originalAttributes["creatable"] = viewOfRelation.Creatable
						}

						originalAttributesJSON, err := json.Marshal(originalAttributes)
						if err != nil {
							return nil, err
						}
						var encodedAttributes []byte
						err = json.Unmarshal(originalAttributesJSON, &encodedAttributes)
						if err != nil {
							return nil, err
						}
						var attributes structpb.Struct
						err = protojson.Unmarshal(encodedAttributes, &attributes)
						if err != nil {
							return nil, err
						}
						field.Attributes = &attributes
						summaryFields = append(summaryFields, field)
						if strings.Contains(fieldReq.Id, "@") {
							field.Id = fieldReq.Id
						} else {
							guid := fieldReq.Id
							fieldQuery := `
							SELECT
								id,
								table_id,
								required,
								"slug",
								"label",
								"default",
								"type",
								"index"

							FROM "field"
							WHERE id = $1
						`
							err := conn.QueryRow(ctx, fieldQuery, guid).Scan(
								&field.Id,
								&field.TableId,
								&field.Required,
								&field.Slug,
								&field.Label,
								&field.Default,
								&field.Type,
								&field.Index,
							)
							if err != nil {
								return nil, err
							}
							if field != nil {
								field.Order = fieldReq.Order
								field.Column = fieldReq.Column
								field.Id = fieldReq.Id
								field.RelationType = fieldReq.RelationType
								summaryFields = append(summaryFields, field)
							}
							summaryFields = append(summaryFields, field)
						}
					}
				}
			}
		}
		fieldsWithPermissions, err := helper.AddPermissionToField(ctx, conn, summaryFields, req.RoleId, req.TableSlug, req.ProjectId)
		if err != nil {
			return nil, err
		}
		layout.SummaryFields = fieldsWithPermissions
		tabs := make([]*nb.TabResponse, 0)
		rows, err := conn.Query(ctx, `SELECT id, "type", "relation_id" FROM "tab" WHERE layout_id = ANY($1)`, layoutIDs)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var tab nb.TabResponse
			err := rows.Scan(&tab.Id, &tab.Type, &tab.RelationId)
			if err != nil {
				return resp, err
			}
			tabs = append(tabs, &tab)
		}

		if rows.Err() != nil {
			return resp, rows.Err()
		}

		for _, tab := range tabs {
			if tab.Type == "section" {
				sections, err := (*sectionRepo).GetAll(&sectionRepo{}, ctx, &nb.GetAllSectionsRequest{
					ProjectId: req.ProjectId,
					RoleId:    req.RoleId,
					TableSlug: req.TableSlug,
					TableId:   req.TableId,
				})
				if err != nil {
					return resp, err
				}
				tab.Sections = sections.Sections
			} else if tab.Type == "relation" && tab.RelationId != "" {
				relation, err := (*relationRepo).GetSingleViewForRelation(&relationRepo{}, ctx, models.ReqForViewRelation{
					Id:        tab.RelationId,
					ProjectId: req.ProjectId,
					RoleId:    req.RoleId,
					TableSlug: req.TableSlug,
				})
				if err != nil {
					return resp, err
				}

				var newRelation nb.RelationForSection
				newRelation.Id = relation.Id
				newRelation.Type = relation.Type

				tab.Relation = &newRelation
			}

			sort.Slice(tabs, func(i, j int) bool {
				return tabs[i].Order < tabs[j].Order
			})

			layout.Tabs = tabs
			layout.Id = resp.Id
		}
	}

	return resp, nil
}

func (l layoutRepo) GetAll(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error) {
	resp = &nb.GetListLayoutResponse{}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if req.TableId == "" {
		table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Slug: req.TableSlug, Id: req.TableId})
		if err != nil {
			return nil, err
		}
		req.TableId = table["id"].(string)
		req.TableSlug = table["slug"].(string)
	}

	payload := make(map[string]interface{})
	payload["table_id"] = req.TableId
	if req.IsDefualt {
		payload["is_default"] = true
	}
	if req.MenuId != "" {
		payload["menu_id"] = req.MenuId
	}

	rows, err := conn.Query(ctx, `
		SELECT
			id,
			label,
			"order",
			"type",
			icon,
			is_default,
			is_modal,
			is_visible_section,
			attributes,
			table_id,
			menu_id
		FROM layout
		WHERE table_id = $1 AND menu_id = $2
		ORDER BY created_at DESC
	`, payload["table_id"], payload["menu_id"])

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	layouts := make([]*nb.LayoutResponse, 0)
	for rows.Next() {
		layout := nb.LayoutResponse{}
		err := rows.Scan(&layout.Id, &layout.Label, &layout.Order, &layout.Type,
			&layout.Icon, &layout.IsDefault, &layout.IsModal,
			&layout.IsVisibleSection, &layout.Attributes, &layout.TableId, &layout.MenuId)
		if err != nil {
			return nil, err
		}
		layouts = append(layouts, &layout)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var layoutIDs []string
	for _, layout := range layouts {
		summaryFields := make([]*nb.FieldResponse, 0)
		if layout.SummaryFields == nil {
			layout.SummaryFields = []*nb.FieldResponse{}
		}
		if layout.SummaryFields != nil && len(layout.SummaryFields) > 0 {
			for _, fieldReq := range layout.SummaryFields {
				field := &nb.FieldResponse{}

				if strings.Contains(fieldReq.Id, "#") {
					field.Id = fieldReq.Id
					field.Label = fieldReq.Label
					field.Order = fieldReq.Order
					field.RelationType = fieldReq.RelationType
					relationID := strings.Split(fieldReq.Id, "#")[1]
					var fieldResp *nb.Field
					err := conn.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
					if err != nil {
						if err != pgx.ErrNoRows {
							return nil, err
						}
					}

					var relation *nb.RelationForGetAll
					err = conn.QueryRow(ctx, "SELECT id FROM relation WHERE id = $1", relationID).Scan(&relation.Id)
					if err != nil {
						if err == pgx.ErrNoRows {
							continue
						}
						return nil, err
					}

					var viewOfRelation *nb.View
					err = conn.QueryRow(ctx, "SELECT view_fields FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, req.TableSlug).Scan(&viewOfRelation.ViewFields)
					var viewFieldIds []*nb.Field
					if viewOfRelation.Id != "" {
						for _, id := range viewOfRelation.ViewFields {
							viewFieldIds = append(viewFieldIds, &nb.Field{Id: id})
						}
					}
					viewFieldIds = relation.ViewFields
					if viewOfRelation != nil {
						if len(viewOfRelation.ViewFields) > 0 {
							viewFieldIds = make([]*nb.Field, len(viewOfRelation.ViewFields))
							for i, id := range viewOfRelation.ViewFields {
								viewFieldIds[i] = &nb.Field{Id: id}
							}
						}
					}

					if relation.Id != "" {
						for _, fieldID := range viewFieldIds {
							var field *nb.FieldResponse
							err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
							if err != nil {
								if err == pgx.ErrNoRows {
									continue
								}
								return nil, err
							}
							var fieldAsAttribute []string

							if req.LanguageSetting != "" && field.EnableMultilanguage {
								if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
									fieldAsAttribute = append(fieldAsAttribute, field.Slug)
								} else {
									continue
								}
							} else {
								fieldAsAttribute = append(fieldAsAttribute, field.Slug)
							}

							if viewOfRelation.Id != "" {
								field.IsEditable = viewOfRelation.IsEditable
							}
						}

						tableFields := []nb.Field{}
						rows, err = conn.Query(ctx, "SELECT id, auto_fill_table, auto_fill_field FROM field WHERE table_id = $1", &tableFields, req.TableId)
						if err != nil {
							return nil, err
						}
						defer rows.Close()

						autofillFields := []map[string]interface{}{}
						for i := range tableFields {
							field := &tableFields[i]
							autoFillTable := field.AutofillTable
							splitedAutoFillTable := []string{}
							if strings.Contains(field.AutofillTable, "#") {
								splitedAutoFillTable = strings.Split(field.AutofillTable, "#")
								autoFillTable = splitedAutoFillTable[0]
							}
							if field.AutofillField != "" && autoFillTable != "" && autoFillTable == strings.Split(fieldReq.Id, "#")[0] {
								autofill := map[string]interface{}{
									"field_from": field.AutofillField,
									"field_to":   field.Slug,
									"automatic":  field.Automatic,
								}
								if fieldResp.Slug == splitedAutoFillTable[1] {
									autofillFields = append(autofillFields, autofill)
								}
							}
						}

						originalAttributes := make(map[string]interface{})
						dynamicTables := []string{}
						if relation.Type == "Many2Dynamic" {
							for _, dynamicTable := range relation.DynamicTables {
								dynamicTableInfo, err := helper.TableVer(ctx, helper.TableVerReq{Slug: dynamicTable.TableSlug})
								if err != nil {
									return nil, err
								}
								viewFieldsOfDynamicRelation := dynamicTable.ViewFields
								viewOfDynamicRelation := &nb.View{}
								err = conn.QueryRow(ctx, "SELECT id, relation_id, relation_table_slug FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, dynamicTable.TableSlug).Scan(&viewOfDynamicRelation.Id, &viewOfDynamicRelation.RelationId, &viewOfDynamicRelation.RelationTableSlug)
								if err != nil {
									if err != pgx.ErrNoRows {
										return nil, err
									}
								}
								if err != nil {
									return nil, err
								}
								if viewOfDynamicRelation != nil && len(viewOfDynamicRelation.ViewFields) > 0 {
									viewFieldsOfDynamicRelation = viewOfDynamicRelation.ViewFields
								}

								dynamicTableToAttribute := make(map[string]interface{})
								viewFieldsInDynamicTable := []string{}
								for _, fieldID := range viewFieldsOfDynamicRelation {
									field := &nb.Field{}
									err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
									if err != nil {
										if err == pgx.ErrNoRows {
											continue
										}
										return nil, err
									}
									fieldAsAttribute := []string{}
									if req.LanguageSetting != "" && field.EnableMultilanguage {
										if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
											fieldAsAttribute = append(fieldAsAttribute, field.Slug)
										} else {
											continue
										}
									} else {
										fieldAsAttribute = append(fieldAsAttribute, field.Slug)
									}
									viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, fieldAsAttribute...)
								}
								dynamicTableToAttribute["table"] = dynamicTableInfo
								dynamicTableToAttribute["viewFields"] = viewFieldsInDynamicTable

								if field != nil {
									if field.Attributes != nil {
										attributesBytes, err := field.Attributes.MarshalJSON()
										if err != nil {
											return nil, err
										}
										err = json.Unmarshal(attributesBytes, &field)
										if err != nil {
											return nil, err
										}
									}

									if req.LanguageSetting != "" && field.EnableMultilanguage {
										if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
											viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
										} else {
											continue
										}
									} else {
										viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
									}
								}
								dynamicTableToAttribute["view_fields"] = viewFieldsInDynamicTable
								dynamicTables = append(dynamicTables, fmt.Sprintf("%v", dynamicTableToAttribute))
							}

							originalAttributes = make(map[string]interface{})

							originalAttributes["autofill"] = fieldResp.AutofillField
							originalAttributes["view_fields"] = relation.ViewFields
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
						} else {
							originalAttributes["autofill"] = fieldResp.AutofillField
							originalAttributes["view_fields"] = relation.ViewFields
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath
							if viewOfRelation != nil {
								for k, v := range viewOfRelation.Attributes.AsMap() {
									originalAttributes[k] = v
								}
							}

						}
						if viewOfRelation != nil {
							if len(viewOfRelation.DefaultValues) > 0 {
								originalAttributes["default_values"] = viewOfRelation.DefaultValues
							}
							originalAttributes["creatable"] = viewOfRelation.Creatable
						}

						originalAttributesJSON, err := json.Marshal(originalAttributes)
						if err != nil {
							return nil, err
						}
						var encodedAttributes []byte
						err = json.Unmarshal(originalAttributesJSON, &encodedAttributes)
						if err != nil {
							return nil, err
						}
						var attributes structpb.Struct
						err = protojson.Unmarshal(encodedAttributes, &attributes)
						if err != nil {
							return nil, err
						}
						field.Attributes = &attributes
						summaryFields = append(summaryFields, field)
						if strings.Contains(fieldReq.Id, "@") {
							field.Id = fieldReq.Id
						} else {
							guid := fieldReq.Id
							fieldQuery := `
							SELECT
								id,
								table_id,
								required,
								"slug",
								"label",
								"default",
								"type",
								"index"

							FROM "field"
							WHERE id = $1
						`
							err := conn.QueryRow(ctx, fieldQuery, guid).Scan(
								&field.Id,
								&field.TableId,
								&field.Required,
								&field.Slug,
								&field.Label,
								&field.Default,
								&field.Type,
								&field.Index,
							)
							if err != nil {
								return nil, err
							}
							if field != nil {
								field.Order = fieldReq.Order
								field.Column = fieldReq.Column
								field.Id = fieldReq.Id
								field.RelationType = fieldReq.RelationType
								summaryFields = append(summaryFields, field)
							}
							summaryFields = append(summaryFields, field)
						}
					}
				}
			}
		}
		layoutIDs = append(layoutIDs, layout.Id)

		tabs := make([]*nb.TabResponse, 0)
		sqlQuery := `
			SELECT
				"id",
				"type",
				"relation_id"
			FROM "tab"
			WHERE "layout_id"::VARCHAR IN ($1)
		`
		rows, err := conn.Query(ctx, sqlQuery, pq.Array(layoutIDs))
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var tab nb.TabResponse
			err := rows.Scan(&tab.Id, &tab.Type, &tab.RelationId)
			if err != nil {
				return nil, err
			}
			tabs = append(tabs, &tab)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		layout.Tabs = tabs

		for _, tab := range tabs {
			if tab.Type == "section" {

				sections, err := (*sectionRepo).GetAll(&sectionRepo{}, ctx, &nb.GetAllSectionsRequest{
					ProjectId: req.ProjectId,
					RoleId:    req.RoleId,
					TableSlug: req.TableSlug,
					TableId:   req.TableId,
				})
				if err != nil {
					return nil, err
				}
				tab.Sections = sections.Sections
			} else if tab.Type == "relation" && tab.RelationId != "" {
				relation, err := (*relationRepo).GetSingleViewForRelation(&relationRepo{}, ctx, models.ReqForViewRelation{
					Id:        tab.RelationId,
					ProjectId: req.ProjectId,
					RoleId:    req.RoleId,
					TableSlug: req.TableSlug,
				})
				if err != nil {
					return nil, err
				}

				var newRelation nb.RelationForSection
				newRelation.Id = relation.Id
				newRelation.Type = relation.Type

				tab.Relation = &newRelation
			}

			sort.Slice(tabs, func(i, j int) bool {
				return tabs[i].Order < tabs[j].Order
			})
		}
		layout.Tabs = tabs
	}

	resp.Layouts = layouts

	return resp, nil
}

func (l layoutRepo) RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) error {

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return err
	}
	defer conn.Close()

	tx, err := l.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM section WHERE tab_id IN (SELECT id FROM tab WHERE layout_id = $1)", req.Id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "DELETE FROM tab WHERE layout_id = $1", req.Id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	_, err = tx.Exec(ctx, "DELETE FROM layout WHERE id = $1", req.Id)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return nil

}

func (l layoutRepo) GetByID(ctx context.Context, req *nb.LayoutPrimaryKey) (resp *nb.LayoutResponse, err error) {
	resp = &nb.LayoutResponse{}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	layoutQuery := `
        SELECT
            id,
            label,
            "order",
            "type",
            icon,
            is_default,
            is_modal,
            is_visible_section,
            attributes,
            table_id,
            menu_id
        FROM layout
        WHERE id = $1
    `

	var menu_id sql.NullString
	row := conn.QueryRow(ctx, layoutQuery, req.Id)
	err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.Attributes, &resp.TableId, &menu_id)
	if err != nil {
		return nil, err
	}

	tableQuery := `
            SELECT
                id,
				slug
            FROM "table" WHERE "id" = $1
        `
	var tableID string
	var tableSlug string
	err = conn.QueryRow(ctx, tableQuery, resp.TableId).Scan(&tableID, &tableSlug)
	if err != nil {
		return resp, err
	}

	// summaryFields := make([]*nb.FieldResponse, 0)
	// if resp.SummaryFields != nil && len(resp.SummaryFields) > 0 {
	// 	for _, fieldReq := range resp.SummaryFields {
	// 		field := &nb.FieldResponse{}

	// 		if strings.Contains(fieldReq.Id, "#") {
	// 			field.Id = fieldReq.Id
	// 			field.Label = fieldReq.Label
	// 			field.Order = fieldReq.Order
	// 			field.RelationType = fieldReq.RelationType
	// 			relationID := strings.Split(fieldReq.Id, "#")[1]
	// 			var fieldResp *nb.Field
	// 			err := conn.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
	// 			if err != nil {
	// 				if err != pgx.ErrNoRows {
	// 					return nil, err
	// 				}
	// 			}

	// 			var relation *nb.RelationForGetAll
	// 			err = conn.QueryRow(ctx, "SELECT id FROM relation WHERE id = $1", relationID).Scan(&relation.Id)
	// 			if err != nil {
	// 				if err == pgx.ErrNoRows {
	// 					continue
	// 				}
	// 				return nil, err
	// 			}

	// 			var viewOfRelation *nb.View
	// 			err = conn.QueryRow(ctx, "SELECT view_fields FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, req.TableSlug).Scan(&viewOfRelation.ViewFields)
	// 			var viewFieldIds []*nb.Field
	// 			if viewOfRelation.Id != "" {
	// 				for _, id := range viewOfRelation.ViewFields {
	// 					viewFieldIds = append(viewFieldIds, &nb.Field{Id: id})
	// 				}
	// 			}
	// 			viewFieldIds = relation.ViewFields
	// 			if viewOfRelation != nil {
	// 				if len(viewOfRelation.ViewFields) > 0 {
	// 					viewFieldIds = make([]*nb.Field, len(viewOfRelation.ViewFields))
	// 					for i, id := range viewOfRelation.ViewFields {
	// 						viewFieldIds[i] = &nb.Field{Id: id}
	// 					}
	// 				}
	// 			}

	// 			if relation.Id != "" {
	// 				for _, fieldID := range viewFieldIds {
	// 					var field *nb.FieldResponse
	// 					err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
	// 					if err != nil {
	// 						if err == pgx.ErrNoRows {
	// 							continue
	// 						}
	// 						return nil, err
	// 					}
	// 					var fieldAsAttribute []string

	// 					if req.LanguageSetting != "" && field.EnableMultilanguage {
	// 						if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
	// 							fieldAsAttribute = append(fieldAsAttribute, field.Slug)
	// 						} else {
	// 							continue
	// 						}
	// 					} else {
	// 						fieldAsAttribute = append(fieldAsAttribute, field.Slug)
	// 					}

	// 					if viewOfRelation.Id != "" {
	// 						field.IsEditable = viewOfRelation.IsEditable
	// 					}
	// 				}

	// 				tableFields := []nb.Field{}
	// 				rows, err = conn.Query(ctx, "SELECT id, auto_fill_table, auto_fill_field FROM field WHERE table_id = $1", &tableFields, req.TableId)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				defer rows.Close()

	// 				autofillFields := []map[string]interface{}{}
	// 				for i := range tableFields {
	// 					field := &tableFields[i]
	// 					autoFillTable := field.AutofillTable
	// 					splitedAutoFillTable := []string{}
	// 					if strings.Contains(field.AutofillTable, "#") {
	// 						splitedAutoFillTable = strings.Split(field.AutofillTable, "#")
	// 						autoFillTable = splitedAutoFillTable[0]
	// 					}
	// 					if field.AutofillField != "" && autoFillTable != "" && autoFillTable == strings.Split(fieldReq.Id, "#")[0] {
	// 						autofill := map[string]interface{}{
	// 							"field_from": field.AutofillField,
	// 							"field_to":   field.Slug,
	// 							"automatic":  field.Automatic,
	// 						}
	// 						if fieldResp.Slug == splitedAutoFillTable[1] {
	// 							autofillFields = append(autofillFields, autofill)
	// 						}
	// 					}
	// 				}

	// 				originalAttributes := make(map[string]interface{})
	// 				dynamicTables := []string{}
	// 				if relation.Type == "Many2Dynamic" {
	// 					for _, dynamicTable := range relation.DynamicTables {
	// 						dynamicTableInfo, err := helper.TableVer(ctx, helper.TableVerReq{Slug: dynamicTable.TableSlug})
	// 						if err != nil {
	// 							return nil, err
	// 						}
	// 						viewFieldsOfDynamicRelation := dynamicTable.ViewFields
	// 						viewOfDynamicRelation := &nb.View{}
	// 						err = conn.QueryRow(ctx, "SELECT id, relation_id, relation_table_slug FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, dynamicTable.TableSlug).Scan(&viewOfDynamicRelation.Id, &viewOfDynamicRelation.RelationId, &viewOfDynamicRelation.RelationTableSlug)
	// 						if err != nil {
	// 							if err != pgx.ErrNoRows {
	// 								return nil, err
	// 							}
	// 						}
	// 						if err != nil {
	// 							return nil, err
	// 						}
	// 						if viewOfDynamicRelation != nil && len(viewOfDynamicRelation.ViewFields) > 0 {
	// 							viewFieldsOfDynamicRelation = viewOfDynamicRelation.ViewFields
	// 						}

	// 						dynamicTableToAttribute := make(map[string]interface{})
	// 						viewFieldsInDynamicTable := []string{}
	// 						for _, fieldID := range viewFieldsOfDynamicRelation {
	// 							field := &nb.Field{}
	// 							err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
	// 							if err != nil {
	// 								if err == pgx.ErrNoRows {
	// 									continue
	// 								}
	// 								return nil, err
	// 							}
	// 							fieldAsAttribute := []string{}
	// 							if req.LanguageSetting != "" && field.EnableMultilanguage {
	// 								if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
	// 									fieldAsAttribute = append(fieldAsAttribute, field.Slug)
	// 								} else {
	// 									continue
	// 								}
	// 							} else {
	// 								fieldAsAttribute = append(fieldAsAttribute, field.Slug)
	// 							}
	// 							viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, fieldAsAttribute...)
	// 						}
	// 						dynamicTableToAttribute["table"] = dynamicTableInfo
	// 						dynamicTableToAttribute["viewFields"] = viewFieldsInDynamicTable

	// 						if field != nil {
	// 							if field.Attributes != nil {
	// 								attributesBytes, err := field.Attributes.MarshalJSON()
	// 								if err != nil {
	// 									return nil, err
	// 								}
	// 								err = json.Unmarshal(attributesBytes, &field)
	// 								if err != nil {
	// 									return nil, err
	// 								}
	// 							}

	// 							if req.LanguageSetting != "" && field.EnableMultilanguage {
	// 								if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
	// 									viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
	// 								} else {
	// 									continue
	// 								}
	// 							} else {
	// 								viewFieldsInDynamicTable = append(viewFieldsInDynamicTable, field.Slug)
	// 							}
	// 						}
	// 						dynamicTableToAttribute["view_fields"] = viewFieldsInDynamicTable
	// 						dynamicTables = append(dynamicTables, fmt.Sprintf("%v", dynamicTableToAttribute))
	// 					}

	// 					originalAttributes = make(map[string]interface{})

	// 					originalAttributes["autofill"] = fieldResp.AutofillField
	// 					originalAttributes["view_fields"] = relation.ViewFields
	// 					originalAttributes["auto_filters"] = relation.AutoFilters
	// 					originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
	// 					originalAttributes["dynamic_tables"] = dynamicTables
	// 					originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
	// 					originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
	// 					originalAttributes["cascadings"] = relation.Cascadings
	// 					originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
	// 					originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
	// 					originalAttributes["function_path"] = viewOfRelation.FunctionPath
	// 				} else {
	// 					originalAttributes["autofill"] = fieldResp.AutofillField
	// 					originalAttributes["view_fields"] = relation.ViewFields
	// 					originalAttributes["auto_filters"] = relation.AutoFilters
	// 					originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
	// 					originalAttributes["dynamic_tables"] = dynamicTables
	// 					originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
	// 					originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
	// 					originalAttributes["cascadings"] = relation.Cascadings
	// 					originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
	// 					originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
	// 					originalAttributes["function_path"] = viewOfRelation.FunctionPath
	// 					if viewOfRelation != nil {
	// 						for k, v := range viewOfRelation.Attributes.AsMap() {
	// 							originalAttributes[k] = v
	// 						}
	// 					}

	// 				}
	// 				if viewOfRelation != nil {
	// 					if len(viewOfRelation.DefaultValues) > 0 {
	// 						originalAttributes["default_values"] = viewOfRelation.DefaultValues
	// 					}
	// 					originalAttributes["creatable"] = viewOfRelation.Creatable
	// 				}

	// 				originalAttributesJSON, err := json.Marshal(originalAttributes)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				var encodedAttributes []byte
	// 				err = json.Unmarshal(originalAttributesJSON, &encodedAttributes)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				var attributes structpb.Struct
	// 				err = protojson.Unmarshal(encodedAttributes, &attributes)
	// 				if err != nil {
	// 					return nil, err
	// 				}
	// 				field.Attributes = &attributes
	// 				summaryFields = append(summaryFields, field)
	// 				if strings.Contains(fieldReq.Id, "@") {
	// 					field.Id = fieldReq.Id
	// 				} else {
	// 					guid := fieldReq.Id
	// 					fieldQuery := `
	// 						SELECT
	// 							id,
	// 							table_id,
	// 							required,
	// 							"slug",
	// 							"label",
	// 							"default",
	// 							"type",
	// 							"index"

	// 						FROM "field"
	// 						WHERE id = $1
	// 					`
	// 					err := conn.QueryRow(ctx, fieldQuery, guid).Scan(
	// 						&field.Id,
	// 						&field.TableId,
	// 						&field.Required,
	// 						&field.Slug,
	// 						&field.Label,
	// 						&field.Default,
	// 						&field.Type,
	// 						&field.Index,
	// 					)
	// 					if err != nil {
	// 						return nil, err
	// 					}
	// 					if field != nil {
	// 						field.Order = fieldReq.Order
	// 						field.Column = fieldReq.Column
	// 						field.Id = fieldReq.Id
	// 						field.RelationType = fieldReq.RelationType
	// 						summaryFields = append(summaryFields, field)
	// 					}
	// 					summaryFields = append(summaryFields, field)
	// 				}
	// 			}
	// 		}
	// 	}

	// 	fieldsWithPermissions, err := helper.AddPermissionToField(ctx, conn, summaryFields, req.RoleId, req.TableSlug, req.ProjectId)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	resp.SummaryFields = fieldsWithPermissions
	// }
	tabs := make([]*nb.TabResponse, 0)
	var (
		id          string
		typeNull    sql.NullString
		relation_id sql.NullString
	)
	err = conn.QueryRow(ctx, `SELECT id, "type", "relation_id" FROM "tab" WHERE layout_id = $1`, req.Id).Scan(&id, &typeNull, &relation_id)
	if err != nil {
		return nil, err
	}
	tab := &nb.TabResponse{
		Id:         id,
		Type:       typeNull.String,
		RelationId: relation_id.String,
	}
	tabs = append(tabs, tab)

	for _, tab := range tabs {
		if tab.Type == "section" {
			sections, err := (*sectionRepo).GetAll(&sectionRepo{}, ctx, &nb.GetAllSectionsRequest{
				ProjectId: req.ProjectId,
				RoleId:    req.RoleId,
				TableSlug: tableSlug,
				TableId:   tableID,
				TabId:     id,
			})
			if err != nil {
				return resp, err
			}
			tab.Sections = sections.Sections
		} else if tab.Type == "relation" && tab.RelationId != "" {
			relation, err := (*relationRepo).GetSingleViewForRelation(&relationRepo{}, ctx, models.ReqForViewRelation{
				Id:        tab.RelationId,
				ProjectId: req.ProjectId,
				RoleId:    req.RoleId,
				TableSlug: tableSlug,
			})
			if err != nil {
				return resp, err
			}

			var newRelation nb.RelationForSection
			newRelation.Id = relation.Id
			newRelation.Type = relation.Type

			tab.Relation = &newRelation
		}

		sort.Slice(tabs, func(i, j int) bool {
			return tabs[i].Order < tabs[j].Order
		})

		resp.Tabs = tabs

	}

	return resp, nil
}
