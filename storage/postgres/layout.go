package postgres

import (
	"context"
	"fmt"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
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
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return resp, err
	}
	defer tx.Rollback(ctx)
	layout := &nb.LayoutResponse{}
	if req.Id == "" {
		layout.Id = uuid.New().String()
	} else {
		layout.Id = req.Id
	}

	query := `
        INSERT INTO "layout" (
            "id", "label", "order", "type", "icon", "is_default", 
            "is_modal", "is_visible_section", "summary_fields", 
            "attributes", "table_id", "menu_id"
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        ON CONFLICT (id) DO UPDATE
        SET 
            "label" = EXCLUDED.label,
            "order" = EXCLUDED.order,
            "type" = EXCLUDED.type,
            "icon" = EXCLUDED.icon,
            "is_default" = EXCLUDED.is_default,
            "is_modal" = EXCLUDED.is_modal,
            "is_visible_section" = EXCLUDED.is_visible_section,
            "summary_fields" = EXCLUDED.summary_fields,
            "attributes" = EXCLUDED.attributes,
            "table_id" = EXCLUDED.table_id,
            "menu_id" = EXCLUDED.menu_id
    `
	err = tx.QueryRow(ctx, query,
		layout.Id, req.Label, req.Order, req.Type, req.Icon,
		req.IsDefault, req.IsModal, req.IsVisibleSection, req.SummaryFields,
		req.Attributes, req.TableId, req.MenuId,
	).Scan(&layout.Id)
	if err != nil {
		tx.Rollback(ctx)
		return nil, fmt.Errorf("error inserting layout: %w", err)

	}

	if req.IsDefault {
		_, err = tx.Exec(ctx, `
            UPDATE layout
            SET is_default = false
            WHERE table_id = $1 AND id != $2
        `, req.TableId, layout.Id)
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

	rows, err := tx.Query(ctx, "SELECT id FROM tab WHERE layout_id = $1", layout.Id)
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

		bulkWriteTab = append(bulkWriteTab, fmt.Sprintln(`
            INSERT INTO "tab" (
                "id", "label", "layout_id", "relation_id", "type",
                "order", "icon", "attributes"
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            ON CONFLICT (id) DO UPDATE
            SET
                "label" = EXCLUDED.label,
                "layout_id" = EXCLUDED.layout_id,
                "relation_id" = EXCLUDED.relation_id,
                "type" = EXCLUDED.type,
                "order" = EXCLUDED.order,
                "icon" = EXCLUDED.icon,
                "attributes" = EXCLUDED.attributes
        `, tab.Id, tab.Label, tab.LayoutId, tab.RelationId, tab.Type,
			tab.Order, tab.Icon, tab.Attributes))

		for _, section := range tab.Sections {
			if section.Id == "" {
				section.Id = uuid.New().String()
			}

			if _, ok := mapSections[section.Id]; ok {
				mapSections[section.Id] = 2
			}

			bulkWriteSection = append(bulkWriteSection, fmt.Sprintln(`
                INSERT INTO "section" (
                    "id", "tab_id", "label", "order", "icon", "attributes"
                ) VALUES ($1, $2, $3, $4, $5, $6)
                ON CONFLICT (id) DO UPDATE
                SET
                    "tab_id" = EXCLUDED.tab_id,
                    "label" = EXCLUDED.label,
                    "order" = EXCLUDED.order,
                    "icon" = EXCLUDED.icon,
                    "attributes" = EXCLUDED.attributes
            `, section.Id, tab.Id, section.Label, section.Order,
				section.Icon, section.Attributes))
		}
	}

	for _, query := range bulkWriteTab {
		_, err := tx.Exec(ctx, query)
		if err != nil {
			tx.Rollback(ctx)
			return nil, fmt.Errorf("error executing bulkWriteTab query: %w", err)
		}
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

	return &nb.LayoutResponse{}, nil

}

func (l layoutRepo) RemoveLayout(ctx context.Context, req *nb.LayoutPrimaryKey) error {

	conn := psqlpool.Get(req.ProjectId)

	tx, err := conn.Begin(ctx)
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
