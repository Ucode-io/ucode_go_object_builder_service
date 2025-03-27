package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/models"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/security"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/structpb"
)

type loginRepo struct {
	db *psqlpool.Pool
}

func NewLoginRepo(db *psqlpool.Pool) storage.LoginRepoI {
	return &loginRepo{
		db: db,
	}
}

func (l *loginRepo) LoginData(ctx context.Context, req *nb.LoginDataReq) (resp *nb.LoginDataRes, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "login.LoginData")
	defer dbSpan.Finish()

	var (
		conn                                        = psqlpool.Get(req.GetResourceEnvironmentId())
		clientType                                  models.ClientType
		tableSlug                                   = `user`
		userId, roleId, guid                        string
		userFound, comparePassword, isLoginStrategy bool
		role                                        models.Role
		clientPlatform                              models.ClientPlatform
		connections                                 = []*nb.TableClientType{}
		permissions                                 = []*nb.RecordPermission{}
		tableSlugNull, defaultPageNull              sql.NullString
		globalPermission                            = &nb.GlobalPermission{}
		errResp                                     = &nb.LoginDataRes{UserFound: false}
	)

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
			"default_page",
			"session_limit"
		FROM client_type WHERE "guid" = $1 OR "name" = $1::varchar
	`

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
		&clientType.SessionLimit,
	)
	if err != nil {
		return errResp, errors.Wrap(err, "error getting client type")
	}

	clientType.TableSlug = tableSlugNull.String
	clientType.DefaultPage = defaultPageNull.String

	if clientType.TableSlug != "" && clientType.TableSlug != "user" {
		tableSlug = clientType.TableSlug
	}

	if req.UserId == "" {
		return errResp, nil
	}

	userInfo, err := helper.GetItemLogin(ctx, conn, tableSlug, req.UserId, req.ClientType)
	if err != nil {
		return errResp, nil
	}

	if len(userInfo) == 0 {
		return errResp, nil
	}

	guid = cast.ToString(userInfo["guid"])
	userId = cast.ToString(userInfo["user_id_auth"])
	roleId = cast.ToString(userInfo["role_id"])

	if userId != "" {
		userFound = true
		if len(req.GetPassword()) != 0 {
			var attrData []byte

			query = `SELECT attributes FROM "table" where slug = $1`
			if err := conn.QueryRow(ctx, query, tableSlug).Scan(&attrData); err != nil {
				return errResp, nil
			}

			var attrDataStruct *structpb.Struct
			if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
				return errResp, nil
			}

			attrDataMap, err := helper.ConvertStructToMap(attrDataStruct)
			if err != nil {
				return errResp, nil
			}

			authInfo, ok := attrDataMap["auth_info"].(map[string]any)
			if !ok {
				return errResp, nil
			}

			loginStrategy := cast.ToStringSlice(authInfo["login_strategy"])

			for _, strategy := range loginStrategy {
				if config.CheckPasswordLoginStrategies[strategy] {
					isLoginStrategy = true
					break
				}
			}

			if isLoginStrategy {
				checkPassword, err := security.ComparePasswordBcrypt(cast.ToString(userInfo[cast.ToString(authInfo["password"])]), req.Password)
				if err != nil {
					return &nb.LoginDataRes{UserFound: false, ComparePassword: false}, nil
				}

				if !checkPassword {
					return &nb.LoginDataRes{UserFound: false, ComparePassword: false}, nil
				}
				comparePassword = true
			}
		} else {
			comparePassword = true
		}
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
		return errResp, errors.Wrap(err, "error getting role")
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
			return errResp, errors.Wrap(err, "error getting client platform")
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
		return errResp, errors.Wrap(err, "error getting connections")
	}
	defer rows.Close()

	for rows.Next() {
		var slug, viewSlug, viewLabel, icon, label sql.NullString

		err = rows.Scan(
			&slug,
			&viewSlug,
			&viewLabel,
			&icon,
			&label,
		)
		if err != nil {
			return errResp, errors.Wrap(err, "error scanning connections")
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
		return errResp, errors.Wrap(err, "error getting record permissions")
	}
	defer recPermissions.Close()

	for recPermissions.Next() {
		var permission nb.RecordPermission

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
			return errResp, errors.Wrap(err, "error scanning record permissions")
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
			"version_button",
			"chatwoot_button",
			"gitbook_button",
			"gpt_button"
		FROM global_permission
		WHERE role_id = $1
		LIMIT 1
	`

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
		&globalPermission.ChatwootButton,
		&globalPermission.GitbookButton,
		&globalPermission.GptButton,
	)
	if err != nil {
		return errResp, errors.Wrap(err, "error getting global permission")
	}

	userdata, err := helper.ConvertMapToStruct(userInfo)
	if err != nil {
		return errResp, errors.Wrap(err, "convert map to struct")
	}

	return &nb.LoginDataRes{
		UserFound:       userFound,
		ComparePassword: comparePassword,
		UserId:          guid,
		LoginTableSlug:  tableSlug,
		ClientType: &nb.ClientType{
			Guid:         clientType.Guid,
			Name:         clientType.Name,
			ConfirmBy:    nb.ConfirmStrategies(0),
			SelfRegister: clientType.SelfRegister,
			SelfRecover:  clientType.SelfRecover,
			ProjectId:    clientType.ProjectId,
			Tables:       connections,
			DefaultPage:  clientType.DefaultPage,
			SessionLimit: clientType.SessionLimit,
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
		UserIdAuth:       userId,
		UserData:         userdata,
	}, nil
}

func (l *loginRepo) GetConnectionOptions(ctx context.Context, req *nb.GetConnetionOptionsRequest) (resp *nb.GetConnectionOptionsResponse, err error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "login.GetConnectionOptions")
	defer dbSpan.Finish()

	var (
		conn       = psqlpool.Get(req.GetResourceEnvironmentId())
		options    []map[string]any
		connection models.Connection
		clientType models.ClientType
		user       map[string]any
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

		query = fmt.Sprintf(`SELECT * FROM %s WHERE user_id_auth = $1`, tableSlug)
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

			user = make(map[string]any, len(values))
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
			params := make(map[string]any)

			switch fieldValue := user[connection.FieldSlug].(type) {
			case []any:
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

				option := make(map[string]any, len(values))
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

	data, err := helper.ConvertMapToStruct(map[string]any{
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
