package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"google.golang.org/protobuf/types/known/structpb"
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
	conn := psqlpool.Get(req.ProjectId)

	query := `
		SELECT
			t.id,
			t.slug,
			t.label,
			t.show_in_menu,
			t.is_changed,
			t.is_cached,
			t.icon,
			t.is_system,
			t.attributes
		FROM "table" t
		LEFT JOIN record_permission rp ON t.slug = rp.table_slug AND rp.role_id = $1
		WHERE t.id NOT IN (SELECT unnest($2::uuid[]))
	`
	rows, err := conn.Query(ctx, query, req.RoleId, pq.Array(config.STATIC_TABLE_IDS))
	if err != nil {
		return err
	}
	defer rows.Close()

	tables := []models.TablePermission{}
	for rows.Next() {
		table := models.TablePermission{}
		attributes := []byte{}

		err = rows.Scan(
			&table.Id,
			&table.Slug,
			&table.Label,
			&table.ShowInMenu,
			&table.IsChanged,
			&table.IsCached,
			&table.Icon,
			&table.IsSystem,
			&attributes,
		)
		if err != nil {
			return err
		}

		var attrStruct *structpb.Struct
		if err := json.Unmarshal(attributes, &attrStruct); err != nil {
			return err
		}
		table.Attributes = attrStruct

		tables = append(tables, table)
	}

	query = `
		SELECT
			f.id,
			f.label,
			f.table_id,
			f.attributes
		FROM "field" f
	`

	rows, err = conn.Query(
		ctx,
		query,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	fields := map[string][]models.Field{}
	for rows.Next() {
		field := models.Field{}
		attributes := []byte{}

		err = rows.Scan(
			&field.Id,
			&field.Label,
			&field.TableId,
			&attributes,
		)
		if err != nil {
			return err
		}

		var attrStruct *structpb.Struct
		if err := json.Unmarshal(attributes, &attrStruct); err != nil {
			return err
		}
		field.Attributes = attrStruct

		if _, ok := fields[field.TableId]; !ok {
			fields[field.TableId] = []models.Field{field}
		} else {
			fields[field.TableId] = append(fields[field.TableId], field)
		}
	}

	query = `
		SELECT
			v.id,
			v.name,
			v.table_slug,
			v.attributes
		FROM "view" v
	`

	rows, err = conn.Query(
		ctx,
		query,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	views := map[string][]models.View{}
	for rows.Next() {
		view := models.View{}
		attributes := []byte{}

		name := sql.NullString{}

		err = rows.Scan(
			&view.Id,
			&name,
			&view.TableSlug,
			&attributes,
		)
		if err != nil {
			return err
		}

		var attrStruct map[string]interface{}
		if err := json.Unmarshal(attributes, &attrStruct); err != nil {
			return err
		}
		view.Attributes = attrStruct
		view.Name = name.String

		if _, ok := views[view.TableSlug]; !ok {
			views[view.TableSlug] = []models.View{view}
		} else {
			views[view.TableSlug] = append(views[view.TableSlug], view)
		}
	}

	var (
		recordPermissions = make([]models.RecordPermission, 0)
		fieldPermissions  = make([]models.FieldPermission, 0)
		viewPermissions   = make([]models.ViewPermission, 0)
	)

	for _, table := range tables {
		IsHaveCondition := false

		recordPermissionDocument := models.RecordPermission{
			Read:            "Yes",
			Write:           "Yes",
			Update:          "Yes",
			Delete:          "Yes",
			IsHaveCondition: IsHaveCondition,
			IsPublic:        true,
			RoleID:          req.RoleId,
			TableSlug:       table.Slug,
			LanguageBtn:     "Yes",
			Automation:      "Yes",
			Settings:        "Yes",
			ShareModal:      "Yes",
			ViewCreate:      "Yes",
			PDFAction:       "Yes",
			AddField:        "Yes",
		}
		recordPermissions = append(recordPermissions, recordPermissionDocument)

		tableFields := fields[table.Id]
		for _, tableField := range tableFields {
			fieldPermission := models.FieldPermission{
				ViewPermission: true,
				EditPermission: true,
				FieldId:        tableField.Id,
				TableSlug:      table.Slug,
				RoleId:         req.RoleId,
				Label:          tableField.Label,
				Guid:           uuid.NewString(),
			}
			fieldPermissions = append(fieldPermissions, fieldPermission)
		}

		tableViews := views[table.Slug]
		for _, view := range tableViews {
			viewPermission := models.ViewPermission{
				Guid:   uuid.NewString(),
				View:   true,
				Edit:   true,
				Delete: true,
				ViewId: view.Id,
				RoleId: req.RoleId,
			}
			viewPermissions = append(viewPermissions, viewPermission)
		}
	}

	customPermission := models.CustomPermission{
		Chat:                  true,
		MenuButton:            true,
		SettingsButton:        true,
		ProjectsButton:        true,
		EnvironmentsButton:    true,
		APIKeysButton:         true,
		RedirectsButton:       true,
		MenuSettingButton:     true,
		ProfileSettingsButton: true,
		ProjectButton:         true,
		SMSButton:             true,
		VersionButton:         true,
	}

	query = `
		INSERT INTO global_permission (role_id, chat, menu_button, settings_button, projects_button, environments_button, api_keys_button, redirects_button, menu_setting_button, profile_settings_button, project_button, sms_button, version_button)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (role_id) DO UPDATE
		SET chat = EXCLUDED.chat, menu_button = EXCLUDED.menu_button, settings_button = EXCLUDED.settings_button, projects_button = EXCLUDED.projects_button, environments_button = EXCLUDED.environments_button, api_keys_button = EXCLUDED.api_keys_button, redirects_button = EXCLUDED.redirects_button, menu_setting_button = EXCLUDED.menu_setting_button, profile_settings_button = EXCLUDED.profile_settings_button, project_button = EXCLUDED.project_button, sms_button = EXCLUDED.sms_button, version_button = EXCLUDED.version_button`

	_, err = conn.Exec(context.Background(), query,
		req.RoleId,
		customPermission.Chat,
		customPermission.MenuButton,
		customPermission.SettingsButton,
		customPermission.ProjectsButton,
		customPermission.EnvironmentsButton,
		customPermission.APIKeysButton,
		customPermission.RedirectsButton,
		customPermission.MenuSettingButton,
		customPermission.ProfileSettingsButton,
		customPermission.ProjectButton,
		customPermission.SMSButton,
		customPermission.VersionButton,
	)
	if err != nil {
		return err
	}

	query = `
		SELECT
			id
		FROM "menu" m
		LEFT JOIN 
			menu_permission mp ON m.id = mp.menu_id AND mp.role_id = $1
		ORDER BY 
			m.order
	`
	rows, err = conn.Query(ctx, query, req.RoleId)
	if err != nil {
		return err
	}
	defer rows.Close()

	menus := []models.Menu{}
	for rows.Next() {
		menu := models.Menu{}
		err := rows.Scan(&menu.Id)
		if err != nil {
			return err
		}
		menus = append(menus, menu)
	}

	menuPermissions := []models.MenuPermission{}
	for _, menu := range menus {
		menuPermission := models.MenuPermission{
			MenuID:       menu.Id,
			RoleID:       req.RoleId,
			Delete:       true,
			GUID:         uuid.NewString(),
			MenuSettings: true,
			Read:         true,
			Update:       true,
			Write:        true,
		}
		menuPermissions = append(menuPermissions, menuPermission)
	}

	values := []string{}

	for _, v := range recordPermissions {
		values = append(values, fmt.Sprintf("('%v', '%v', '%v', '%v', %v, %v, '%v', '%v', '%v', '%v', '%s', '%s', '%s', '%s', '%s')",
			v.Read, v.Write, v.Update,
			v.Delete, v.IsHaveCondition, v.IsPublic,
			v.RoleID, v.TableSlug, v.LanguageBtn,
			v.Automation, v.Settings, v.ShareModal,
			v.ViewCreate, v.PDFAction, v.AddField,
		))
	}

	query = fmt.Sprintf(`
		INSERT INTO record_permission (read, write, update, delete, is_have_condition, is_public, role_id, table_slug, language_btn, automation, settings, share_modal, view_create, pdf_action, add_field)
		VALUES %v
		ON CONFLICT (role_id, table_slug) DO UPDATE
		SET
			read = EXCLUDED.read,
			write = EXCLUDED.write,
			update = EXCLUDED.update,
			delete = EXCLUDED.delete,
			is_have_condition = EXCLUDED.is_have_condition,
			is_public = EXCLUDED.is_public,
			language_btn = EXCLUDED.language_btn,
			automation = EXCLUDED.automation,
			settings = EXCLUDED.settings,
			share_modal = EXCLUDED.share_modal,
			view_create = EXCLUDED.view_create,
			pdf_action = EXCLUDED.pdf_action,
			add_field = EXCLUDED.add_field
	`, strings.Join(values, ", "))

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	values = []string{}

	for _, v := range menuPermissions {
		values = append(values, fmt.Sprintf("('%v', '%v', %v, '%v', %v, %v, %v, %v)",
			v.MenuID, v.RoleID, v.Delete,
			v.GUID, v.MenuSettings, v.Read,
			v.Update, v.Write,
		))
	}

	query = fmt.Sprintf(`
		INSERT INTO menu_permission (menu_id, role_id, delete, guid, menu_settings, read, update, write)
		VALUES %s
		ON CONFLICT (menu_id, role_id) DO UPDATE
		SET
			delete = EXCLUDED.delete,
			guid = EXCLUDED.guid,
			menu_settings = EXCLUDED.menu_settings,
			read = EXCLUDED.read,
			update = EXCLUDED.update,
			write = EXCLUDED.write
	`, strings.Join(values, ", "))

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	values = []string{}

	for _, v := range fieldPermissions {
		label := strings.ReplaceAll(v.Label, "'", "''")

		values = append(values, fmt.Sprintf("(%v, %v, '%v', '%v', '%v', '%v', '%v')",
			v.ViewPermission, v.EditPermission, v.FieldId,
			v.TableSlug, v.RoleId, label, v.Guid,
		))
	}

	templates := helper.CreateTemplate(req.RoleId)
	for _, v := range templates {
		label := strings.ReplaceAll(v.Label, "'", "''")

		values = append(values, fmt.Sprintf("(%v, %v, '%v', '%v', '%v', '%v', '%v')",
			v.ViewPermission, v.EditPermission, v.FieldId,
			v.TableSlug, v.RoleId, label, v.Guid,
		))
	}

	query = fmt.Sprintf(`
        INSERT INTO field_permission ("view_permission", "edit_permission", "field_id", "table_slug", "role_id", "label", "guid")
        VALUES %s
        ON CONFLICT (field_id, role_id) DO UPDATE
        SET
            view_permission = EXCLUDED.view_permission,
            edit_permission = EXCLUDED.edit_permission,
            table_slug = EXCLUDED.table_slug,
            label = EXCLUDED.label,
            guid = EXCLUDED.guid
    `, strings.Join(values, ", "))

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	values = []string{}

	for _, v := range viewPermissions {
		values = append(values, fmt.Sprintf("('%v', '%v', '%v', '%v', '%v', '%v')",
			v.Guid, v.View, v.Edit, v.Delete, v.ViewId, v.RoleId,
		))
	}

	query = fmt.Sprintf(`
		INSERT INTO view_permission (guid, view, edit, delete, view_id, role_id)
		VALUES %s
		ON CONFLICT (view_id, role_id) DO UPDATE
		SET
			guid = EXCLUDED.guid,
			view = EXCLUDED.view,
			edit = EXCLUDED.edit,
			delete = EXCLUDED.delete
	`, strings.Join(values, ", "))

	_, err = conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}

func (p *permissionRepo) UpdateRoleAppTablePermissions(ctx context.Context, req *nb.UpdateRoleAppTablePermissionsRequest) error {

	conn := psqlpool.Get(req.GetProjectId())

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()

	query := `UPDATE "role" SET "name" = $1`

	_, err = tx.Exec(ctx, query, req.Data.Name)
	if err != nil {
		return err
	}

	gP := req.Data.GlobalPermission

	globalPermission := `UPDATE "global_permission" SET
		chat = $2,
		menu_button = $3,
		settings_button = $4,
		projects_button = $5,
		environments_button = $6,
		api_keys_button = $7,
		menu_setting_button = $8,
		redirects_button = $9,
		profile_settings_button = $10,
		project_settings_button = $11,
		project_button = $12,
		sms_button = $13,
		version_button = $14
	WHERE guid = $1
	`

	_, err = tx.Exec(ctx, globalPermission, gP.Id,
		gP.Chat,
		gP.MenuButton,
		gP.SettingsButton,
		gP.ProjectsButton,
		gP.EnvironmentButton,
		gP.ApiKeysButton,
		gP.MenuSettingButton,
		gP.RedirectsButton,
		gP.ProfileSettingsButton,
		gP.ProjectSettingsButton,
		gP.ProjectButton,
		gP.SmsButton,
		gP.VersionButton,
	)
	if err != nil {
		return err
	}

	recordPermission := `UPDATE "record_permission" SET 
		read = $2,
		write = $3,
		update = $4,
		delete = $5,
		is_public = $6
	WHERE guid = $1
	`

	fieldPermission := `UPDATE "field_permission" SET
		edit_permission = $2,
		view_permission = $3
	WHERE guid = $1
	`

	viewPermission := `UPDATE "view_permission" SET
		view = $2,
		edit = $3,
		delete = $4
	WHERE guid = $1`

	for _, table := range req.Data.Tables {
		rp := table.RecordPermissions
		_, err = tx.Exec(ctx, recordPermission, rp.Guid, rp.Read, rp.Write, rp.Update, rp.Delete, rp.IsPublic)
		if err != nil {
			return err
		}

		for _, fp := range table.FieldPermissions {
			_, err = tx.Exec(ctx, fieldPermission, fp.Guid, fp.EditPermission, fp.ViewPermission)
			if err != nil {
				return err
			}
		}

		for _, vP := range table.ViewPermissions {
			_, err = tx.Exec(ctx, viewPermission, vP.Guid, vP.ViewPermission, vP.EditPermission, vP.DeletePermission)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *permissionRepo) UpdateMenuPermissions(ctx context.Context, req *nb.UpdateMenuPermissionsRequest) error {
	conn := psqlpool.Get(req.ProjectId)

	values := []string{}

	for _, v := range req.Menus {
		values = append(values, fmt.Sprintf("('%v', '%v', %v, '%v', %v, %v, %v, %v)",
			v.Id, req.RoleId, v.Permission.Delete,
			uuid.NewString(), v.Permission.MenuSettings, v.Permission.Read,
			v.Permission.Update, v.Permission.Write,
		))
	}

	query := fmt.Sprintf(`
		INSERT INTO menu_permission (menu_id, role_id, delete, guid, menu_settings, read, update, write)
		VALUES %s
		ON CONFLICT (menu_id, role_id) DO UPDATE
		SET
			delete = EXCLUDED.delete,
			guid = EXCLUDED.guid,
			menu_settings = EXCLUDED.menu_settings,
			read = EXCLUDED.read,
			update = EXCLUDED.update,
			write = EXCLUDED.write
	`, strings.Join(values, ", "))

	_, err := conn.Exec(context.Background(), query)
	if err != nil {
		return err
	}

	return nil
}
