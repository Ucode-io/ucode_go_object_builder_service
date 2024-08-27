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
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
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

	conn := psqlpool.Get(req.GetProjectId())

	// tx, err := conn.Begin(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// defer func() {
	// 	if err != nil {
	// tx.Rollback()
	// 		return
	// 	}
	// 	err = tx.Commit(ctx)
	// }()

	var roleGUID string
	rows, err := conn.Query(ctx, "SELECT guid FROM role")
	if err != nil {
		return nil, fmt.Errorf("error fetching roles: %w", err)
	}
	defer rows.Close()

	roles := []string{}
	for rows.Next() {
		if err := rows.Scan(&roleGUID); err != nil {
			return nil, fmt.Errorf("error scanning role GUID: %w", err)
		}
		roles = append(roles, roleGUID)
	}

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
	attributesJSON, err := json.Marshal(req.Attributes)
	if err != nil {
		return nil, fmt.Errorf("error marshaling attributes to JSON: %w", err)
	}

	query := `
        INSERT INTO "layout" (
            "id", "label", "order", "type", "icon", "is_default", 
            "is_modal", "is_visible_section",
             "table_id", "menu_id", "attributes"
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
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
            "menu_id" = EXCLUDED.menu_id,
			"attributes" = EXCLUDED.attributes
    `
	_, err = conn.Exec(ctx, query,
		layoutId, req.Label, req.Order, req.Type, req.Icon,
		req.IsDefault, req.IsModal, req.IsVisibleSection,
		req.TableId, req.MenuId, attributesJSON)
	if err != nil {
		return nil, fmt.Errorf("error inserting layout: %w", err)
	}

	if req.IsDefault {
		_, err = conn.Exec(ctx, `
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
		tab_ids                       []string
	)

	rows, err = conn.Query(ctx, "SELECT id FROM tab WHERE layout_id = $1", layoutId)

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
		tab_ids = append(tab_ids, tabId)
	}

	rows, err = conn.Query(ctx, "SELECT id FROM section WHERE tab_id = ANY($1)", pq.Array(tab_ids))
	if err != nil {
		return nil, fmt.Errorf("error fetching sections: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sectionId string
		if err := rows.Scan(&sectionId); err != nil {
			return nil, fmt.Errorf("error scanning section ID: %w", err)
		}
		mapSections[sectionId] = 1
	}

	for i := 0; i < len(req.Tabs); i++ {
		tab := req.Tabs[i]
		if tab.Id == "" {
			tab.Id = uuid.New().String()
		}
		if tab.Type == "relation" {
			relationIds = append(relationIds, tab.RelationId)
		}

		if _, ok := mapTabs[tab.Id]; ok {
			mapTabs[tab.Id] = 2
		}
		attributesJSON, err := json.Marshal(tab.Attributes)
		if err != nil {
			return nil, fmt.Errorf("error marshaling attributes to JSON: %w", err)
		}

		query := ""

		if tab.RelationId != "" {

			// atr, err := helper.ConvertStructToMap(tab.Attributes)
			// if err != nil {
			// 	return nil, err
			// }

			// tab.Label = cast.ToString(atr["label_to_en"])

			query = fmt.Sprintf(`
			INSERT INTO "tab" (
				"id", "label", "layout_id",  "type",
				"order", "icon", relation_id, "attributes"
			) VALUES ('%s', '%s', '%s', '%s', %d, '%s', '%s', '%s')
			ON CONFLICT (id) DO UPDATE
			SET
				"label" = EXCLUDED.label,
				"layout_id" = EXCLUDED.layout_id,
				"type" = EXCLUDED.type,
				"order" = EXCLUDED.order,
				"icon" = EXCLUDED.icon,
				"relation_id" = EXCLUDED.relation_id,
				"attributes" = EXCLUDED.attributes
			`,
				tab.Id, tab.Label, layoutId, tab.Type, i+1, tab.Icon, tab.RelationId, string(attributesJSON))
		} else {
			query = fmt.Sprintf(`
			INSERT INTO "tab" (
				"id", "label", "layout_id",  "type",
				"order", "icon", "attributes"
			) VALUES ('%s', '%s', '%s', '%s', %d, '%s', '%s')
			ON CONFLICT (id) DO UPDATE
			SET
				"label" = EXCLUDED.label,
				"layout_id" = EXCLUDED.layout_id,
				"type" = EXCLUDED.type,
				"order" = EXCLUDED.order,
				"icon" = EXCLUDED.icon,
				"attributes" = EXCLUDED.attributes
			`,
				tab.Id, tab.Label, layoutId, tab.Type, i+1, tab.Icon, string(attributesJSON))
		}

		bulkWriteTab = append(bulkWriteTab, query)

		for i, section := range tab.Sections {
			if section.Id == "" {
				section.Id = uuid.New().String()
			}

			if _, ok := mapSections[section.Id]; ok {
				mapSections[section.Id] = 2
			}

			jsonFields := []byte(`[]`)

			if section.Fields != nil {
				jsonFields, err = json.Marshal(section.Fields)
				if err != nil {
					return nil, fmt.Errorf("error marhaling section.Fields to JSON: %w", err)
				}
			}

			attributes := []byte(`{}`)

			if section.Attributes != nil {
				attributes, err = json.Marshal(section.Attributes)
				if err != nil {
					return nil, fmt.Errorf("error marshaling section attributes to JSON: %w", err)
				}
			}

			bulkWriteSection = append(bulkWriteSection, fmt.Sprintf(`
			INSERT INTO "section" (
				"id", "tab_id", "label", "order", "icon", 
				"column",  is_summary_section, "fields", "table_id", "attributes"
			) VALUES ('%s', '%s', '%s', %d, '%s', '%s', '%t', '%s', '%s', '%s')
			ON CONFLICT (id) DO UPDATE
			SET
				"tab_id" = EXCLUDED.tab_id,
				"label" = EXCLUDED.label,
				"order" = EXCLUDED.order,
				"icon" = EXCLUDED.icon,
				"column" = EXCLUDED.column,
				"is_summary_section" = EXCLUDED.is_summary_section,
				"fields" = EXCLUDED.fields,
				"table_id" = EXCLUDED.table_id,
				"attributes" = EXCLUDED.attributes

		
			`, section.Id, tab.Id, section.Label, i, section.Icon, section.Column, section.IsSummarySection, jsonFields, req.TableId, attributes))
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

	for _, roleId := range roles {
		for _, relationID := range relationIds {
			var exists int
			query := `
					SELECT COUNT(*)
					FROM view_relation_permission
					WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
				`

			err := conn.QueryRow(ctx, query, roleId, req.TableId, relationID).Scan(&exists)
			if err != nil {
				return nil, fmt.Errorf("error checking relation permission existence: %w", err)
			}

			if exists == 0 {
				insertManyRelationPermissions = append(insertManyRelationPermissions, fmt.Sprintf(`
						INSERT INTO view_relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
						VALUES ('%s', '%s', '%s', true, true, true, true)
					`, roleId, req.TableId, relationID))

			}

			if len(insertManyRelationPermissions) > 0 {
				for _, query := range insertManyRelationPermissions {
					_, err := conn.Exec(ctx, query)
					if err != nil {
						return nil, fmt.Errorf("error inserting relation permissions: %w", err)
					}
				}
			}

		}
	}

	if len(insertManyRelationPermissions) > 0 {
		for _, role := range roles {
			for _, relationID := range relationIds {

				var relationPermission bool
				err := conn.QueryRow(ctx, `
				SELECT EXISTS (
					SELECT 1 
					FROM view_relation_permission 
					WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
				)`, role, tableSlug1, relationID).Scan(&relationPermission)
				if err != nil {
					return nil, err
				}
				if !relationPermission {
					_, err := conn.Exec(ctx, `
					INSERT INTO view_relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
					VALUES ($1, $2, $3, true, true, true, true)`, roleGUID, tableSlug1, relationID)
					if err != nil {
						return nil, err
					}
				}
			}
		}
	}

	if len(deletedTabIds) > 0 {
		_, err := conn.Exec(ctx, "DELETE FROM tab WHERE id = ANY($1)", pq.Array(deletedTabIds))
		if err != nil {
			return nil, fmt.Errorf("error deleting tabs: %w", err)
		}
	}

	if len(deletedSectionIds) > 0 {
		_, err := conn.Exec(ctx, "DELETE FROM section WHERE id = ANY($1)", pq.Array(deletedSectionIds))
		if err != nil {
			return nil, err
		}
	}

	if len(bulkWriteTab) > 0 {
		for _, query := range bulkWriteTab {
			_, err := conn.Exec(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("error executing bulkWriteTab query: %w", err)
			}

		}
	}

	if len(bulkWriteSection) > 0 {
		for _, query := range bulkWriteSection {
			_, err := conn.Exec(ctx, query)
			if err != nil {
				return nil, fmt.Errorf("error executing bulkWriteSection query: %w", err)
			}
		}

	}
	return l.GetByID(ctx, &nb.LayoutPrimaryKey{Id: layoutId, ProjectId: req.ProjectId})
}

func (l *layoutRepo) GetSingleLayout(ctx context.Context, req *nb.GetSingleLayoutRequest) (resp *nb.LayoutResponse, err error) {
	resp = &nb.LayoutResponse{}
	conn := psqlpool.Get(req.GetProjectId())

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

	var menuID sql.NullString

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
	err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.Attributes, &resp.TableId, &menuID)
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
			err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.Attributes, &resp.TableId, &menuID)

			if err != nil {
				return resp, err
			}
		}
	}

	resp = &nb.LayoutResponse{
		Id:        resp.Id,
		Label:     resp.Label,
		Order:     resp.Order,
		TableId:   resp.TableId,
		Type:      resp.Type,
		IsDefault: resp.IsDefault,
		IsModal:   resp.IsModal,
		// Attributes: &structpb.Struct{
		// 	Fields: resp.Attributes.Fields,
		// },
		Icon:             resp.Icon,
		IsVisibleSection: resp.IsVisibleSection,
		MenuId:           menuID.String,
	}
	resp.Tabs = []*nb.TabResponse{}

	table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Id: req.TableId})
	if err != nil {
		return resp, err
	}
	req.TableId = table["id"].(string)
	req.TableSlug = table["slug"].(string)

	summaryFields := make([]*nb.FieldResponse, 0)

	fieldsWithPermissions, err := helper.AddPermissionToField(ctx, conn, summaryFields, req.RoleId, req.TableSlug, req.ProjectId)
	if err != nil {
		return nil, err
	}
	resp.SummaryFields = fieldsWithPermissions

	var (
		label      sql.NullString
		order      sql.NullInt32
		icon       sql.NullString
		relationID sql.NullString
	)

	tabs := make([]*nb.TabResponse, 0)
	sqlQuery := `
		SELECT
			"id",
			"type",
			"order",
			"label",
			"icon",
			"layout_id",
			"relation_id",
			"attributes"
		FROM "tab"
		WHERE "layout_id" = $1;
	`

	rows, err := conn.Query(ctx, sqlQuery, resp.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tab nb.TabResponse
		err := rows.Scan(&tab.Id, &tab.Type, &order, &label, &icon, &tab.LayoutId, &relationID, &tab.Attributes)
		if err != nil {
			return nil, err
		}
		if label.Valid {
			tab.Label = label.String
		}
		if order.Valid {
			tab.Order = order.Int32
		}
		if icon.Valid {
			tab.Icon = icon.String
		}
		if relationID.Valid {
			tab.RelationId = relationID.String
		}
		tabs = append(tabs, &tab)
	}

	sectionRepo := NewSectionRepo(conn)
	relationRepo := NewRelationRepo(conn)

	for _, tab := range tabs {
		if tab.Type == "section" {
			sections, err := sectionRepo.GetAll(ctx, &nb.GetAllSectionsRequest{
				ProjectId: req.ProjectId,
				RoleId:    req.RoleId,
				TableSlug: req.TableSlug,
				TableId:   req.TableId,
				TabId:     tab.Id,
			})
			if err != nil {
				return resp, err
			}
			tab.Sections = sections.Sections
		} else if tab.Type == "relation" && tab.RelationId != "" {
			relation, err := relationRepo.GetSingleViewForRelation(ctx, models.ReqForViewRelation{
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
	}

	sort.Slice(tabs, func(i, j int) bool {
		return tabs[i].Order < tabs[j].Order
	})
	resp.Tabs = tabs

	return resp, nil
}

func (l *layoutRepo) GetAll(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error) {
	resp = &nb.GetListLayoutResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	if req.TableId == "" {
		table, err := helper.TableVer(ctx, helper.TableVerReq{Conn: conn, Slug: req.TableSlug, Id: req.TableId})
		if err != nil {
			return nil, err
		}
		req.TableId = cast.ToString(table["id"])
		req.TableSlug = cast.ToString(table["slug"])
	}

	payload := make(map[string]interface{})
	payload["table_id"] = req.TableId
	if req.IsDefault {
		payload["is_default"] = true
	}
	if req.MenuId != "" {
		payload["menu_id"] = req.MenuId
	}
	var menu_id sql.NullString
	query := `
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
		WHERE table_id = $1`

	var args []interface{}
	args = append(args, payload["table_id"])

	if menuID, ok := payload["menu_id"]; ok {
		query += ` AND menu_id = $2`
		args = append(args, menuID)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err

	}
	defer rows.Close()
	layouts := make([]*nb.LayoutResponse, 0)
	for rows.Next() {
		layout := nb.LayoutResponse{}
		err := rows.Scan(&layout.Id, &layout.Label, &layout.Order, &layout.Type,
			&layout.Icon, &layout.IsDefault, &layout.IsModal,
			&layout.IsVisibleSection, &layout.Attributes, &layout.TableId, &menu_id)
		if err != nil {
			return nil, err
		}
		if menu_id.Valid {
			layout.MenuId = menu_id.String
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
					var fieldResp nb.Field
					err := conn.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
					if err != nil {
						if err != pgx.ErrNoRows {
							return nil, err
						}
					}

					var relation nb.RelationForGetAll
					err = conn.QueryRow(ctx, `SELECT
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
				"cascading_tree_field_slug",
				"dynamic_tables" FROM relation WHERE id = $1`, relationID).Scan(&relationID,
						&relation.TableFrom,
						&relation.TableTo,
						&relation.FieldFrom,
						&relation.FieldTo,
						&relation.Type,
						&relation.ViewFields,
						&relation.RelationFieldSlug,
						&relation.DynamicTables,
						&relation.IsEditable,
						&relation.IsUserIdDefault,
						&relation.Cascadings,
						&relation.IsSystem,
						&relation.ObjectIdFromJwt,
						&relation.CascadingTreeTableSlug,
						&relation.CascadingTreeFieldSlug,
						&relation.DynamicTables)

					if err != nil {
						if err == pgx.ErrNoRows {
							continue
						}
						return nil, err
					}
					viewOfRelation := nb.View{}
					err = conn.QueryRow(ctx, "SELECT view_fields, function_path, attributes, is_editable FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, req.TableSlug).Scan(viewOfRelation.ViewFields, &viewOfRelation.FunctionPath, &viewOfRelation.Attributes, &viewOfRelation.IsEditable)
					if err != nil {
						if err != pgx.ErrNoRows {
							return nil, err
						}
					}
					var viewFieldIds []*nb.Field
					if len(viewOfRelation.ViewFields) > 0 {
						viewFieldIds = make([]*nb.Field, len(viewOfRelation.ViewFields))
						for i, id := range viewOfRelation.ViewFields {
							viewFieldIds[i] = &nb.Field{Id: id}
						}
					}

					var fieldAsAttribute []string

					if relation.Id != "" {
						for _, fieldID := range viewFieldIds {
							var field nb.FieldResponse
							err := conn.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
							if err != nil {
								if err == pgx.ErrNoRows {
									continue
								}
								return nil, err
							}

							if req.LanguageSetting != "" && field.EnableMultilanguage {
								if strings.HasSuffix(field.Slug, "_"+req.LanguageSetting) {
									fieldAsAttribute = append(fieldAsAttribute, field.Slug)
								} else {
									continue
								}
							} else {
								fieldAsAttribute = append(fieldAsAttribute, field.Slug)
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
								var viewOfDynamicRelation nb.View
								err = conn.QueryRow(ctx, "SELECT id, relation_id, relation_table_slug FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relation.Id, dynamicTable.TableSlug).Scan(&viewOfDynamicRelation.Id, &viewOfDynamicRelation.RelationId, &viewOfDynamicRelation.RelationTableSlug)
								if err != nil {
									if err != pgx.ErrNoRows {
										return nil, err
									}
								}
								if err != nil {
									return nil, err
								}
								if len(viewOfDynamicRelation.ViewFields) > 0 {
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

							originalAttributes["autofill"] = autofillFields
							originalAttributes["view_fields"] = fieldAsAttribute
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
							originalAttributes["autofill"] = autofillFields
							originalAttributes["view_fields"] = fieldAsAttribute
							originalAttributes["auto_filters"] = relation.AutoFilters
							originalAttributes["relation_field_slug"] = relation.RelationFieldSlug
							originalAttributes["dynamic_tables"] = dynamicTables
							originalAttributes["is_user_id_default"] = relation.IsUserIdDefault
							originalAttributes["object_id_from_jwt"] = relation.ObjectIdFromJwt
							originalAttributes["cascadings"] = relation.Cascadings
							originalAttributes["cascading_tree_table_slug"] = relation.CascadingTreeTableSlug
							originalAttributes["cascading_tree_field_slug"] = relation.CascadingTreeFieldSlug
							originalAttributes["function_path"] = viewOfRelation.FunctionPath

							for k, v := range viewOfRelation.Attributes.AsMap() {
								originalAttributes[k] = v
							}

						}
						if len(viewOfRelation.DefaultValues) > 0 {
							originalAttributes["default_values"] = viewOfRelation.DefaultValues
						}
						originalAttributes["creatable"] = viewOfRelation.Creatable

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
		layoutIDs = append(layoutIDs, layout.Id)

		var (
			label      sql.NullString
			order      sql.NullInt32
			icon       sql.NullString
			relationId sql.NullString
		)
		tabs := make([]*nb.TabResponse, 0)
		sqlQuery := `
		SELECT
			"id",
			"type",
			"order",
			"label",
			"icon",
			"layout_id",
			"relation_id",
			"attributes"
		FROM "tab"
		WHERE "layout_id"::varchar = ANY($1);

		`

		rows, err := conn.Query(ctx, sqlQuery, pq.Array(layoutIDs))
		if err != nil {
			return nil, err
		}

		defer rows.Close()
		for rows.Next() {
			var tab nb.TabResponse
			err := rows.Scan(&tab.Id, &tab.Type, &order, &label, &icon, &tab.LayoutId, &relationId, &tab.Attributes)
			if err != nil {
				return nil, err
			}
			if label.Valid {
				tab.Label = label.String
			}
			if order.Valid {
				tab.Order = order.Int32
			}

			if icon.Valid {
				tab.Icon = icon.String
			}

			if relationId.Valid {
				tab.RelationId = relationId.String
			}

			tabs = append(tabs, &tab)
		}
		if rows.Err() != nil {
			return nil, rows.Err()
		}
		layout.Tabs = tabs

		sectionRepo := NewSectionRepo(conn)
		relationRepo := NewRelationRepo(conn)

		for _, tab := range tabs {
			if tab.Type == "section" {
				sections, err := sectionRepo.GetAll(ctx, &nb.GetAllSectionsRequest{
					ProjectId: req.ProjectId,
					RoleId:    req.RoleId,
					TableSlug: req.TableSlug,
					TableId:   req.TableId,
					TabId:     tab.Id,
				})
				if err != nil {
					return nil, err
				}
				tab.Sections = sections.Sections
			} else if tab.Type == "relation" && tab.RelationId != "" {

				relation, err := relationRepo.GetSingleViewForRelation(ctx, models.ReqForViewRelation{
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
			mapTab := make(map[string][]*nb.TabResponse)
			for _, tab := range tabs {
				if _, ok := mapTab[tab.LayoutId]; ok {
					mapTab[tab.LayoutId] = append(mapTab[tab.LayoutId], tab)
					arrOfObjects := mapTab[tab.LayoutId]
					sort.Slice(arrOfObjects, func(i, j int) bool {
						return arrOfObjects[i].Order < arrOfObjects[j].Order
					})
					mapTab[tab.LayoutId] = arrOfObjects
				} else {
					mapTab[tab.LayoutId] = []*nb.TabResponse{tab}
				}
			}

			if len(mapTab) > 0 {
				for _, layout := range layouts {
					layout.Tabs = mapTab[layout.Id]
				}
			}
		}
		layout.Tabs = tabs
	}

	resp.Layouts = layouts

	return resp, nil
}

func (l *layoutRepo) RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) error {

	conn := psqlpool.Get(req.GetProjectId())

	rows, err := conn.Query(ctx, "SELECT id FROM tab WHERE layout_id = $1", req.Id)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tabIDs []string
	for rows.Next() {
		var tabID string
		if err := rows.Scan(&tabID); err != nil {
			return err
		}
		tabIDs = append(tabIDs, tabID)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if _, err := conn.Exec(ctx, "DELETE FROM section WHERE tab_id = ANY($1)", pq.Array(tabIDs)); err != nil {
		return err
	}

	if _, err := conn.Exec(ctx, "DELETE FROM tab WHERE id = ANY($1)", pq.Array(tabIDs)); err != nil {
		return err
	}

	if _, err := conn.Exec(ctx, "DELETE FROM layout WHERE id = $1", req.Id); err != nil {
		return err
	}

	return nil
}

func (l *layoutRepo) GetByID(ctx context.Context, req *nb.LayoutPrimaryKey) (*nb.LayoutResponse, error) {
	resp := &nb.LayoutResponse{}
	conn := psqlpool.Get(req.GetProjectId())

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
	var (
		menuID     sql.NullString
		attributes []byte
	)

	err := conn.QueryRow(ctx, layoutQuery, req.Id).Scan(
		&resp.Id,
		&resp.Label,
		&resp.Order,
		&resp.Type,
		&resp.Icon,
		&resp.IsDefault,
		&resp.IsModal,
		&resp.IsVisibleSection,
		&attributes,
		&resp.TableId,
		&menuID,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return nil, err
	}

	tabQuery := `
        SELECT
            id,
            "order",
            label,
            icon,
            "type",
            layout_id,
            relation_id,
            attributes
        FROM "tab"
        WHERE layout_id = $1
    `

	rows, err := conn.Query(ctx, tabQuery, req.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			tab           nb.TabResponse
			typeNull      sql.NullString
			relationID    sql.NullString
			tabAttributes []byte
			icon          sql.NullString
		)
		err := rows.Scan(
			&tab.Id,
			&tab.Order,
			&tab.Label,
			&icon,
			&typeNull,
			&tab.LayoutId,
			&relationID,
			&tabAttributes,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(attributes, &tab.Attributes); err != nil {
			return nil, err
		}

		if icon.Valid {
			tab.Icon = icon.String
		}
		if relationID.Valid {
			tab.RelationId = relationID.String
		}

		if typeNull.Valid {
			tab.Type = typeNull.String
		}

		resp.Tabs = append(resp.Tabs, &tab)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(resp.Tabs, func(i, j int) bool {
		return resp.Tabs[i].Order < resp.Tabs[j].Order
	})

	return resp, nil
}

func (l *layoutRepo) GetAllV2(ctx context.Context, req *nb.GetListLayoutRequest) (resp *nb.GetListLayoutResponse, err error) {
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()
	resp = &nb.GetListLayoutResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	query := `SELECT jsonb_build_object (
		'id', l.id,
		'label', l.label,
		'order', l."order",
		'table_id', l.table_id,
		'type', l."type",
		'is_default', l.is_default,
		'is_modal', l.is_modal,
		'tabs', (
			SELECT jsonb_agg(
					jsonb_build_object(
						'id', t.id,
						'label', t.label,
						'layout_id', t.layout_id,
						'type', t.type,
						'order', t."order",
						'relation_id', t.relation_id::varchar,
						'attributes', t.attributes
					)
				)
			FROM tab t 
			WHERE t.layout_id = l.id
		)
	) AS DATA 
	FROM layout l 
	JOIN "table" ta ON ta.id = l.table_id
	WHERE ta.slug = $1
	GROUP BY l.id
	ORDER BY l."order" ASC;
	`

	layoutRows, err := conn.Query(ctx, query, req.TableSlug)
	if err != nil {
		return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error querying layout")
	}
	defer layoutRows.Close()

	for layoutRows.Next() {
		var (
			layout = nb.LayoutResponse{}
			body   = []byte{}
		)

		err = layoutRows.Scan(
			&body,
		)
		if err != nil {
			return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error scanning layout")
		}

		if err := json.Unmarshal(body, &layout); err != nil {
			return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error unmarshalling layout")
		}

		resp.Layouts = append(resp.Layouts, &layout)
	}

	fieldQuery := `SELECT 
		f.id,
		f.type,
		f.index,
		f.label,
		f.slug,
		f.table_id,
		f.attributes
	FROM "field" f 
	JOIN "table" t ON t.id = f.table_id 
	WHERE t.slug = $1`

	fieldRows, err := conn.Query(ctx, fieldQuery, req.TableSlug)
	if err != nil {

		return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error querying field")
	}
	defer fieldRows.Close()

	fields := make(map[string]*nb.FieldResponse)

	for fieldRows.Next() {
		var (
			field     = nb.FieldResponse{}
			att       = []byte{}
			indexNull sql.NullString
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&indexNull,
			&field.Label,
			&field.Slug,
			&field.TableId,
			&att,
		)
		if err != nil {
			return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error scanning field")
		}

		field.Index = indexNull.String

		fields[field.Id] = &field
	}

	for _, layout := range resp.Layouts {
		for _, tab := range layout.Tabs {
			if tab.Type == "section" {
				section, err := GetSections(ctx, conn, tab.Id, "", "", fields)
				if err != nil {
					return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error getting sections")
				}
				tab.Sections = section
			} else if tab.Type == "relation" {
				relation, err := GetRelation(ctx, conn, tab.RelationId)
				if err != nil {
					return &nb.GetListLayoutResponse{}, errors.Wrap(err, "error getting relation")
				}
				relation.Attributes = tab.Attributes
				relation.RelationTableSlug = relation.TableFrom.Slug
				tab.Relation = relation
			}
		}
	}

	return resp, nil
}

func (l *layoutRepo) GetSingleLayoutV2(ctx context.Context, req *nb.GetSingleLayoutRequest) (resp *nb.LayoutResponse, err error) {

	resp = &nb.LayoutResponse{}

	conn := psqlpool.Get(req.GetProjectId())

	if req.MenuId == "" {
		return &nb.LayoutResponse{}, fmt.Errorf("menu_id is required")
	}

	if req.TableId == "" {
		query := `SELECT id FROM "table" WHERE slug = $1`

		err = conn.QueryRow(ctx, query, req.TableSlug).Scan(&req.TableId)
		if err != nil {
			return &nb.LayoutResponse{}, fmt.Errorf("table_id is required")
		}
	}

	count := 0

	query := `SELECT COUNT(*) FROM "layout" WHERE table_id = $1 AND menu_id = $2`

	err = conn.QueryRow(ctx, query, req.TableId, req.MenuId).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return &nb.LayoutResponse{}, err
	}

	query = ``

	var (
		layout = nb.LayoutResponse{}
		body   = []byte{}
	)

	if count == 0 {
		query = `SELECT jsonb_build_object (
			'id', l.id,
			'label', l.label,
			'order', l."order",
			'table_id', l.table_id,
			'type', l."type",
			'is_default', l.is_default,
			'is_modal', l.is_modal,
			'tabs', (
				SELECT jsonb_agg(
						jsonb_build_object(
							'id', t.id,
							'label', t.label,
							'layout_id', t.layout_id,
							'type', t.type,
							'order', t."order",
							'relation_id', t.relation_id::varchar,
							'attributes', t.attributes
						)
						ORDER BY t."order" ASC
					)
				FROM tab t 
				WHERE t.layout_id = l.id
			)
		) AS DATA 
		FROM layout l 
		JOIN "table" ta ON ta.id = l.table_id
		WHERE ta.slug = $1 AND l.is_default = true
		GROUP BY l.id
		ORDER BY l."order" ASC;`

		err = conn.QueryRow(ctx, query, req.TableSlug).Scan(&body)
		if err != nil {
			return &nb.LayoutResponse{}, err
		}
	} else {
		query = `SELECT jsonb_build_object (
			'id', l.id,
			'label', l.label,
			'order', l."order",
			'table_id', l.table_id,
			'type', l."type",
			'is_default', l.is_default,
			'is_modal', l.is_modal,
			'tabs', (
				SELECT jsonb_agg(
						jsonb_build_object(
							'id', t.id,
							'label', t.label,
							'layout_id', t.layout_id,
							'type', t.type,
							'order', t."order",
							'relation_id', t.relation_id::varchar,
							'attributes', t.attributes
						)
					)
				FROM tab t 
				WHERE t.layout_id = l.id
			)
		) AS DATA 
		FROM layout l 
		JOIN "table" ta ON ta.id = l.table_id
		WHERE ta.slug = $1 AND l.menu_id = $2
		GROUP BY l.id
		ORDER BY l."order" ASC;`

		err = conn.QueryRow(ctx, query, req.TableSlug, req.MenuId).Scan(&body)
		if err != nil {
			return &nb.LayoutResponse{}, err
		}
	}

	if err := json.Unmarshal(body, &layout); err != nil {
		return &nb.LayoutResponse{}, err
	}

	fieldQuery := `SELECT 
		f.id,
		f.type,
		f.index,
		f.label,
		f.slug,
		f.table_id,
		f.attributes,

		fp.guid,
		fp.field_id,
		fp.role_id,
		fp.table_slug,
		fp.label,
		fp.view_permission,
		fp.edit_permission

	FROM "field" f 
	JOIN "table" t ON t.id = f.table_id 
	LEFT JOIN "field_permission" fp ON fp.field_id = f.id
	WHERE t.slug = $1 AND fp.role_id = $2`

	fieldRows, err := conn.Query(ctx, fieldQuery, req.TableSlug, req.RoleId)
	if err != nil {
		return &nb.LayoutResponse{}, err
	}
	defer fieldRows.Close()

	fields := make(map[string]*nb.FieldResponse)

	for fieldRows.Next() {
		var (
			field       = nb.FieldResponse{}
			att         = []byte{}
			indexNull   sql.NullString
			fPermission = FieldPermission{}
			attributes  = make(map[string]interface{})
		)

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&indexNull,
			&field.Label,
			&field.Slug,
			&field.TableId,
			&att,

			&fPermission.Guid,
			&fPermission.FieldId,
			&fPermission.RoleId,
			&fPermission.TableSlug,
			&fPermission.Label,
			&fPermission.ViewPermission,
			&fPermission.EditPermission,
		)
		if err != nil {
			return &nb.LayoutResponse{}, err
		}

		if err := json.Unmarshal(att, &attributes); err != nil {
			return &nb.LayoutResponse{}, err
		}

		attributes["field_permission"] = fPermission

		atr, err := helper.ConvertMapToStruct(attributes)
		if err != nil {
			return &nb.LayoutResponse{}, err
		}

		field.Attributes = atr
		field.Index = indexNull.String

		fields[field.Id] = &field
	}

	for _, tab := range layout.Tabs {
		if tab.Type == "section" {
			section, err := GetSections(ctx, conn, tab.Id, req.RoleId, req.TableSlug, fields)
			if err != nil {
				return &nb.LayoutResponse{}, err
			}
			tab.Sections = section
		} else if tab.Type == "relation" {
			relation, err := GetRelation(ctx, conn, tab.RelationId)
			if err != nil {
				return &nb.LayoutResponse{}, err
			}
			relation.Attributes = tab.Attributes
			relation.RelationTableSlug = relation.TableFrom.Slug
			tab.Relation = relation
		}
	}

	return &layout, nil
}

func GetSections(ctx context.Context, conn *pgxpool.Pool, tabId, roleId, tableSlug string, fields map[string]*nb.FieldResponse) ([]*nb.SectionResponse, error) {

	sectionQuery := `SELECT 
		id,
		"order",
		fields,
		attributes
	FROM "section" WHERE tab_id = $1
	`

	sectionRows, err := conn.Query(ctx, sectionQuery, tabId)
	if err != nil {
		return nil, errors.Wrap(err, "error querying section")
	}
	defer sectionRows.Close()

	sections := []*nb.SectionResponse{}

	for sectionRows.Next() {
		var (
			section    = nb.SectionResponse{}
			body       = []byte{}
			fieldBody  = []SectionFields{}
			attributes = []byte{}
		)

		err = sectionRows.Scan(
			&section.Id,
			&section.Order,
			&body,
			&attributes,
		)
		if err != nil {
			return nil, errors.Wrap(err, "error scanning section")
		}
		if err := json.Unmarshal(body, &fieldBody); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling section")
		}

		if err := json.Unmarshal(attributes, &section.Attributes); err != nil {
			return nil, errors.Wrap(err, "error unmarshalling section attributes")
		}

		for i, f := range fieldBody {

			if strings.Contains(f.Id, "#") {

				fBody := []nb.FieldResponse{}

				if err := json.Unmarshal(body, &fBody); err != nil {
					return nil, errors.Wrap(err, "error unmarshalling field body")
				}

				temp, err := helper.ConvertStructToMap(fBody[i].Attributes)
				if err != nil {
					return nil, errors.Wrap(err, "error converting struct to map")
				}

				if temp == nil {
					temp = make(map[string]interface{})
				}

				fieldsSlice := cast.ToSlice(temp["fields"])
				if fieldsSlice != nil {
					attributes := cast.ToStringMap(cast.ToStringMap(fieldsSlice[0])["attributes"])
					for key, val := range attributes {
						temp[key] = val
					}
				}

				newAttributes, err := helper.ConvertMapToStruct(temp)
				if err != nil {
					return nil, errors.Wrap(err, "error converting map to struct")
				}

				fBody[i].Attributes = newAttributes

				if roleId != "" {

					relationId := strings.Split(f.Id, "#")[1]

					fieldId := ""

					queryF := `SELECT f.id FROM "field" f JOIN "table" t ON t.id = f.table_id WHERE f.relation_id = $1 AND t.slug = $2`

					err = conn.QueryRow(ctx, queryF, relationId, tableSlug).Scan(&fieldId)
					if err != nil {
						return nil, errors.Wrap(err, "error querying field")
					}

					autoFiltersBody := []byte{}
					autoFilters := []map[string]interface{}{}
					viewFieldsBody := []map[string]interface{}{}
					viewFields := []string{}

					queryR := `SELECT COALESCE(r."auto_filters", '[{}]'), r."view_fields" FROM "relation" r WHERE r."id" = $1`

					err = conn.QueryRow(ctx, queryR, relationId).Scan(&autoFiltersBody, &viewFields)
					if err != nil {
						return nil, errors.Wrap(err, "error querying autoFiltersBody")
					}

					if err = json.Unmarshal(autoFiltersBody, &autoFilters); err != nil {
						return nil, errors.Wrap(err, "error unmarshal")
					}

					for _, id := range viewFields {

						slug := ""
						queryF := `SELECT slug FROM field WHERE id = $1`

						err = conn.QueryRow(ctx, queryF, id).Scan(&slug)
						if err != nil {
							return nil, errors.Wrap(err, "error get field")
						}

						viewFieldsBody = append(viewFieldsBody, map[string]interface{}{
							"slug": slug,
						})
					}

					attributes, err := helper.ConvertStructToMap(fBody[i].Attributes)
					if err != nil {
						return nil, errors.Wrap(err, "error converting struct to map")
					}

					permission := FieldPermission{}

					query := `SELECT
						guid,
						field_id,
						role_id,
						table_slug,
						label,
						view_permission,
						edit_permission
					FROM field_permission WHERE field_id = $1 AND role_id = $2`

					err = conn.QueryRow(ctx, query, fieldId, roleId).Scan(
						&permission.Guid,
						&permission.FieldId,
						&permission.RoleId,
						&permission.TableSlug,
						&permission.Label,
						&permission.ViewPermission,
						&permission.EditPermission,
					)
					if err != nil {
						return nil, errors.Wrap(err, "error querying field permission")
					}

					attributes["field_permission"] = permission
					attributes["auto_filters"] = autoFilters
					attributes["view_fields"] = viewFieldsBody

					bodyAtt, err := helper.ConvertMapToStruct(attributes)
					if err != nil {
						return nil, errors.Wrap(err, "error converting map to struct")
					}

					fBody[i].Attributes = bodyAtt
				}

				section.Fields = append(section.Fields, &fBody[i])
			} else {
				fBody, ok := fields[f.Id]

				if !ok {
					field := &nb.FieldResponse{}
					field.Attributes = f.Attributes
					field.Order = int32(f.Order)
					field.Id = f.Id
					section.Fields = append(section.Fields, field)
					continue
				}

				if fBody != nil {
					fBody.Order = int32(f.Order)
				}

				section.Fields = append(section.Fields, fBody)
			}

		}

		sections = append(sections, &section)
	}

	return sections, nil
}

func GetRelation(ctx context.Context, conn *pgxpool.Pool, relationId string) (*nb.RelationForSection, error) {
	query := `SELECT
		r.id,
		r.type,
		r.view_fields,


		t1.id,
		t1.label,
		t1.slug,
		t1.show_in_menu,
		t1.icon,

		t2.id,
		t2.label,
		t2.slug,
		t2.show_in_menu,
		t2.icon

	FROM "relation" r 
	LEFT JOIN "table" t1 ON t1.slug = table_from
	LEFT JOIN "table" t2 ON t2.slug = table_to
	WHERE r.id = $1
	`

	relation := nb.RelationForSection{
		TableFrom: &nb.TableForSection{},
		TableTo:   &nb.TableForSection{},
	}
	viewFields := []string{}

	err := conn.QueryRow(ctx, query, relationId).Scan(
		&relation.Id,
		&relation.Type,
		&viewFields,

		&relation.TableFrom.Id,
		&relation.TableFrom.Label,
		&relation.TableFrom.Slug,
		&relation.TableFrom.ShowInMenu,
		&relation.TableFrom.Icon,
		&relation.TableTo.Id,
		&relation.TableTo.Label,
		&relation.TableTo.Slug,
		&relation.TableTo.ShowInMenu,
		&relation.TableTo.Icon,
	)
	if err != nil {
		return &nb.RelationForSection{}, err
	}

	query = `SELECT
		id,
		type,
		index,
		label,
		slug,
		table_id,
		attributes,
		is_search
	FROM "field" WHERE id IN ($1)`

	fieldRows, err := conn.Query(ctx, query, pq.Array(viewFields))
	if err != nil {
		return &nb.RelationForSection{}, err
	}
	defer fieldRows.Close()

	for fieldRows.Next() {
		att := []byte{}
		field := nb.Field{}

		err = fieldRows.Scan(
			&field.Id,
			&field.Type,
			&field.Index,
			&field.Label,
			&field.TableId,
			&att,
			&field.IsSearch,
		)
		if err != nil {
			return &nb.RelationForSection{}, err
		}
		if err := json.Unmarshal(att, &field.Attributes); err != nil {
			return &nb.RelationForSection{}, err
		}

		relation.ViewFields = append(relation.ViewFields, &field)
	}

	permission := RelationFields{}

	query = `SELECT 
		guid,
		role_id,
		relation_id,
		table_slug,
		view_permission,
		create_permission,
		edit_permission,
		delete_permission
	FROM "view_relation_permission" WHERE relation_id = $1`

	err = conn.QueryRow(ctx, query, relationId).Scan(
		&permission.Guid,
		&permission.RoleId,
		&permission.RelationId,
		&permission.TableSlug,
		&permission.ViewPermission,
		&permission.CreatePermission,
		&permission.EditPermission,
		&permission.DeletePermission,
	)
	if err != nil {
		return &nb.RelationForSection{}, err
	}

	marshledInputMap, err := json.Marshal(permission)
	outputStruct := &structpb.Struct{}
	if err != nil {
		return &nb.RelationForSection{}, err
	}
	err = protojson.Unmarshal(marshledInputMap, outputStruct)
	if err != nil {
		return &nb.RelationForSection{}, err
	}

	relation.Permission = outputStruct

	return &relation, nil
}

type SectionFields struct {
	Id         string           `json:"id"`
	Order      int              `json:"order"`
	Attributes *structpb.Struct `json:"attributes"`
}
type RelationFields struct {
	Guid             string `json:"guid"`
	RoleId           string `json:"role_id"`
	RelationId       string `json:"relation_id"`
	TableSlug        string `json:"table_slug"`
	ViewPermission   bool   `json:"view_permission"`
	CreatePermission bool   `json:"create_permission"`
	EditPermission   bool   `json:"edit_permission"`
	DeletePermission bool   `json:"delete_permission"`
}

type FieldPermission struct {
	Guid           string `json:"guid"`
	Label          string `json:"label"`
	FieldId        string `json:"field_id"`
	RoleId         string `json:"role_id"`
	TableSlug      string `json:"table_slug"`
	ViewPermission bool   `json:"view_permission"`
	EditPermission bool   `json:"edit_permission"`
}
