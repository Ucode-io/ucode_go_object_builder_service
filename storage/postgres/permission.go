package postgres

import (
	"context"
	"encoding/json"
	"ucode/ucode_go_object_builder_service/storage"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgx/v5/pgxpool"
)

type permissionRepo struct {
	db *pgxpool.Pool
}

func NewPermissionRepo(db *pgxpool.Pool) storage.PermissionRepoI {
	return &permissionRepo{
		db: db,
	}
}

func (p *permissionRepo) GetAllMenuPermissions(ctx context.Context, req *nb.GetAllMenuPermissionsRequest) (*nb.GetAllMenuPermissionsResponse, error) {
	conn := p.db

	query := `
		SELECT 
			m."id",
			m."label",
			m."attributes",
			m."type",

			mp."write",
			mp."read",
			mp."delete",
			mp."update",
			mp."menu_settings"
		FROM "menu" m
		LEFT JOIN
			menu_permission mp ON m.id = mp."menu_id" AND mp.role_id = $1
		WHERE 
			m.parent_id = $2
		ORDER BY
			m.created_at DESC
	`

	var (
		resp = &nb.GetAllMenuPermissionsResponse{}
	)

	rows, err := conn.Query(ctx, query, req.RoleId, req.ParentId)
	if err != nil {
		return &nb.GetAllMenuPermissionsResponse{}, err
	}

	for rows.Next() {
		var (
			attributes = []byte{}
			menu       = &nb.MenuPermission{}
			permission = &nb.MenuPermission_Permission{}
		)

		err := rows.Scan(
			&menu.Id,
			&menu.Label,
			&attributes,
			&menu.Type,

			&permission.Write,
			&permission.Read,
			&permission.Delete,
			&permission.Update,
			&permission.MenuSettings,
		)
		if err != nil {
			return &nb.GetAllMenuPermissionsResponse{}, err
		}

		if err := json.Unmarshal(attributes, &menu.Attributes); err != nil {
			return &nb.GetAllMenuPermissionsResponse{}, err
		}

		menu.Permission = permission

		resp.Menus = append(resp.Menus, menu)
	}

	return resp, nil
}
