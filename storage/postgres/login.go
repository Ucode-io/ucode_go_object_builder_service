package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type loginRepo struct {
	db *pgxpool.Pool
}

func NewLoginRepo(db *pgxpool.Pool) storage.LoginRepoI {
	return &loginRepo{
		db: db,
	}
}

func (l *loginRepo) LoginData(ctx context.Context, req *nb.LoginDataReq) (resp *nb.LoginDataRes, err error) {
	conn := psqlpool.Get(req.GetResourceEnvironmentId())

	query := `
		SELECT
			"guid",
			COALESCE("project_id"::varchar, ''),
			"name",
			"self_register",
			"self_recover",
			"client_platform_ids",
			"confirm_by",
			"is_system",
			"table_slug",
			"default_page"
		FROM client_type WHERE "guid" = $1 OR "name" = $1::varchar
	`

	var (
		clientType      models.ClientType
		tableSlug       = `"user"`
		userId, roleId  string
		userFound       bool
		role            models.Role
		clientPlatform  models.ClientPlatform
		connections     = []*nb.TableClientType{}
		permissions     = []*nb.RecordPermission{}
		tableSlugNull   sql.NullString
		defaultPageNull sql.NullString
	)

	err = conn.QueryRow(ctx, query, req.ClientType).Scan(
		&clientType.Guid,
		&clientType.ProjectId,
		&clientType.Name,
		&clientType.SelfRegister,
		&clientType.SelfRecover,
		&clientType.ClientPlatformIds,
		&clientType.ConfirmBy,
		&clientType.IsSystem,
		&tableSlugNull,
		&defaultPageNull,
	)
	if err != nil {
		return &nb.LoginDataRes{
			UserFound: false,
		}, errors.Wrap(err, "error getting client type")
	}

	clientType.TableSlug = tableSlugNull.String
	clientType.DefaultPage = defaultPageNull.String

	if clientType.TableSlug != "" && clientType.TableSlug != "user" {
		tableSlug = clientType.TableSlug
	}

	query = `SELECT guid, role_id FROM ` + tableSlug + ` WHERE guid::varchar = $1 AND client_type_id::varchar = $2`

	err = conn.QueryRow(ctx, query, req.UserId, req.ClientType).Scan(
		&userId,
		&roleId,
	)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return &nb.LoginDataRes{
				UserFound: false,
			}, nil
		}
		return &nb.LoginDataRes{
			UserFound: false,
		}, errors.Wrap(err, "error getting user")
	}

	if userId != "" {
		userFound = true
	}

	query = `SELECT 
		"guid",
		"name",
		"project_id",
		COALESCE("client_platform_id"::varchar, ''),
		"client_type_id"
	FROM "role" WHERE "guid" = $1`

	err = conn.QueryRow(ctx, query, roleId).Scan(
		&role.Guid,
		&role.Name,
		&role.ProjectId,
		&role.ClientPlatformId,
		&role.ClientTypeId,
	)
	if err != nil {
		return &nb.LoginDataRes{
			UserFound: false,
		}, errors.Wrap(err, "error getting role")
	}

	query = `SELECT 
		"guid",
		"name",
		"project_id",
		"subdomain"
	FROM "client_platform" WHERE "guid" = $1
	`

	if role.ClientPlatformId != "" {
		err = conn.QueryRow(ctx, query, role.ClientPlatformId).Scan(
			&clientPlatform.Guid,
			&clientPlatform.Name,
			&clientPlatform.ProjectId,
			&clientPlatform.Subdomain,
		)
		if err != nil {
			return &nb.LoginDataRes{
				UserFound: false,
			}, errors.Wrap(err, "error getting client platform")
		}
	}

	query = `SELECT 
		"table_slug",
		"view_slug",
		"view_label",
		"icon",
		"name"
	FROM "connections" WHERE deleted_at IS NULL AND client_type_id = $1`

	rows, err := conn.Query(ctx, query, clientType.Guid)
	if err != nil {
		return &nb.LoginDataRes{}, errors.Wrap(err, "error getting connections")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			slug      sql.NullString
			viewSlug  sql.NullString
			viewLabel sql.NullString
			icon      sql.NullString
			label     sql.NullString
		)

		err = rows.Scan(
			&slug,
			&viewSlug,
			&viewLabel,
			&icon,
			&label,
		)
		if err != nil {
			return &nb.LoginDataRes{}, errors.Wrap(err, "error scanning connections")
		}

		connections = append(connections, &nb.TableClientType{
			Slug:      slug.String,
			ViewSlug:  viewSlug.String,
			ViewLabel: viewLabel.String,
			Icon:      icon.String,
			Label:     label.String,
		})
	}

	query = `SELECT 
		"guid",
		"role_id",
		"read",
		"write",
		"update",
		"delete",
		"table_slug",
		"automation",
		"language_btn",
		"settings",
		"share_modal",
		"view_create",
		"add_field",
		"pdf_action",
		add_filter,
		field_filter,
		fix_column,
		tab_group,
		columns,
		"group",
		excel_menu,
		search_button
	FROM "record_permission" WHERE role_id = $1`

	recPermissions, err := conn.Query(ctx, query, roleId)
	if err != nil {
		return &nb.LoginDataRes{}, errors.Wrap(err, "error getting record permissions")
	}
	defer recPermissions.Close()

	for recPermissions.Next() {
		permission := nb.RecordPermission{}

		err = recPermissions.Scan(
			&permission.Guid,
			&permission.RoleId,
			&permission.Read,
			&permission.Write,
			&permission.Update,
			&permission.Delete,
			&permission.TableSlug,
			&permission.Automation,
			&permission.LanguageBtn,
			&permission.Settings,
			&permission.ShareModal,
			&permission.ViewCreate,
			&permission.AddField,
			&permission.PdfAction,
			&permission.AddFilter,
			&permission.FieldFilter,
			&permission.FixColumn,
			&permission.TabGroup,
			&permission.Columns,
			&permission.Group,
			&permission.ExcelMenu,
			&permission.SearchButton,
		)
		if err != nil {
			return &nb.LoginDataRes{}, errors.Wrap(err, "error scanning record permissions")
		}

		permissions = append(permissions, &permission)
	}

	query = `
		SELECT 
			"guid",
			"chat",
			"menu_button",
			"settings_button",
			"projects_button",
			"environments_button",
			"api_keys_button",
			"menu_setting_button",
			"redirects_button",
			"profile_settings_button",
			"project_settings_button",
			"project_button",
			"sms_button",
			"version_button"
		FROM global_permission
		WHERE role_id = $1
		LIMIT 1
	`

	globalPermission := &nb.GlobalPermission{}
	err = conn.QueryRow(ctx, query, roleId).Scan(
		&globalPermission.Id,
		&globalPermission.Chat,
		&globalPermission.MenuButton,
		&globalPermission.SettingsButton,
		&globalPermission.ProjectsButton,
		&globalPermission.EnvironmentsButton,
		&globalPermission.ApiKeysButton,
		&globalPermission.MenuSettingButton,
		&globalPermission.RedirectsButton,
		&globalPermission.ProfileSettingsButton,
		&globalPermission.ProjectSettingsButton,
		&globalPermission.ProjectButton,
		&globalPermission.SmsButton,
		&globalPermission.VersionButton,
	)
	if err != nil {
		return &nb.LoginDataRes{}, errors.Wrap(err, "error getting global permission")
	}

	return &nb.LoginDataRes{
		UserFound:      userFound,
		UserId:         userId,
		LoginTableSlug: tableSlug,
		ClientType: &nb.ClientType{
			Guid:         clientType.Guid,
			Name:         clientType.Name,
			ConfirmBy:    nb.ConfirmStrategies(0),
			SelfRegister: clientType.SelfRegister,
			SelfRecover:  clientType.SelfRecover,
			ProjectId:    clientType.ProjectId,
			Tables:       connections,
			DefaultPage:  clientType.DefaultPage,
		},
		Role: &nb.Role{
			Guid:             role.Guid,
			ClientTypeId:     role.ClientTypeId,
			Name:             role.Name,
			ClientPlatformId: role.ClientPlatformId,
			ProjectId:        role.ProjectId,
		},
		ClientPlatform: &nb.ClientPlatform{
			Guid:      clientPlatform.Guid,
			Name:      clientPlatform.Name,
			ProjectId: clientPlatform.ProjectId,
			Subdomain: clientPlatform.Subdomain,
		},
		Permissions:      permissions,
		GlobalPermission: globalPermission,
	}, nil
}

func (l *loginRepo) GetConnectionOptions(ctx context.Context, req *nb.GetConnetionOptionsRequest) (resp *nb.GetConnectionOptionsResponse, err error) {
	var (
		conn       = psqlpool.Get(req.GetResourceEnvironmentId())
		options    []map[string]interface{}
		connection models.Connection
		clientType models.ClientType
		user       map[string]interface{}
	)

	query := `SELECT table_slug, field_slug, client_type_id FROM "connections" WHERE guid = $1`
	err = conn.QueryRow(ctx, query, req.ConnectionId).Scan(&connection.TableSlug, &connection.FieldSlug, &connection.ClientTypeId)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get connection")
	}

	if connection.TableSlug != "" && connection.FieldSlug != "" {
		query = `SELECT table_slug FROM client_type WHERE guid = $1`
		err = conn.QueryRow(ctx, query, connection.ClientTypeId).Scan(&clientType.TableSlug)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get client type")
		}

		tableSlug := "user"
		if clientType.TableSlug != "" {
			tableSlug = clientType.TableSlug
		}

		query = fmt.Sprintf(`SELECT * FROM %s WHERE guid = $1`, tableSlug)
		rows, err := conn.Query(ctx, query, req.UserId)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get user data")
		}
		defer rows.Close()

		for rows.Next() {
			values, err := rows.Values()
			if err != nil {
				return nil, errors.Wrap(err, "failed to get user values")
			}

			user = make(map[string]interface{}, len(values))
			for i, value := range values {
				fieldName := rows.FieldDescriptions()[i].Name
				if strings.Contains(fieldName, "_id") || fieldName == "guid" {
					if arr, ok := value.([16]uint8); ok {
						value = helper.ConvertGuid(arr)
					}
				}
				user[fieldName] = value
			}
		}

		if user[connection.FieldSlug] != nil || user["guid"] != nil {
			params := make(map[string]interface{})

			switch fieldValue := user[connection.FieldSlug].(type) {
			case []interface{}:
				params["guid"] = fieldValue
			case string:
				params["guid"] = fmt.Sprintf(`%%%s%%`, user[connection.TableSlug+"_id"])
			case nil:
				if guid, ok := user[connection.TableSlug+"_id"]; ok {
					params["guid"] = guid
				}
			}

			query = fmt.Sprintf(`SELECT * FROM %s WHERE deleted_at IS NULL AND guid = $1`, connection.TableSlug)
			rows, err := conn.Query(ctx, query, params["guid"])
			if err != nil {
				return nil, errors.Wrap(err, "failed to get connection options")
			}
			defer rows.Close()

			for rows.Next() {
				values, err := rows.Values()
				if err != nil {
					return nil, errors.Wrap(err, "failed to get option values")
				}

				option := make(map[string]interface{}, len(values))
				for i, value := range values {
					fieldName := rows.FieldDescriptions()[i].Name
					if strings.Contains(fieldName, "_id") || fieldName == "guid" {
						if arr, ok := value.([16]uint8); ok {
							value = helper.ConvertGuid(arr)
						}
					}
					option[fieldName] = value
				}
				options = append(options, option)
			}
		}
	}

	data, err := helper.ConvertMapToStruct(map[string]interface{}{
		"response": options,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert options to struct")
	}

	return &nb.GetConnectionOptionsResponse{
		TableSlug: connection.TableSlug,
		Data:      data,
	}, nil
}
