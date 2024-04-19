package postgres

import (
	"context"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
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
	// l.db := psqlpool.Get(req.ProjectId)
	// defer l.db.Close()

	tx, err := l.db.Begin(ctx)
	if err != nil {
		return resp, err
	}
	var layoutId string
	defer tx.Rollback(ctx)
	if req.Id == "" {
		layoutId = uuid.New().String()
	} else {
		layoutId = req.Id
	}

	query := `
        INSERT INTO "layout" (
            "id", "label", "order", "type", "icon", "is_default", 
            "is_modal", "is_visible_section",
            "attributes", "table_id", "menu_id"
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
            "attributes" = EXCLUDED.attributes,	
            "table_id" = EXCLUDED.table_id,
            "menu_id" = EXCLUDED.menu_id
    `
	_, err = tx.Exec(ctx, query,
		layoutId, req.Label, req.Order, req.Type, req.Icon,
		req.IsDefault, req.IsModal, req.IsVisibleSection,
		req.Attributes, req.TableId, req.MenuId)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("error inserting layout: %w", err)
	}

	if req.IsDefault {
		_, err = tx.Exec(ctx, `
            UPDATE layout
            SET is_default = false
            WHERE table_id = $1 AND id != $2
        `, req.TableId, layoutId)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error updating layout: %w", err)

		}
	}

	var (
		bulkWriteTab      []string
		bulkWriteSection  []string
		mapTabs           = make(map[string]int)
		mapSections       = make(map[string]int)
		deletedTabIds     []string
		deletedSectionIds []string
		// insertManyRelationPermissions []string
	)

	rows, err := tx.Query(ctx, "SELECT id FROM tab WHERE layout_id = $1", layoutId)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("error fetching tabs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tabId string
		if err := rows.Scan(&tabId); err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error scanning tab ID: %w", err)
		}
		mapTabs[tabId] = 1
	}

	for _, tab := range req.Tabs {
		if tab.Id == "" {
			tab.Id = uuid.New().String()
		}

		if _, ok := mapTabs[tab.Id]; ok {
			mapTabs[tab.Id] = 2
		}

		bulkWriteTab = append(bulkWriteTab, fmt.Sprintf(`
			INSERT INTO "tab" (
				"id", "label", "layout_id",  "type",
				"order", "icon"
			) VALUES ('%s', '%s', '%s', '%s', %d, '%s')
			ON CONFLICT (id) DO UPDATE
			SET
				"label" = EXCLUDED.label,
				"layout_id" = EXCLUDED.layout_id,
				"type" = EXCLUDED.type,
				"order" = EXCLUDED.order,
				"icon" = EXCLUDED.icon
		`, tab.Id, tab.Label, layoutId, tab.Type,
			tab.Order, tab.Icon))

		for _, query := range bulkWriteTab {
			_, err := tx.Exec(ctx, query)
			if err != nil {
				tx.Rollback(ctx)
				return nil, fmt.Errorf("error executing bulkWriteTab query: %w", err)
			}
		}
		for _, section := range tab.Sections {
			if section.Id == "" {
				section.Id = uuid.New().String()

			}

			if _, ok := mapSections[section.Id]; ok {
				mapSections[section.Id] = 2

			}

			for _, section := range tab.Sections {
				if section.Id == "" {
					section.Id = uuid.New().String()
					fmt.Println("_-----------d------3333---------")

				}
				if _, ok := mapSections[section.Id]; ok {
					mapSections[section.Id] = 2
				fmt.Println("_-----------d---------------")
			}
				bulkWriteSection = append(bulkWriteSection, fmt.Sprintf(`
					INSERT INTO "section" (
						"id", "tab_id", "label", "order", "icon"
					) VALUES ('%s', '%s', '%s', %d, '%s')
					ON CONFLICT (id) DO UPDATE
					SET
						"tab_id" = EXCLUDED.tab_id,
						"label" = EXCLUDED.label,
						"order" = EXCLUDED.order,
						"icon" = EXCLUDED.icon
				`, section.Id, tab.Id, section.Label, section.Order, section.Icon))
			}

			for _, query := range bulkWriteSection {
				_, err := tx.Exec(ctx, query)
				if err != nil {
					tx.Rollback(ctx)
					return nil, fmt.Errorf("error executing bulkWriteSection query: %w", err)
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
				fmt.Println(deletedTabIds , "deletedTabIds")
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

			// rows, err = tx.Query(ctx, "SELECT guid FROM role")
			// if err != nil {
			// tx.Rollback(ctx)
			// 	return nil, fmt.Errorf("error fetching roles: %w", err)
			// }
			// defer rows.Close()

			// roles := []string{}
			// for rows.Next() {
			// 	var roleGUID string
			// 	if err := rows.Scan(&roleGUID); err != nil {
			// 		return nil, fmt.Errorf("error scanning role GUID: %w", err)
			// 	}
			// 	roles = append(roles, roleGUID)
			// }

			//Permission

			// relation_ids := make([]string, 0)
			// for _, role := range roles {
			// 	for _, relationID := range relation_ids {

			// 		var exists int
			// 		query := `
			//         SELECT COUNT(*)
			//         FROM relation_permission
			//         WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
			//     `
			// 		err := tx.QueryRow(ctx, query, role, req.TableId, relationID).Scan(&exists)
			// 		if err != nil {
			// 			return nil, fmt.Errorf("error checking relation permission existence: %w", err)
			// 		}
			// 		if exists == 0 {
			// 			insertManyRelationPermissions = append(insertManyRelationPermissions, fmt.Sprintf(`
			//             INSERT INTO relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
			//             VALUES ('%s', '%s', '%s', true, true, true, true)
			//         `, role, req.TableId, relationID))
			// 		}
			// 	}

			// 	if len(insertManyRelationPermissions) > 0 {
			// 		for _, query := range insertManyRelationPermissions {
			// 			_, err := tx.Exec(ctx, query)
			// 			if err != nil {
			// 				return nil, fmt.Errorf("error inserting relation permissions: %w", err)
			// 			}
			// 		}
			// 	}

			// }

		}
	}
	return &nb.LayoutResponse{}, nil
}
func (l layoutRepo) RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) error {

	// l.db := psqlpool.Get(req.ProjectId)

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

// func (l layoutRepo) GetSingle(ctx context.Context, req *nb.GetSingleLayoutRequest) (*nb.LayoutResponse, error) {
// 	l.db := psqlpool.Get(req.ProjectId)
// 	defer l.db.Close()

// 	resp := &nb.LayoutResponse{}

// 	if req.TableId == "" {
// 		tableQuery := `
//             SELECT
//                 id
//             FROM "table" WHERE "slug" = $1
//         `
// 		var tableID string
// 		err := l.db.QueryRow(ctx, tableQuery, req.TableSlug).Scan(&tableID)
// 		if err != nil {
// 			return resp, err
// 		}
// 		req.TableId = tableID
// 	}

// 	layoutQuery := `
//         SELECT
//             id,
//             label,
//             "order",
//             "type",
//             icon,
//             is_default,
//             is_modal,
//             is_visible_section,
//             summary_fields,
//             attributes,
//             table_id,
//             menu_id
//         FROM layout
//         WHERE table_id = $1 AND menu_id = $2
//     `

// 	row := l.db.QueryRow(ctx, layoutQuery, req.TableId, req.MenuId)
// 	err := row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.SummaryFields, &resp.Attributes, &resp.TableId, &resp.MenuId)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			layoutQuery = `
//                 SELECT
//                     id,
//                     label,
//                     "order",
//                     "type",
//                     icon,
//                     is_default,
//                     is_modal,
//                     is_visible_section,
//                     summary_fields,
//                     attributes,
//                     table_id,
//                     menu_id
//                 FROM layout
//                 WHERE table_id = $1 AND is_default = true
//             `
// 			row := l.db.QueryRow(ctx, layoutQuery, req.TableId)
// 			err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.SummaryFields, &resp.Attributes, &resp.TableId, &resp.MenuId)
// 			if err != nil {
// 				return resp, err
// 			}
// 		} else {
// 			return resp, err
// 		}
// 	}

// 	table, err := helper.TableVer(ctx, l.db, req.TableId, req.TableSlug)
// 	if err != nil {
// 		return resp, err
// 	}
// 	req.TableId = table["id"].(string)
// 	req.TableSlug = table["slug"].(string)

// 	if len(resp.SummaryFields) > 0 {
// 		for _, fieldReq := range resp.SummaryFields {
// 			field := &nb.Field2{}

// 			if strings.Contains(fieldReq.Id, "#") {
// 				field.Id = fieldReq.Id
// 				field.Label = fieldReq.Label
// 				field.Type = fieldReq.RelationType
// 				relationID := strings.Split(fieldReq.Id, "#")[1]

// 				var fieldResp = &nb.Field{}
// 				err := l.db.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
// 				if err != nil {
// 					if err != pgx.ErrNoRows {
// 						return resp, err
// 					}
// 					continue
// 				}
// 				field.Slug = fieldResp.Slug
// 				field.Required = fieldResp.Required

// 				var view_of_relation = &nb.View{}
// 				err = l.db.QueryRow(ctx, "SELECT relation_table_slug, view_fields FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relationID, req.TableSlug).Scan(&view_of_relation.RelationTableSlug, &view_of_relation.ViewFields)
// 				if err != nil && err != pgx.ErrNoRows {
// 					return resp, err
// 				}

// 				viewFieldIds := view_of_relation.ViewFields
// 				if view_of_relation.RelationTableSlug != "" && len(view_of_relation.ViewFields) > 0 {
// 					viewFieldIds = view_of_relation.ViewFields
// 				}

// 				for _, fieldID := range viewFieldIds {
// 					var field = &nb.Field{}
// 					err := l.db.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
// 					if err != nil {
// 						if err != pgx.ErrNoRows {
// 							return resp, err
// 						}
// 						continue
// 					}

// 					var isEditable bool
// 					err = l.db.QueryRow(ctx, "SELECT is_editable FROM view WHERE relation_id = $1", relationID).Scan(&isEditable)
// 					if err != nil && err != pgx.ErrNoRows {
// 						return resp, err
// 					}
// 				}

// 				resp.SummaryFields = append(resp.SummaryFields, &nb.FieldResponse{
// 					Id:       field.Id,
// 					Label:    field.Label,
// 					Type:     field.Type,
// 					Slug:     field.Slug,
// 					Required: field.Required,
// 					// EnableMultilanguage: field.
// 				})
// 			}
// 		}
// 	}
// 	tabsQuery := `
//     SELECT
//         id,
//         label,
//         "order",
//         type,
//         relation_id
//     FROM tab
//     WHERE layout_id = $1
// `
// 	rows, err := l.db.Query(ctx, tabsQuery, resp.Id)
// 	if err != nil {
// 		return resp, err
// 	}
// 	defer rows.Close()

// 	// var tabs = resp.Tabs
// 	sectionsQuery := `
//     SELECT
//         id,
//         label,
//         "order",
//         type,
//         relation_id
//     FROM
//         tab
//     WHERE
//         layout_id = $1 AND type = 'section'
// `
// 	sectionsRows, err := l.db.Query(ctx, sectionsQuery, resp.Id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer sectionsRows.Close()

// 	var sections []*nb.Section

// 	for sectionsRows.Next() {
// 		var section nb.Section
// 		err := sectionsRows.Scan(&section.Id, &section.Label, &section.Order) //&section.Type, &section.RelationId
// 		if err != nil {
// 			return nil, err
// 		}
// 		sections = append(sections, &section)
// 	}

// 	relationsQuery := `
//     SELECT
//         id,
//         label,
//         "order",
//         type,
//         relation_id
//     FROM
//         tab
//     WHERE
//         layout_id = $1 AND type = 'relation'
// `
// 	relationsRows, err := l.db.Query(ctx, relationsQuery, resp.Id)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer relationsRows.Close()

// 	var relations []*nb.RelationForGetAll
// 	for relationsRows.Next() {
// 		var relation nb.RelationForGetAll
// 		err := relationsRows.Scan(&resp.Id, &relation.Title, &relation.Type, &relation.Id)
// 		if err != nil {
// 			return nil, err
// 		}
// 		relations = append(relations, &relation)
// 	}

// 	return resp, nil
// }

func (l layoutRepo) GetSingle(ctx context.Context, req *nb.GetSingleLayoutRequest) (*nb.LayoutResponse, error) {
	// l.db := psqlpool.Get(req.ProjectId)
	// defer l.db.Close()

	resp := &nb.LayoutResponse{}

	if req.TableId == "" {
		tableQuery := `
			SELECT
				id
			FROM "table" WHERE "slug" = $1
		`
		var tableID string
		err := l.db.QueryRow(ctx, tableQuery, req.TableSlug).Scan(&tableID)
		if err != nil {
			return resp, err
		}
		req.TableId = tableID
	}

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
			summary_fields,
			attributes,
			table_id,
			menu_id
		FROM layout
		WHERE table_id = $1 AND menu_id = $2
	`

	row := l.db.QueryRow(ctx, layoutQuery, req.TableId, req.MenuId)
	err := row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.SummaryFields, &resp.Attributes, &resp.TableId, &resp.MenuId)
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
					summary_fields,
					attributes,
					table_id,
					menu_id
				FROM layout
				WHERE table_id = $1 AND is_default = true
			`
			row := l.db.QueryRow(ctx, layoutQuery, req.TableId)
			err = row.Scan(&resp.Id, &resp.Label, &resp.Order, &resp.Type, &resp.Icon, &resp.IsDefault, &resp.IsModal, &resp.IsVisibleSection, &resp.SummaryFields, &resp.Attributes, &resp.TableId, &resp.MenuId)
			if err != nil {
				return resp, err
			}
		} else {
			return resp, err
		}
	}

	table, err := helper.TableVer(ctx, l.db, req.TableId, req.TableSlug)
	if err != nil {
		return resp, err
	}
	req.TableId = table["id"].(string)
	req.TableSlug = table["slug"].(string)

	if len(resp.SummaryFields) > 0 {
		for _, fieldReq := range resp.SummaryFields {
			field := &nb.Field2{}

			if strings.Contains(fieldReq.Id, "#") {
				field.Id = fieldReq.Id
				field.Label = fieldReq.Label
				field.Type = fieldReq.RelationType
				relationID := strings.Split(fieldReq.Id, "#")[1]

				var fieldResp = &nb.Field{}
				err := l.db.QueryRow(ctx, "SELECT slug, required FROM field WHERE relation_id = $1 AND table_id = $2", relationID, req.TableId).Scan(&fieldResp.Slug, &fieldResp.Required)
				if err != nil {
					if err != pgx.ErrNoRows {
						return resp, err
					}
					continue
				}
				field.Slug = fieldResp.Slug
				field.Required = fieldResp.Required

				var view_of_relation = &nb.View{}
				err = l.db.QueryRow(ctx, "SELECT relation_table_slug, view_fields FROM view WHERE relation_id = $1 AND relation_table_slug = $2", relationID, req.TableSlug).Scan(&view_of_relation.RelationTableSlug, &view_of_relation.ViewFields)
				if err != nil && err != pgx.ErrNoRows {
					return resp, err
				}

				viewFieldIds := view_of_relation.ViewFields
				if view_of_relation.RelationTableSlug != "" && len(view_of_relation.ViewFields) > 0 {
					viewFieldIds = view_of_relation.ViewFields
				}

				for _, fieldID := range viewFieldIds {
					var field = &nb.Field{}
					err := l.db.QueryRow(ctx, "SELECT slug, enable_multilanguage FROM field WHERE id = $1", fieldID).Scan(&field.Slug, &field.EnableMultilanguage)
					if err != nil {
						if err != pgx.ErrNoRows {
							return resp, err
						}
						continue
					}

					var isEditable bool
					err = l.db.QueryRow(ctx, "SELECT is_editable FROM view WHERE relation_id = $1", relationID).Scan(&isEditable)
					if err != nil && err != pgx.ErrNoRows {
						return resp, err
					}
				}
			}
			var autofillFields []*nb.AutoFilter

			tableFieldsQuery := `
		SELECT id, slug, autofill_table, autofill_field, automatic
		FROM field
		WHERE table_id = $1
	`
			rows, err := l.db.Query(ctx, tableFieldsQuery, req.TableId)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			for rows.Next() {
				field := &nb.Field{}
				if err := rows.Scan(&field.Id, &field.Slug, &field.AutofillTable, &field.AutofillField, &field.Automatic); err != nil {
					return nil, err
				}

				autoFillTable := field.AutofillTable
				splitedAutoFillTable := strings.Split(autoFillTable, "#")
				if len(splitedAutoFillTable) > 1 {
					autoFillTable = splitedAutoFillTable[0]
				}

				if field.AutofillField != "" && autoFillTable != "" && autoFillTable == strings.Split(fieldReq.Id, "#")[0] {
					autofill := &nb.AutoFilter{
						FieldFrom: field.AutofillField,
						FieldTo:   field.Slug,
						// Automatic: field.Automatic,
					}
					if field.Slug == splitedAutoFillTable[1] {
						_ = append(autofillFields, autofill)
					}
				}
			}
			if err := rows.Err(); err != nil {
				return nil, err
			}

		}

	}

	tabsQuery := `
	SELECT
		id,
		label,
		"order",
		type,
		relation_id
	FROM tab
	WHERE layout_id = $1
`

	rows, err := l.db.Query(ctx, tabsQuery, resp.Id)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	var tabs []*nb.TabResponse

	for rows.Next() {
		var tab nb.TabResponse
		err := rows.Scan(&tab.Id, &tab.Label, &tab.Order, &tab.Type, &tab.RelationId)
		if err != nil {
			return resp, err
		}
		tabs = append(tabs, &tab)
	}

	for _, tab := range tabs {
		if tab.Type == "section" {
			sectionsQuery := `
			SELECT
				id,
				label,
				"order"
			FROM
				tab
			WHERE
				layout_id = $1 AND type = 'section' AND relation_id = $2
		`
			rows, err := l.db.Query(ctx, sectionsQuery, resp.Id, tab.Id)
			if err != nil {
				return resp, err
			}
			defer rows.Close()

			var sectionResponses []*nb.SectionResponse
			for rows.Next() {
				var section nb.SectionResponse
				err := rows.Scan(&section.Id, &section.Label, &section.Order)
				if err != nil {
					return resp, err
				}
				sectionResponses = append(sectionResponses, &section)
			}
			tab.Sections = sectionResponses
		} else if tab.Type == "relation" && tab.RelationId != "" {
			relationQuery := `
			SELECT
				id,
				label,
				"order",
				type
			FROM
				tab
			WHERE
				layout_id = $1 AND type = 'relation' AND relation_id = $2
		`
			rows, err := l.db.Query(ctx, relationQuery, resp.Id, tab.RelationId)
			if err != nil {
				return resp, err
			}
			defer rows.Close()

			var relation nb.RelationForSection
			if rows.Next() {
				err := rows.Scan(&relation.Id, &relation.Title, &relation.Type)
				if err != nil {
					return resp, err
				}
				tab.Relation = &relation
			}
		}
	}

	resp.Tabs = tabs

	return resp, nil
}
