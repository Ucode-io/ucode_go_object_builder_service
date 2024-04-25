package postgres

import (
	"context"
	"database/sql"
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
			attributes   = []byte{}
			menu         = &nb.MenuPermission{}
			permission   = &nb.MenuPermission_Permission{}
			read         = sql.NullBool{}
			write        = sql.NullBool{}
			update       = sql.NullBool{}
			delete       = sql.NullBool{}
			menuSettings = sql.NullBool{}
		)

		err := rows.Scan(
			&menu.Id,
			&menu.Label,
			&attributes,
			&menu.Type,

			&write,
			&read,
			&delete,
			&update,
			&menuSettings,
		)
		if err != nil {
			return &nb.GetAllMenuPermissionsResponse{}, err
		}

		if err := json.Unmarshal(attributes, &menu.Attributes); err != nil {
			return &nb.GetAllMenuPermissionsResponse{}, err
		}

		permission.Read = read.Bool
		permission.Write = write.Bool
		permission.Delete = delete.Bool
		permission.Update = update.Bool
		permission.MenuSettings = menuSettings.Bool
		menu.Permission = permission

		resp.Menus = append(resp.Menus, menu)
	}

	return resp, nil
}
