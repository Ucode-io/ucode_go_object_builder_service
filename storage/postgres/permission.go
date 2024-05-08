package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/pkg/errors"

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
	conn := psqlpool.Get(req.GetProjectId())

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

func (p *permissionRepo) CreateDefaultPermission(ctx context.Context, req *nb.CreateDefaultPermissionRequest) error {
	conn := psqlpool.Get(req.GetProjectId())

	if req.GetRoleId() == "" {
		return errors.New("role_id is required")
	}

	tablePipeline := `
		SELECT t.id AS id,
			t.slug AS slug,
			t.label AS label,
			t.show_in_menu AS show_in_menu,
			t.is_cached AS is_changed,
			t.icon AS icon,
			t.is_cached AS is_changed,
			t.is_system AS is_system,
			rp.* AS record_permissions,
			t.attributes AS attributes
		FROM "table" t
		LEFT JOIN (
		SELECT *
		FROM record_permissions rp
		WHERE rp.table_slug = t.slug
		AND rp.role_id = $1
		LIMIT 1
		) AS rp ON true
		WHERE t.deleted_at = '1970-01-01 18:00:00+00:00'
		AND t.id NOT IN ($2);
	`

	_, err := conn.Query(ctx, tablePipeline, req.GetRoleId(), helper.STATIC_TABLE_IDS)
	if err != nil {
		return errors.Wrap(err, "failed to create default permission")
	}

	return nil
}
