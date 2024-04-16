package postgres

import (
	"context"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
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
	conn := psqlpool.Get(req.GetProjectId())

	query := `
		SELECT
			"guid",
			"project_id",
			"name",
			"self_register",
			"self_recover",
			"client_platform_ids",
			"confirm_by",
			"is_system",
			"table_slug",
			"default_page"
		FROM client_type WHERE id = $1 OR "name" = $1
	`

	var (
		clientType     ClientType
		tableSlug      string
		userId         string
		roleId         string
		userFound      bool
		role           Role
		clientPlatform ClientPlatform
		connections    = []*nb.TableClientType{}
		permissions    = []*nb.RecordPermission{}
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
		&clientType.TableSlug,
		&clientType.DefaultPage,
	)
	if err != nil {
		return &nb.LoginDataRes{
			UserFound: false,
		}, err
	}

	if clientType.TableSlug != "" {
		tableSlug = clientType.TableSlug
	}

	query = `SELECT guid, role_id FROM ` + tableSlug + ` WHERE guid = $1 AND client_type_id = $2`

	err = conn.QueryRow(ctx, query, req.UserId, req.ClientType).Scan(
		&userId,
		&roleId,
	)
	if err != nil {
		return &nb.LoginDataRes{
			UserFound: false,
		}, err
	}

	if userId != "" {
		userFound = true
	}

	query = `SELECT 
		"guid",
		"name",
		"project_id",
		"client_platform_id",
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
		}, err
	}

	query = `SELECT 
		"guid",
		"name",
		"project_id",
		"subdomain"
	FROM "client_platform" WHERE "guid" = $1
	`

	err = conn.QueryRow(ctx, query, role.ClientPlatformId).Scan(
		&clientPlatform.Guid,
		&clientPlatform.Name,
		&clientPlatform.ProjectId,
		&clientPlatform.Subdomain,
	)
	if err != nil {
		return &nb.LoginDataRes{
			UserFound: false,
		}, err
	}

	query = `SELECT 
		"table_slug",
		"view_slug",
		"view_label",
		"icon",
		"name"
	FROM "connection" WHERE client_type_id = $1`

	rows, err := conn.Query(ctx, query, clientType.Guid)
	if err != nil {
		return &nb.LoginDataRes{}, err
	}
	defer rows.Close()

	for rows.Next() {
		connection := nb.TableClientType{}

		err = rows.Scan(
			&connection.Slug,
			&connection.ViewSlug,
			&connection.ViewLabel,
			&connection.Icon,
			&connection.Label,
		)

		connections = append(connections, &connection)
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
		"pdf_action"
	FROM "record_permission" WHERE role_id = $1`

	rows, err = conn.Query(ctx, query, roleId)
	if err != nil {
		return &nb.LoginDataRes{}, err
	}
	defer rows.Close()

	for rows.Next() {
		permission := nb.RecordPermission{}

		err = rows.Scan(
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
		)
		if err != nil {
			return &nb.LoginDataRes{}, err
		}

		permissions = append(permissions, &permission)
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
		Permissions: permissions,
	}, nil
}
