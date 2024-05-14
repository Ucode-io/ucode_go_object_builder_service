package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
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
	// conn := psqlpool.Get(req.GetProjectId())

	// if req.GetRoleId() == "" {
	// 	return errors.New("role_id is required")
	// }

	// query := `
	// 	SELECT
	// 		t.id,
	// 		t.slug,
	// 		t.label,
	// 		t.show_in_menu,
	// 		t.is_changed,
	// 		t.is_cached,
	// 		t.icon,
	// 		t.is_system,
	// 		t.attributes
	// 	FROM "table" t
	// 	LEFT JOIN record_permission rp ON t.slug = rp.table_slug AND rp.role_id = $1
	// 	WHERE t.id NOT IN (SELECT unnest($2::uuid[]))
	// `

	// rows, err := conn.Query(
	// 	ctx,
	// 	query,
	// 	req.RoleId,
	// 	pq.Array(config.STATIC_TABLE_IDS),
	// )
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// tables := []models.TablePermission{}
	// for rows.Next() {
	// 	table := models.TablePermission{}
	// 	attributes := []byte{}

	// 	err = rows.Scan(
	// 		&table.Id,
	// 		&table.Slug,
	// 		&table.Label,
	// 		&table.ShowInMenu,
	// 		&table.IsChanged,
	// 		&table.IsCached,
	// 		&table.Icon,
	// 		&table.IsSystem,
	// 		&attributes,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	var attrStruct *structpb.Struct
	// 	if err := json.Unmarshal(attributes, &attrStruct); err != nil {
	// 		return err
	// 	}
	// 	table.Attributes = attrStruct

	// 	tables = append(tables, table)
	// }

	// query = `
	// 	SELECT
	// 		f.id,
	// 		f.label,
	// 		f.table_id,
	// 		f.attributes
	// 	FROM "field" f
	// `

	// rows, err = conn.Query(
	// 	ctx,
	// 	query,
	// )
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// testFields := []models.Field{}
	// for rows.Next() {
	// 	field := models.Field{}
	// 	attributes := []byte{}

	// 	err = rows.Scan(
	// 		&field.Id,
	// 		&field.Label,
	// 		&field.TableId,
	// 		&attributes,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	var attrStruct *structpb.Struct
	// 	if err := json.Unmarshal(attributes, &attrStruct); err != nil {
	// 		return err
	// 	}
	// 	field.Attributes = attrStruct

	// 	testFields = append(testFields, field)
	// }

	// query = `
	// 	SELECT
	// 		v.id,
	// 		v.name,
	// 		v.table_slug,
	// 		v.attributes
	// 	FROM "view" v
	// `

	// rows, err = conn.Query(
	// 	ctx,
	// 	query,
	// )
	// if err != nil {
	// 	return err
	// }
	// defer rows.Close()

	// views := []models.View{}
	// for rows.Next() {
	// 	view := models.View{}
	// 	attributes := []byte{}

	// 	err = rows.Scan(
	// 		&view.Id,
	// 		&view.Name,
	// 		&view.TableSlug,
	// 		&attributes,
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	var attrStruct map[string]interface{}
	// 	if err := json.Unmarshal(attributes, &attrStruct); err != nil {
	// 		return err
	// 	}
	// 	view.Attributes = attrStruct

	// 	views = append(views, view)
	// }

	// var (
	// 	recordPermissions = make([]models.RecordPermission, 0)
	// )
	// for _, table := range tables {
	// 	IsHaveCondition := false

	// 	recordPermissionDocument := models.RecordPermission{
	// 		Read:            "Yes",
	// 		Write:           "Yes",
	// 		Update:          "Yes",
	// 		Delete:          "Yes",
	// 		IsHaveCondition: IsHaveCondition,
	// 		IsPublic:        true,
	// 		RoleID:          req.RoleId,
	// 		TableSlug:       table.Slug,
	// 		LanguageBtn:     "Yes",
	// 		Automation:      "Yes",
	// 		Settings:        "Yes",
	// 		ShareModal:      "Yes",
	// 		ViewCreate:      "Yes",
	// 		PDFAction:       "Yes",
	// 		AddField:        "Yes",
	// 		DeleteAll:       "Yes",
	// 	}
	// 	recordPermissions = append(recordPermissions, recordPermissionDocument)
	// }

	// return nil
	return nil
}
