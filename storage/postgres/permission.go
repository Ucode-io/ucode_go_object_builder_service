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
			AddFilter:       "Yes",
			FieldFilter:     "Yes",
			FixColumn:       "Yes",
			TabGroup:        "Yes",
			Columns:         "Yes",
			Group:           "Yes",
			ExcelMenu:       "Yes",
			SearchButton:    "Yes",
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
		values = append(values, fmt.Sprintf("('%v', '%v', '%v', '%v', %v, %v, '%v', '%v', '%v', '%v', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s', '%s')",
			v.Read, v.Write, v.Update,
			v.Delete, v.IsHaveCondition, v.IsPublic,
			v.RoleID, v.TableSlug, v.LanguageBtn,
			v.Automation, v.Settings, v.ShareModal,
			v.ViewCreate, v.PDFAction, v.AddField,
			v.AddFilter, v.FieldFilter, v.FixColumn,
			v.TabGroup, v.Columns, v.Group, v.ExcelMenu,
			v.SearchButton,
		))
	}

	query = fmt.Sprintf(`
		INSERT INTO record_permission (
			read, 
			write, 
			update, 
			delete, 
			is_have_condition, 
			is_public, 
			role_id,
			table_slug, 
			language_btn, 
			automation, 
			settings, 
			share_modal, 
			view_create, 
			pdf_action, 
			add_field,
			add_filter,
			field_filter,
			fix_column,
			tab_group,
			columns,
			"group",
			excel_menu,
			search_button
		) VALUES %v
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
			add_field = EXCLUDED.add_field,
			add_filter = EXCLUDED.add_filter,
			field_filter = EXCLUDED.field_filter,
			fix_column = EXCLUDED.fix_column,
			tab_group = EXCLUDED.tab_group,
			columns = EXCLUDED.columns,
			"group" = EXCLUDED."group",
			excel_menu = EXCLUDED.excel_menu,
			search_button = EXCLUDED.search_button
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

	if len(values) > 0 {
		query := fmt.Sprintf(`
			INSERT INTO view_permission (guid, view, edit, delete, view_id, role_id)
			VALUES %s 
			ON CONFLICT (view_id, role_id) DO UPDATE
			SET
				guid = EXCLUDED.guid,
				view = EXCLUDED.view,
				edit = EXCLUDED.edit,
				delete = EXCLUDED.delete
		`, strings.Join(values, ", "))

		_, err := conn.Exec(context.Background(), query)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *permissionRepo) GetListWithRoleAppTablePermissions(ctx context.Context, req *nb.GetListWithRoleAppTablePermissionsRequest) (resp *nb.GetListWithRoleAppTablePermissionsResponse, err error) {

	conn := psqlpool.Get(req.GetProjectId())

	var (
		role               models.Role
		FieldPermissions   []nb.RoleWithAppTablePermissions_Table_FieldPermission
		fieldPermissionMap = make(map[string]nb.RoleWithAppTablePermissions_Table_FieldPermission)
		ViewPermissions    []nb.RoleWithAppTablePermissions_Table_ViewPermission
		// AutomaticFilters     nb.RoleWithAppTablePermissions_Table_AutomaticFilterWithMethod
		ActionPermissions   []nb.RoleWithAppTablePermissions_Table_ActionPermission
		tableViewPermission []models.TableViewPermission
		tables              []nb.RoleWithAppTablePermissions_Table
		response            nb.RoleWithAppTablePermissions
	)

	query := `SELECT guid, name, project_id, COALESCE(client_platform_id::text, ''), COALESCE(client_type_id::text, ''), is_system FROM role WHERE guid = $1`

	err = conn.QueryRow(ctx, query, req.GetRoleId()).Scan(&role.Guid, &role.Name, &role.ProjectId, &role.ClientPlatformId, &role.ClientTypeId, &role.IsSystem)
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}

	roleCopy := role

	queryGetTables := `
		SELECT
			t.id,
			t.slug,
			t.label,
			t.show_in_menu,
			t.is_changed,
			t.icon,
			t.attributes,
			rp.guid,
			COALESCE(rp.read, 'No') AS read,
    		COALESCE(rp.write, 'No') AS write,
    		COALESCE(rp.update, 'No') AS update,
    		COALESCE(rp.delete, 'No') AS delete,
    		COALESCE(rp.is_public, false) AS is_public,
    		COALESCE(rp.is_have_condition, false) AS is_have_condition_other,
    		COALESCE(rp.view_create, 'No') AS view_create,
    		COALESCE(rp.share_modal, 'No') AS share_modal,
    		COALESCE(rp.settings, 'No') AS settings,
    		COALESCE(rp.automation, 'No') AS automation,
    		COALESCE(rp.language_btn, 'No') AS language_btn,
    		COALESCE(rp.pdf_action, 'No') AS pdf_action,
    		COALESCE(rp.add_field, 'No') AS add_field,
    		COALESCE(rp."add_filter", 'Yes') AS add_filter,
    		COALESCE(rp."field_filter", 'Yes') AS field_filter,
    		COALESCE(rp."fix_column", 'Yes') AS fix_column,
    		COALESCE(rp."columns", 'Yes') AS columns,
    		COALESCE(rp."group", 'Yes') AS group,
    		COALESCE(rp."excel_menu", 'Yes') AS excel_menu,
    		COALESCE(rp."tab_group", 'Yes') AS tab_group,
			COALESCE(rp."search_button", 'Yes') AS search_button
		FROM "table" t
		LEFT JOIN record_permission rp ON t.slug = rp.table_slug AND rp.role_id = $1
		WHERE t.id NOT IN (SELECT unnest($2::uuid[]))
	`
	rows, err := conn.Query(ctx, queryGetTables, req.RoleId, pq.Array(config.STATIC_TABLE_IDS))
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var table = nb.RoleWithAppTablePermissions_Table{
			RecordPermissions: &nb.RoleWithAppTablePermissions_Table_RecordPermission{},
			CustomPermission:  &nb.RoleWithAppTablePermissions_Table_CustomPermission{},
			Attributes:        structpb.NewNullValue().GetStructValue(),
		}
		attributes := []byte{}

		guid := sql.NullString{}

		err = rows.Scan(
			&table.Id,
			&table.Slug,
			&table.Label,
			&table.ShowInMenu,
			&table.IsChanged,
			&table.Icon,
			&attributes,
			&guid,
			&table.RecordPermissions.Read,
			&table.RecordPermissions.Write,
			&table.RecordPermissions.Update,
			&table.RecordPermissions.Delete,
			&table.RecordPermissions.IsPublic,
			&table.RecordPermissions.IsHaveCondition,
			&table.CustomPermission.ViewCreate,
			&table.CustomPermission.ShareModal,
			&table.CustomPermission.Settings,
			&table.CustomPermission.Automation,
			&table.CustomPermission.LanguageBtn,
			&table.CustomPermission.PdfAction,
			&table.CustomPermission.AddField,
			&table.CustomPermission.AddFilter,
			&table.CustomPermission.FieldFilter,
			&table.CustomPermission.FixColumn,
			&table.CustomPermission.Columns,
			&table.CustomPermission.Group,
			&table.CustomPermission.ExcelMenu,
			&table.CustomPermission.TabGroup,
			&table.CustomPermission.SearchButton,
		)
		if err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}
		var attrStruct *structpb.Struct
		if err := json.Unmarshal(attributes, &attrStruct); err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}

		table.Attributes = attrStruct

		tables = append(tables, table)
	}

	queryFieldPermission := `
		SELECT
			"guid",
			"label",
			"table_slug",
			"field_id",
			"edit_permission",
			"view_permission"
		FROM "field_permission" WHERE role_id = $1`

	rowsFieldPermission, err := conn.Query(ctx, queryFieldPermission, req.GetRoleId())
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}
	defer rowsFieldPermission.Close()
	for rowsFieldPermission.Next() {
		fp := nb.RoleWithAppTablePermissions_Table_FieldPermission{}
		err = rowsFieldPermission.Scan(
			&fp.Guid,
			&fp.Label,
			&fp.TableSlug,
			&fp.FieldId,
			&fp.EditPermission,
			&fp.ViewPermission,
		)
		if err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}
		FieldPermissions = append(FieldPermissions, fp)
		fieldPermissionMap[fp.FieldId] = fp
	}

	queryViewRelationPermission := `
	SELECT 
    	COALESCE(guid::text, ''),
    	COALESCE(label, ''),
    	COALESCE(relation_id::text, ''),
    	COALESCE(table_slug, ''),
    	COALESCE(view_permission, false),
    	COALESCE(create_permission, false),
    	COALESCE(edit_permission, false),
    	COALESCE(delete_permission, false)
	FROM view_relation_permission
	WHERE role_id = $1;
	`

	rowsViewRelationPermission, err := conn.Query(ctx, queryViewRelationPermission, req.GetRoleId())
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}
	defer rowsViewRelationPermission.Close()
	for rowsViewRelationPermission.Next() {
		viewRelationPermission := nb.RoleWithAppTablePermissions_Table_ViewPermission{}

		err = rowsViewRelationPermission.Scan(
			&viewRelationPermission.Guid,
			&viewRelationPermission.Label,
			&viewRelationPermission.RelationId,
			&viewRelationPermission.TableSlug,
			&viewRelationPermission.ViewPermission,
			&viewRelationPermission.CreatePermission,
			&viewRelationPermission.EditPermission,
			&viewRelationPermission.DeletePermission,
		)
		if err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}

		ViewPermissions = append(ViewPermissions, viewRelationPermission)
	}
	queryViewPermission := `
	  	SELECT 
			vp.guid,
			v.table_slug,
			vp.view,
			vp.view_id,
			vp.edit,
			vp.delete
		FROM view AS v
		LEFT JOIN view_permission AS vp ON v.id = vp.view_id
        WHERE vp.role_id = $1`

	rowsViewPermission, err := conn.Query(ctx, queryViewPermission, req.RoleId)
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}
	defer rowsViewPermission.Close()

	for rowsViewPermission.Next() {
		var viewPermission models.TableViewPermission

		err = rowsViewPermission.Scan(
			&viewPermission.Guid,
			&viewPermission.TableSlug,
			&viewPermission.View,
			&viewPermission.ViewId,
			&viewPermission.Edit,
			&viewPermission.Delete,
		)
		if err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}

		tableViewPermission = append(tableViewPermission, viewPermission)
	}

	queryActionPermission := `
		SELECT 
			ap.guid,
			ap.custom_event_id,
			ap.permission,
			ap.label,
			ap.table_slug
		FROM custom_event AS ce
		LEFT JOIN action_permission AS ap ON ce.id = ap.custom_event_id  
		WHERE ap.role_id = $1
	`

	rowsActionPermission, err := conn.Query(ctx, queryActionPermission, req.RoleId)
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}

	defer rowsActionPermission.Close()

	for rowsActionPermission.Next() {
		var actionPermission nb.RoleWithAppTablePermissions_Table_ActionPermission

		err = rowsActionPermission.Scan(
			&actionPermission.Guid,
			&actionPermission.CustomEventId,
			&actionPermission.Permission,
			&actionPermission.Label,
			&actionPermission.TableSlug,
		)
		if err != nil {
			return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
		}
		ActionPermissions = append(ActionPermissions, actionPermission)
	}

	fields := make(map[string][]nb.RoleWithAppTablePermissions_Table_FieldPermission)

	for _, fieldPermission := range FieldPermissions {

		if _, ok := fields[fieldPermission.TableSlug]; !ok {
			fields[fieldPermission.TableSlug] = []nb.RoleWithAppTablePermissions_Table_FieldPermission{fieldPermission}
		} else {
			fields[fieldPermission.TableSlug] = append(fields[fieldPermission.TableSlug], fieldPermission)
		}
	}

	view_relation_permission := make(map[string][]nb.RoleWithAppTablePermissions_Table_ViewPermission)

	for _, viewPermission := range ViewPermissions {

		if _, ok := view_relation_permission[viewPermission.TableSlug]; !ok {
			view_relation_permission[viewPermission.TableSlug] = []nb.RoleWithAppTablePermissions_Table_ViewPermission{viewPermission}
		} else {
			view_relation_permission[viewPermission.TableSlug] = append(view_relation_permission[viewPermission.TableSlug], viewPermission)
		}
	}

	table_view_permission := make(map[string][]models.TableViewPermission)

	for _, tableViewPermission := range tableViewPermission {

		if _, ok := table_view_permission[tableViewPermission.TableSlug]; !ok {
			table_view_permission[tableViewPermission.TableSlug] = []models.TableViewPermission{tableViewPermission}
		} else {
			table_view_permission[tableViewPermission.TableSlug] = append(table_view_permission[tableViewPermission.TableSlug], tableViewPermission)
		}
	}

	actionPermission := make(map[string][]*nb.RoleWithAppTablePermissions_Table_ActionPermission)

	for _, el := range ActionPermissions {
		if el.GetGuid() != "" && actionPermission[el.TableSlug] == nil {
			actionPermission[el.TableSlug] = []*nb.RoleWithAppTablePermissions_Table_ActionPermission{&el}
		} else if el.Guid != "" {
			actionPermission[el.TableSlug] = append(actionPermission[el.TableSlug], &el)
		}
	}

	var tablesList []*nb.RoleWithAppTablePermissions_Table

	for _, table := range tables {
		tableCopy := nb.RoleWithAppTablePermissions_Table{

			Id:                table.Id,
			Slug:              table.Slug,
			Label:             table.Label,
			RecordPermissions: table.RecordPermissions,
			CustomPermission: &nb.RoleWithAppTablePermissions_Table_CustomPermission{
				ViewCreate:   table.CustomPermission.ViewCreate,
				ShareModal:   table.CustomPermission.ShareModal,
				Settings:     table.CustomPermission.Settings,
				Automation:   table.CustomPermission.Automation,
				LanguageBtn:  table.CustomPermission.LanguageBtn,
				PdfAction:    table.CustomPermission.PdfAction,
				AddField:     table.CustomPermission.AddField,
				DeleteAll:    table.CustomPermission.DeleteAll,
				AddFilter:    table.CustomPermission.AddFilter,
				FieldFilter:  table.CustomPermission.FieldFilter,
				FixColumn:    table.CustomPermission.FixColumn,
				TabGroup:     table.CustomPermission.TabGroup,
				Columns:      table.CustomPermission.Columns,
				Group:        table.CustomPermission.Group,
				ExcelMenu:    table.CustomPermission.ExcelMenu,
				SearchButton: table.CustomPermission.SearchButton,
			},
		}

		// If record_permissions is nil, set default permissions
		if tableCopy.RecordPermissions == nil {
			tableCopy.RecordPermissions = &nb.RoleWithAppTablePermissions_Table_RecordPermission{
				Read:            "No",
				Write:           "No",
				Delete:          "No",
				Update:          "No",
				IsHaveCondition: false,
				IsPublic:        false,
			}
		}

		// Retrieve field permissions for the table
		tableFields := fields[table.Slug]
		tableCopy.FieldPermissions = []*nb.RoleWithAppTablePermissions_Table_FieldPermission{}

		// Iterate over fields
		for _, field := range tableFields {
			var fieldPermission nb.RoleWithAppTablePermissions_Table_FieldPermission
			// Check if field has permissions
			if field.GetGuid() != "" {
				temp := field
				fieldPermission = nb.RoleWithAppTablePermissions_Table_FieldPermission{
					FieldId:        temp.FieldId,
					TableSlug:      table.Slug,
					ViewPermission: temp.ViewPermission,
					EditPermission: temp.EditPermission,
					Label:          field.Label,
					Attributes:     field.Attributes,
				}
			} else {
				fieldPermission = nb.RoleWithAppTablePermissions_Table_FieldPermission{
					FieldId:        field.FieldId,
					TableSlug:      table.Slug,
					ViewPermission: false,
					EditPermission: false,
					Label:          field.Label,
					Guid:           "",
					Attributes:     field.Attributes,
				}
			}
			tableCopy.FieldPermissions = append(tableCopy.FieldPermissions, &fieldPermission)
		}

		// Assuming viewPermission is a map with table slug as key and slice of view permissions as value

		// Iterate over tableRelationViews
		for _, el := range view_relation_permission[table.Slug] {
			var viewPermissionEntry nb.RoleWithAppTablePermissions_Table_ViewPermission

			if el.GetGuid() != "" {
				temp := el
				viewPermissionEntry = nb.RoleWithAppTablePermissions_Table_ViewPermission{
					Guid:             temp.Guid,
					RelationId:       temp.RelationId,
					TableSlug:        temp.TableSlug,
					ViewPermission:   temp.ViewPermission,
					EditPermission:   temp.EditPermission,
					CreatePermission: temp.CreatePermission,
					DeletePermission: temp.DeletePermission,
					Label:            el.Label,
					Attributes:       el.Attributes,
				}
			} else {
				viewPermissionEntry = nb.RoleWithAppTablePermissions_Table_ViewPermission{
					Guid:             "",
					RelationId:       el.RelationId,
					TableSlug:        el.TableSlug,
					ViewPermission:   false,
					EditPermission:   false,
					CreatePermission: false,
					DeletePermission: false,
					Label:            el.Label,
					Attributes:       el.Attributes,
				}
			}
			tableCopy.ViewPermissions = append(tableCopy.ViewPermissions, &viewPermissionEntry)
		}

		for _, el := range table_view_permission[table.Slug] {
			var tableViewPermissionEntry nb.RoleWithAppTablePermissions_Table_TableViewPermission

			if el.Guid != "" {
				temp := el
				tableViewPermissionEntry = nb.RoleWithAppTablePermissions_Table_TableViewPermission{
					Guid:       temp.Guid,
					View:       temp.View,
					Edit:       temp.Edit,
					Delete:     temp.Delete,
					ViewId:     temp.ViewId,
					Attributes: el.Attributes,
				}
			} else {
				tableViewPermissionEntry = nb.RoleWithAppTablePermissions_Table_TableViewPermission{
					Guid:       "",
					View:       false,
					Edit:       false,
					Delete:     false,
					ViewId:     el.ViewId,
					Attributes: el.Attributes,
				}
			}
			tableCopy.TableViewPermissions = append(tableCopy.TableViewPermissions, &tableViewPermissionEntry)
		}

		if actionPermission != nil && actionPermission[table.Slug] != nil {
			tableCopy.ActionPermissions = actionPermission[table.Slug]
		} else {
			tableCopy.ActionPermissions = []*nb.RoleWithAppTablePermissions_Table_ActionPermission{}
		}

		tablesList = append(tablesList, &tableCopy)
	}

	queryGlobalPermission := `
	SELECT
    	guid,
    	menu_button,
    	chat,
    	settings_button,
    	project_settings_button,
    	profile_settings_button,
    	menu_setting_button,
    	redirects_button,
    	api_keys_button,
    	environments_button,
    	projects_button,
    	version_button,
    	project_button,
    	sms_button
	FROM global_permission
	WHERE role_id = $1
	`

	globalPermission := nb.GlobalPermission{}

	err = conn.QueryRow(ctx, queryGlobalPermission, req.GetRoleId()).Scan(
		&globalPermission.Id,
		&globalPermission.MenuButton,
		&globalPermission.Chat,
		&globalPermission.SettingsButton,
		&globalPermission.ProjectSettingsButton,
		&globalPermission.ProfileSettingsButton,
		&globalPermission.MenuSettingButton,
		&globalPermission.RedirectsButton,
		&globalPermission.ApiKeysButton,
		&globalPermission.EnvironmentsButton,
		&globalPermission.ProjectsButton,
		&globalPermission.VersionButton,
		&globalPermission.ProjectButton,
		&globalPermission.SmsButton,
	)
	if err != nil {
		return &nb.GetListWithRoleAppTablePermissionsResponse{}, err
	}

	response.ProjectId = req.GetProjectId()
	response.Guid = roleCopy.Guid
	response.ClientPlatformId = roleCopy.ClientPlatformId
	response.ClientTypeId = roleCopy.ClientTypeId
	response.Name = roleCopy.Name
	response.GlobalPermission = &globalPermission
	response.Tables = tablesList

	return &nb.GetListWithRoleAppTablePermissionsResponse{
		Data: &response,
	}, nil
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

	query := `UPDATE "role" SET "name" = $1 WHERE guid = $2`

	_, err = tx.Exec(ctx, query, req.Data.Name, req.Data.Guid)
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
		is_public = $6,

		search_button = $7,
		pdf_action = $8,
		add_field = $9,
		language_btn = $10,
		view_create = $11,
		automation = $12,
		settings = $13,
		share_modal = $14,
		add_filter = $15,
		field_filter = $16,
		fix_column = $17,
		tab_group = $18,
		columns = $19,
		"group" = $20,
		excel_menu = $21
	WHERE table_slug = $1 AND role_id = $22
	`

	recordPermissionInsert := `INSERT INTO "record_permission" (
		table_slug, 
		role_id, 
		read, 
		write, 
		update, 
		delete, 
		is_public,
		search_button,
		pdf_action,
		add_field,
		language_btn,
		view_create,
		automation,
		settings,
		share_modal,
		add_filter,
		field_filter,
		fix_column,
		tab_group,
		columns,
		"group",
		excel_menu,
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`

	fieldPermission := `UPDATE "field_permission" SET
		edit_permission = $2,
		view_permission = $3
	WHERE field_id = $1 AND role_id = $4
	`

	fieldPermissionInsert := `INSERT INTO "field_permission" (field_id, role_id, edit_permission, view_permission) VALUES ($1,$2,$3,$4)`

	viewPermission := `UPDATE "view_permission" SET
		view = $2,
		edit = $3,
		delete = $4
	WHERE guid = $1`

	for _, table := range req.Data.Tables {
		rp := table.RecordPermissions
		cp := table.CustomPermission
		rec, err := tx.Exec(ctx, recordPermission, table.Slug,
			rp.Read,
			rp.Write,
			rp.Update,
			rp.Delete,
			rp.IsPublic,

			cp.SearchButton,
			cp.PdfAction,
			cp.AddField,
			cp.LanguageBtn,
			cp.ViewCreate,
			cp.Automation,
			cp.Settings,
			cp.ShareModal,
			cp.AddFilter,
			cp.FieldFilter,
			cp.FixColumn,
			cp.TabGroup,
			cp.Columns,
			cp.Group,
			cp.ExcelMenu,

			req.Data.Guid)
		if err != nil {
			return err
		}
		if rec.RowsAffected() == 0 {
			_, err := tx.Exec(ctx, recordPermissionInsert, table.Slug, req.Data.Guid,
				rp.Read,
				rp.Write,
				rp.Update,
				rp.Delete,

				cp.SearchButton,
				cp.PdfAction,
				cp.AddField,
				cp.LanguageBtn,
				cp.ViewCreate,
				cp.Automation,
				cp.Settings,
				cp.ShareModal,
				cp.AddFilter,
				cp.FieldFilter,
				cp.FixColumn,
				cp.TabGroup,
				cp.Columns,
				cp.Group,
				cp.ExcelMenu,

				rp.IsPublic)
			if err != nil {
				return err
			}
		}

		for _, fp := range table.FieldPermissions {

			fip, err := tx.Exec(ctx, fieldPermission, fp.FieldId, fp.EditPermission, fp.ViewPermission, req.Data.Guid)
			if err != nil {
				return err
			}

			if fip.RowsAffected() == 0 {
				_, err := tx.Exec(ctx, fieldPermissionInsert, fp.FieldId, req.Data.Guid, fp.EditPermission, fp.ViewPermission)
				if err != nil {
					return err
				}
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

	if len(values) != 0 {
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
	}

	return nil
}

func (p *permissionRepo) GetPermissionsByTableSlug(ctx context.Context, req *nb.GetPermissionsByTableSlugRequest) (resp *nb.GetPermissionsByTableSlugResponse, err error) {

	conn := psqlpool.Get(req.GetProjectId())

	currentUserPermission, err := getTablePermission(conn, req.CurrentRoleId, req.TableSlug)
	if err != nil {
		return &nb.GetPermissionsByTableSlugResponse{}, err
	}

	if req.RoleId == "" || req.RoleId == req.CurrentRoleId {
		return &nb.GetPermissionsByTableSlugResponse{
			CurrentUserPermission: &nb.UpdatePermissionsRequest{
				Table:     currentUserPermission,
				ProjectId: req.ProjectId,
			},
		}, nil
	}

	selectedUserPermission, err := getTablePermission(conn, req.RoleId, req.TableSlug)
	if err != nil {
		return &nb.GetPermissionsByTableSlugResponse{}, err
	}

	return &nb.GetPermissionsByTableSlugResponse{
		CurrentUserPermission: &nb.UpdatePermissionsRequest{
			Table:     currentUserPermission,
			ProjectId: req.ProjectId,
		},
		SelectedUserPermission: &nb.UpdatePermissionsRequest{
			Table:     selectedUserPermission,
			ProjectId: req.ProjectId,
		},
	}, nil
}

func getTablePermission(conn *pgxpool.Pool, roleId, tableSlug string) (*nb.UpdatePermissionsRequest_Table, error) {

	query := `
	SELECT
			t.id,
			t.slug,
			t.label,
			jsonb_build_object(
				'guid', rp.guid,
				'read', rp.read,
				'write', rp.write,
				'update', rp.update,
				'delete', rp.delete
			) as record_permissions,
			COALESCE(
				jsonb_agg(
					jsonb_build_object(
						'field_id', fp.field_id,
						'view_permission', fp.view_permission,
						'edit_permission', fp.edit_permission,
						'label', fp.label,
						'table_slug', fp.table_slug
				)) FILTER (WHERE fp.field_id IS NOT NULL), '[]'::jsonb
			) as field_permissions,
			COALESCE(
				jsonb_agg(
					jsonb_build_object(
						'guid', ap.guid,
						'custom_event_id', ap.custom_event_id,
						'permission', ap.permission,
						'label', ap.label,
						'table_slug', ap.table_slug
					)) FILTER (WHERE ap.guid IS NOT NULL), '[]'::jsonb
			) as action_permissions
	FROM "table" t 
	LEFT JOIN record_permission rp ON rp.table_slug = t.slug AND rp.role_id = $1
	LEFT JOIN field_permission fp ON fp.table_slug = t.slug AND fp.role_id = $1
	LEFT JOIN action_permission ap ON ap.table_slug = t.slug AND ap.role_id = $1
	WHERE t.slug = $2 GROUP BY t.id, rp.guid`

	var (
		table = &nb.UpdatePermissionsRequest_Table{
			RecordPermissions: &nb.UpdatePermissionsRequest_Table_RecordPermission{},
		}
		recordPermissions = []byte{}
		fieldPermissions  = []byte{}
		actionPermissions = []byte{}
	)

	err := conn.QueryRow(context.Background(), query, roleId, tableSlug).Scan(
		&table.Id,
		&table.Slug,
		&table.Label,
		&recordPermissions,
		&fieldPermissions,
		&actionPermissions,
	)
	if err != nil {
		return &nb.UpdatePermissionsRequest_Table{}, err
	}

	if err := json.Unmarshal(recordPermissions, &table.RecordPermissions); err != nil {
		return &nb.UpdatePermissionsRequest_Table{}, err
	}
	if err := json.Unmarshal(fieldPermissions, &table.FieldPermissions); err != nil {
		return &nb.UpdatePermissionsRequest_Table{}, err
	}
	if err := json.Unmarshal(actionPermissions, &table.ActionPermissions); err != nil {
		return &nb.UpdatePermissionsRequest_Table{}, err
	}

	return table, nil
}

func (p *permissionRepo) UpdatePermissionsByTableSlug(ctx context.Context, req *nb.UpdatePermissionsRequest) (err error) {

	conn := psqlpool.Get(req.GetProjectId())

	roleId := req.Guid

	if roleId == "" {
		return fmt.Errorf("role_id is required")
	}

	query := `SELECT COUNT(*) FROM role WHERE guid = $1`

	count := 0

	err = conn.QueryRow(ctx, query, roleId).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("role_id is required")
	}

	query = `UPDATE record_permission SET
		read = $3,
		write = $4,
		update = $5,
		delete = $6
	WHERE role_id = $1 AND table_slug = $2`

	_, err = conn.Exec(ctx, query,
		roleId,
		req.Table.Slug,
		req.Table.RecordPermissions.Read,
		req.Table.RecordPermissions.Write,
		req.Table.RecordPermissions.Update,
		req.Table.RecordPermissions.Delete,
	)
	if err != nil {
		return err
	}

	query = `UPDATE field_permission SET
		edit_permission = $3,
		view_permission = $4
	WHERE role_id = $1 AND field_id = $2`

	for _, fp := range req.Table.FieldPermissions {
		_, err := conn.Exec(ctx, query, roleId, fp.FieldId,
			fp.EditPermission,
			fp.ViewPermission,
		)
		if err != nil {
			return err
		}
	}

	query = `UPDATE action_permission SET
		permission = $3
	WHERE role_id = $1 AND custom_event_id = $2`

	for _, ap := range req.Table.ActionPermissions {
		_, err := conn.Exec(ctx, query, roleId, ap.CustomEventId,
			ap.Permission,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
