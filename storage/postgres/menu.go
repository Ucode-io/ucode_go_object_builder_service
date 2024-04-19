package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/structpb"
)

type menuRepo struct {
	db *pgxpool.Pool
}

func NewMenuRepo(db *pgxpool.Pool) storage.MenuRepoI {
	return &menuRepo{
		db: db,
	}
}

func (m *menuRepo) Create(ctx context.Context, req *nb.CreateMenuRequest) (resp *nb.Menu, err error) {
	if !strings.Contains(strings.Join(config.MENU_TYPES, ","), req.Type) {
		return &nb.Menu{}, errors.New("unsupported menu type")
	}
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return &nb.Menu{}, err
	}
	defer tx.Rollback(ctx)

	if req.Id == "" {
		req.Id = uuid.NewString()
	}

	jsonAttr, err := json.Marshal(req.Attributes)
	if err != nil {
		return &nb.Menu{}, err
	}

	var (
		parentId interface{} = req.ParentId
		layoutId interface{} = req.LayoutId
		tableId  interface{} = req.TableId
	)
	if req.ParentId == "" {
		parentId = nil
	}
	if req.LayoutId == "" {
		layoutId = nil
	}
	if req.TableId == "" {
		tableId = nil
	}

	query := `INSERT INTO "menu" (
		id,
		label,
		parent_id,
		layout_id,
		table_id,
		type,
		icon,
		attributes,
		menu_settings_id
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL)`

	_, err = tx.Exec(ctx, query, req.Id, req.Label, parentId, layoutId, tableId, req.Type, req.Icon, jsonAttr)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Menu{}, err
	}

	query = `SELECT guid FROM "role"`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		tx.Rollback(ctx)
		return &nb.Menu{}, err
	}

	query = `INSERT INTO "menu_permission" (
		menu_id,
		role_id
	) VALUES ($1, $2)`

	for rows.Next() {
		var roleId string

		err := rows.Scan(&roleId)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Menu{}, err
		}

		_, err = tx.Exec(ctx, query, req.Id, roleId)
		if err != nil {
			tx.Rollback(ctx)
			return &nb.Menu{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return &nb.Menu{}, err
	}

	return m.GetById(ctx, &nb.MenuPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (m *menuRepo) GetById(ctx context.Context, req *nb.MenuPrimaryKey) (resp *nb.Menu, err error) {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var (
		id              sql.NullString
		label           sql.NullString
		parentId        sql.NullString
		layoutId        sql.NullString
		tableId         sql.NullString
		menuType        sql.NullString
		icon            sql.NullString
		microfrontendId sql.NullString
		isVisible       sql.NullBool
		isStatic        sql.NullBool
		order           sql.NullInt16
		webpageId       sql.NullString

		guid         sql.NullString
		menuId       sql.NullString
		roleId       sql.NullString
		write        sql.NullBool
		read         sql.NullBool
		update       sql.NullBool
		delete       sql.NullBool
		menuSettings sql.NullBool

		tId                   sql.NullString
		tLabel                sql.NullString
		tSlug                 sql.NullString
		tIcon                 sql.NullString
		tDesc                 sql.NullString
		tFolderID             sql.NullString
		tShowInMenu           sql.NullBool
		tSubtitleFieldSlug    sql.NullString
		tIsChanged            sql.NullBool
		tIsSystem             sql.NullBool
		tIsSoftDelete         sql.NullBool
		tIsCached             sql.NullBool
		tIsLoginTable         sql.NullBool
		tIsOrderBy            sql.NullBool
		tIsSectionColumnCount sql.NullInt16
		tIsWithIncrementId    sql.NullBool
	)

	query := `
		SELECT 
			m."id",
			m."label",
			m."parent_id",
			m."layout_id",
			m."table_id",
			m."type",
			m."icon",
			m."microfrontend_id",
			m."is_visible",
			m."is_static",
			m."order",
			m."webpage_id",
			m."attributes",

			mp."guid",
			mp."menu_id",
			mp."role_id",
			mp."write",
			mp."read",
			mp."update",
			mp."delete",
			mp."menu_settings",

			t."id",
			t."label",
			t."slug",
			t."icon",
			t."description",
			t."folder_id",
			t."show_in_menu",
			t."subtitle_field_slug",
			t."is_changed",
			t."is_system",
			t."soft_delete",
			t."is_cached",
			t."is_changed_by_host",
			t."is_login_table",
			t."attributes",
			t."order_by",
			t."section_column_count",
			t."with_increment_id"
		FROM 
			"menu" m
		LEFT JOIN 
			"menu_permission" mp ON m."id" = mp."menu_id"
		LEFT JOIN
			"table" t ON m."table_id" = t."id"
		WHERE 
			m."id" = $1
	`

	var (
		attrData        []byte
		attrTableData   sql.NullString
		isChangedByHost sql.NullString
	)
	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&id,
		&label,
		&parentId,
		&layoutId,
		&tableId,
		&menuType,
		&icon,
		&microfrontendId,
		&isVisible,
		&isStatic,
		&order,
		&webpageId,
		&attrData,

		&guid,
		&menuId,
		&roleId,
		&write,
		&read,
		&update,
		&delete,
		&menuSettings,

		&tId,
		&tLabel,
		&tSlug,
		&tIcon,
		&tDesc,
		&tFolderID,
		&tShowInMenu,
		&tSubtitleFieldSlug,
		&tIsChanged,
		&tIsSystem,
		&tIsSoftDelete,
		&tIsCached,
		&isChangedByHost,
		&tIsLoginTable,
		&attrTableData,
		&tIsOrderBy,
		&tIsSectionColumnCount,
		&tIsWithIncrementId,
	)
	if err != nil {
		return &nb.Menu{}, err
	}

	var attrDataStruct *structpb.Struct
	if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
		return &nb.Menu{}, err
	}

	var attrTableStruct *structpb.Struct
	if attrTableData.Valid {
		if err := json.Unmarshal([]byte(attrTableData.String), &attrTableStruct); err != nil {
			fmt.Println("dfkgd->", err)
			return &nb.Menu{}, err
		}
	}

	permission := map[string]interface{}{
		"guid":          guid.String,
		"menu_id":       menuId.String,
		"role_id":       roleId.String,
		"write":         write.Bool,
		"read":          read.Bool,
		"update":        update.Bool,
		"delete":        delete.Bool,
		"menu_settings": menuSettings.Bool,
	}
	permissionStruct, err := helper.ConvertMapToStruct(permission)
	if err != nil {
		return &nb.Menu{}, err
	}

	var isChangedByHostStruct *structpb.Struct
	if isChangedByHost.Valid {
		if err := json.Unmarshal([]byte(isChangedByHost.String), &isChangedByHostStruct); err != nil {
			return &nb.Menu{}, err
		}
	}

	table := map[string]interface{}{
		"id":                   tId.String,
		"label":                tLabel.String,
		"slug":                 tSlug.String,
		"icon":                 tIcon.String,
		"description":          tDesc.String,
		"folder_id":            tFolderID.String,
		"show_in_menu":         tShowInMenu.Bool,
		"subtitle_field_slug":  tSubtitleFieldSlug.String,
		"is_changed":           tIsChanged.Bool,
		"is_system":            tIsSystem.Bool,
		"soft_delete":          tIsSoftDelete.Bool,
		"is_cached":            tIsCached.Bool,
		"is_changed_by_host":   isChangedByHostStruct,
		"is_login_table":       tIsLoginTable.Bool,
		"attributes":           attrTableStruct,
		"order_by":             tIsOrderBy.Bool,
		"section_column_count": tIsSectionColumnCount.Int16,
		"with_increment_id":    tIsWithIncrementId.Bool,
	}
	tableStruct, err := helper.ConvertMapToStruct(table)
	if err != nil {
		return &nb.Menu{}, err
	}

	data := map[string]interface{}{
		"permission": permissionStruct,
		"table":      tableStruct,
	}
	dataStruct, err := helper.ConvertMapToStruct(data)
	if err != nil {
		return &nb.Menu{}, err
	}

	return &nb.Menu{
		Id:              id.String,
		Label:           label.String,
		ParentId:        parentId.String,
		LayoutId:        layoutId.String,
		TableId:         tableId.String,
		Type:            menuType.String,
		Icon:            icon.String,
		MicrofrontendId: microfrontendId.String,
		IsVisible:       isVisible.Bool,
		IsStatic:        isStatic.Bool,
		Order:           int32(order.Int16),
		WebpageId:       webpageId.String,
		Attributes:      attrDataStruct,
		Data:            dataStruct,
	}, nil
}

func (m *menuRepo) GetAll(ctx context.Context, req *nb.GetAllMenusRequest) (resp *nb.GetAllMenusResponse, err error) {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	params := make(map[string]interface{})
	resp = &nb.GetAllMenusResponse{}

	query := `
		SELECT 
			m."id",
			m."label",
			m."parent_id",
			m."layout_id",
			m."table_id",
			m."type",
			m."icon",
			m."microfrontend_id",
			m."is_visible",
			m."is_static",
			m."order",
			m."webpage_id",
			m."attributes",

			mp."guid",
			mp."menu_id",
			mp."role_id",
			mp."write",
			mp."read",
			mp."update",
			mp."delete",
			mp."menu_settings",

			t."id",
			t."label",
			t."slug",
			t."icon",
			t."description",
			t."folder_id",
			t."show_in_menu",
			t."subtitle_field_slug",
			t."is_changed",
			t."is_system",
			t."soft_delete",
			t."is_cached",
			t."is_changed_by_host",
			t."is_login_table",
			COALESCE(t."attributes", '{}'::jsonb) AS attributes,
			t."order_by",
			t."section_column_count",
			t."with_increment_id"
		FROM "menu" m
		LEFT JOIN 
			"menu_permission" mp ON m."id" = mp."menu_id"
		LEFT JOIN
			"table" t ON m."table_id" = t."id"
		WHERE m.parent_id = :parent_id
		ORDER BY m."order" ASC
	`

	if req.Offset >= 0 {
		query += ` OFFSET :offset `
		params["offset"] = req.Offset
	}
	if req.Limit > 0 {
		query += ` LIMIT :limit `
		params["limit"] = req.Limit
	}
	params["parent_id"] = req.ParentId

	query, args := helper.ReplaceQueryParams(query, params)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.GetAllMenusResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id              sql.NullString
			label           sql.NullString
			parentId        sql.NullString
			layoutId        sql.NullString
			tableId         sql.NullString
			menuType        sql.NullString
			icon            sql.NullString
			microfrontendId sql.NullString
			isVisible       sql.NullBool
			isStatic        sql.NullBool
			order           sql.NullInt16
			webpageId       sql.NullString
			attrData        []byte

			guid         sql.NullString
			menuId       sql.NullString
			roleId       sql.NullString
			write        sql.NullBool
			read         sql.NullBool
			update       sql.NullBool
			delete       sql.NullBool
			menuSettings sql.NullBool

			tId                   sql.NullString
			tLabel                sql.NullString
			tSlug                 sql.NullString
			tIcon                 sql.NullString
			tDesc                 sql.NullString
			tFolderID             sql.NullString
			tShowInMenu           sql.NullBool
			tSubtitleFieldSlug    sql.NullString
			tIsChanged            sql.NullBool
			tIsSystem             sql.NullBool
			tIsSoftDelete         sql.NullBool
			tIsCached             sql.NullBool
			tIsLoginTable         sql.NullBool
			tIsOrderBy            sql.NullBool
			tIsSectionColumnCount sql.NullInt16
			tIsWithIncrementId    sql.NullBool
			attrTableData         sql.NullString
			isChangedByHost       sql.NullString
		)

		err := rows.Scan(
			&id,
			&label,
			&parentId,
			&layoutId,
			&tableId,
			&menuType,
			&icon,
			&microfrontendId,
			&isVisible,
			&isStatic,
			&order,
			&webpageId,
			&attrData,

			&guid,
			&menuId,
			&roleId,
			&write,
			&read,
			&update,
			&delete,
			&menuSettings,

			&tId,
			&tLabel,
			&tSlug,
			&tIcon,
			&tDesc,
			&tFolderID,
			&tShowInMenu,
			&tSubtitleFieldSlug,
			&tIsChanged,
			&tIsSystem,
			&tIsSoftDelete,
			&tIsCached,
			&isChangedByHost,
			&tIsLoginTable,
			&attrTableData,
			&tIsOrderBy,
			&tIsSectionColumnCount,
			&tIsWithIncrementId,
		)
		if err != nil {
			return &nb.GetAllMenusResponse{}, nil
		}

		var attrDataStruct *structpb.Struct
		if err := json.Unmarshal(attrData, &attrDataStruct); err != nil {
			return &nb.GetAllMenusResponse{}, err
		}

		var attrTableStruct *structpb.Struct
		if attrTableData.Valid {
			if err := json.Unmarshal([]byte(attrTableData.String), &attrTableStruct); err != nil {
				return &nb.GetAllMenusResponse{}, err
			}
		}

		permission := map[string]interface{}{
			"guid":          guid.String,
			"menu_id":       menuId.String,
			"role_id":       roleId.String,
			"write":         write.Bool,
			"read":          read.Bool,
			"update":        update.Bool,
			"delete":        delete.Bool,
			"menu_settings": menuSettings.Bool,
		}
		permissionStruct, err := helper.ConvertMapToStruct(permission)
		if err != nil {
			return &nb.GetAllMenusResponse{}, err
		}

		var isChangedByHostStruct *structpb.Struct
		if isChangedByHost.Valid {
			if err := json.Unmarshal([]byte(isChangedByHost.String), &isChangedByHostStruct); err != nil {
				return &nb.GetAllMenusResponse{}, err
			}
		}

		table := map[string]interface{}{
			"id":                   tId.String,
			"label":                tLabel.String,
			"slug":                 tSlug.String,
			"icon":                 tIcon.String,
			"description":          tDesc.String,
			"folder_id":            tFolderID.String,
			"show_in_menu":         tShowInMenu.Bool,
			"subtitle_field_slug":  tSubtitleFieldSlug.String,
			"is_changed":           tIsChanged.Bool,
			"is_system":            tIsSystem.Bool,
			"soft_delete":          tIsSoftDelete.Bool,
			"is_cached":            tIsCached.Bool,
			"is_changed_by_host":   isChangedByHostStruct,
			"is_login_table":       tIsLoginTable.Bool,
			"attributes":           attrTableStruct,
			"order_by":             tIsOrderBy.Bool,
			"section_column_count": tIsSectionColumnCount.Int16,
			"with_increment_id":    tIsWithIncrementId.Bool,
		}
		tableStruct, err := helper.ConvertMapToStruct(table)
		if err != nil {
			return &nb.GetAllMenusResponse{}, err
		}

		data := map[string]interface{}{
			"permission": permissionStruct,
			"table":      tableStruct,
		}
		dataStruct, err := helper.ConvertMapToStruct(data)
		if err != nil {
			return &nb.GetAllMenusResponse{}, err
		}

		resp.Menus = append(resp.Menus, &nb.MenuForGetAll{
			Id:              id.String,
			Label:           label.String,
			ParentId:        parentId.String,
			LayoutId:        layoutId.String,
			TableId:         tableId.String,
			Type:            menuType.String,
			Icon:            icon.String,
			MicrofrontendId: microfrontendId.String,
			IsStatic:        isStatic.Bool,
			WebpageId:       webpageId.String,
			Attributes:      attrDataStruct,
			Data:            dataStruct,
		})
	}

	query = `SELECT COUNT(*) FROM "menu"`

	err = conn.QueryRow(ctx, query).Scan(&resp.Count)
	if err != nil {
		return &nb.GetAllMenusResponse{}, err
	}

	return resp, nil
}

func (m *menuRepo) Update(ctx context.Context, req *nb.Menu) (resp *nb.Menu, err error) {
	if !strings.Contains(strings.Join(config.MENU_TYPES, ","), req.Type) {
		return &nb.Menu{}, errors.New("unsupported menu type")
	}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var (
		parentId interface{} = req.ParentId
		layoutId interface{} = req.LayoutId
		tableId  interface{} = req.TableId
	)
	if req.ParentId == "" {
		parentId = nil
	}
	if req.LayoutId == "" {
		layoutId = nil
	}
	if req.TableId == "" {
		tableId = nil
	}

	query := `UPDATE "menu" SET
		"label" = $1,
		"parent_id" = $2,
		"layout_id" = $3,
		"table_id" = $4,
		"type" = $5,
		"icon" = $6,
		"attributes" = $7,
		"updated_at" = now()
	WHERE id = $8
 	`

	_, err = conn.Exec(ctx, query, req.Label, parentId, layoutId, tableId, req.Type, req.Icon, req.Attributes, req.Id)
	if err != nil {
		return &nb.Menu{}, err
	}

	return m.GetById(ctx, &nb.MenuPrimaryKey{Id: req.Id, ProjectId: req.ProjectId})
}

func (m *menuRepo) UpdateMenuOrder(ctx context.Context, req *nb.UpdateMenuOrderRequest) error {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return err
	}
	defer conn.Close()

	for i, menu := range req.Menus {
		_, err := conn.Exec(ctx, `UPDATE menu SET "order" = $1 WHERE id = $2`, i+1, menu.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *menuRepo) Delete(ctx context.Context, req *nb.MenuPrimaryKey) error {
	if strings.Contains(strings.Join(config.STATIC_MENU_IDS, ","), req.Id) {
		return errors.New("cannot delete default menu")
	}

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return err
	}
	defer conn.Close()

	query := `DELETE from "menu" WHERE id = $1`

	_, err = conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	query = `DELETE from "menu_permission" WHERE menu_id = $1`

	_, err = conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}

////////// MENU SETTINGS

func (m *menuRepo) GetAllMenuSettings(ctx context.Context, req *nb.GetAllMenuSettingsRequest) (resp *nb.GetAllMenuSettingsResponse, err error) {

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp = &nb.GetAllMenuSettingsResponse{}
	params := make(map[string]interface{})

	query := `
		SELECT 
			"id",
			"icon_style",
			"icon_size"
		FROM "menu_setting"
	`

	if req.Offset >= 0 {
		query += ` OFFSET :offset `
		params["offset"] = req.Offset
	}
	if req.Limit > 0 {
		query += ` LIMIT :limit `
		params["limit"] = req.Limit
	}

	query, args := helper.ReplaceQueryParams(query, params)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return &nb.GetAllMenuSettingsResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id        string
			iconStyle sql.NullString
			iconSize  sql.NullString
		)

		err := rows.Scan(
			&id,
			&iconStyle,
			&iconSize,
		)
		if err != nil {
			return nil, err
		}

		resp.MenuSettings = append(resp.MenuSettings, &nb.MenuSettings{
			Id:        id,
			IconStyle: iconStyle.String,
			IconSize:  iconSize.String,
		})
	}

	query = `SELECT COUNT(*) FROM "menu_setting"`

	err = conn.QueryRow(ctx, query).Scan(&resp.Count)
	if err != nil {
		return &nb.GetAllMenuSettingsResponse{}, err
	}

	return resp, nil
}

func (m *menuRepo) GetByIDMenuSettings(ctx context.Context, req *nb.MenuSettingPrimaryKey) (resp *nb.MenuSettings, err error) {
	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	resp = &nb.MenuSettings{}

	query := `
			SELECT 
				"id",
				"icon_style",
				"icon_size"	
			FROM "menu_setting"
			WHERE id = $1
	`

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.IconStyle,
		&resp.IconSize,
	)
	if err != nil {
		return resp, err
	}

	resp.MenuTemplateId = req.TemplateId

	return resp, nil

}
