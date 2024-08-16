package helper

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/structpb"
)

type ViewRelationModel struct {
	RoleID           string
	TableSlug        string
	RelationID       string
	ViewPermission   bool
	CreatePermission bool
	EditPermission   bool
	DeletePermission bool
}

type RelationHelper struct {
	Tx           pgx.Tx
	Conn         *pgxpool.Pool
	FieldName    string
	TableID      string
	LayoutID     string
	TableSlug    string
	TabID        string
	Fields       []*nb.FieldForSection
	SectionID    string
	View         *nb.CreateViewRequest
	Field        *nb.CreateFieldRequest
	FieldID      string
	RoleID       string
	TableFrom    string
	TableTo      string
	Label        string
	Order        int
	Type         string
	RelationID   string
	RoleIDs      []string
	RelationType string
	FieldFrom    string
	FieldTo      string
	Attributes   *structpb.Struct
}

type RelationLayout struct {
	Tx         pgx.Tx
	Conn       *pgxpool.Pool
	TableId    string
	RelationId string
}

func CheckRelationFieldExists(ctx context.Context, req RelationHelper) (bool, string, error) {
	rows, err := req.Tx.Query(ctx, "SELECT slug FROM field WHERE table_id = $1 AND slug LIKE $2 ORDER BY slug ASC", req.TableID, req.FieldName+"%")
	if err != nil {
		return false, "", err
	}
	defer rows.Close()

	var lastField string
	for rows.Next() {
		var fieldSlug string
		err := rows.Scan(&fieldSlug)
		if err != nil {
			return false, "", err
		}
		lastField = fieldSlug
	}

	if lastField != "" {
		parts := strings.Split(lastField, "_")
		if len(parts) > 2 {
			index, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return false, "", err
			}

			index++
			lastField = fmt.Sprintf("%s_%d", req.FieldName, index)
		} else if len(parts) == 2 {
			lastField = fmt.Sprintf("%s_2", req.FieldName)
		}
	}

	return lastField != "", lastField, nil
}

func LayoutFindOne(ctx context.Context, req RelationHelper) (resp *nb.LayoutResponse, err error) {
	resp = &nb.LayoutResponse{}
	query := `SELECT 
		"id"
	FROM "layout" WHERE table_id = $1 LIMIT 1`

	err = req.Tx.QueryRow(ctx, query, req.TableID).Scan(
		&resp.Id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, err
	}
	return resp, nil
}

func TabFindOne(ctx context.Context, req RelationHelper) (resp *nb.TabResponse, err error) {
	resp = &nb.TabResponse{}
	query := `SELECT 
		id
	FROM "tab" 
	WHERE layout_id = $1 AND type = 'section' LIMIT 1`

	err = req.Tx.QueryRow(ctx, query, req.LayoutID).Scan(
		&resp.Id,
	)
	if err != nil {
		log.Println("Error while finding single tab for relation", err)
		return nil, err
	}
	return resp, nil
}

func TabCreate(ctx context.Context, req RelationHelper) (tab *nb.TabResponse, err error) {
	tab = &nb.TabResponse{}

	id := uuid.New().String()

	atrb := []byte("{}")

	if req.Attributes != nil {
		atrb, err = json.Marshal(req.Attributes)
		if err != nil {
			return &nb.TabResponse{}, err
		}
	}

	query := `
		INSERT INTO "tab" (
			"id",
			"order",
			"label",
			"type",
			"table_slug",
			"layout_id",
			"relation_id",
			attributes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING "id"
	`
	err = req.Tx.QueryRow(ctx, query,
		id,
		req.Order,
		req.Label,
		req.Type,
		req.TableSlug,
		req.LayoutID,
		req.RelationID,
		atrb,
	).Scan(&tab.Id)
	if err != nil {
		log.Println("Error while creating tab for relation", err)
		return tab, err
	}

	return tab, nil
}

func SectionFind(ctx context.Context, req RelationHelper) (resp []*nb.Section, err error) {
	resp = []*nb.Section{}
	query := `SELECT 
		id,
		fields
	FROM "section" 
	WHERE tab_id = $1
	ORDER BY created_at DESC`

	rows, err := req.Tx.Query(ctx, query, req.TabID)
	if err != nil {
		log.Println("Error while finding single section for relation", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			section nb.Section
			fields  sql.NullString
		)
		if err := rows.Scan(&section.Id, &fields); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		var fieldsForSection []*nb.FieldForSection
		if fields.Valid {
			if err := json.Unmarshal([]byte(fields.String), &fieldsForSection); err != nil {
				return nil, err
			}
		}

		section.Fields = fieldsForSection
		resp = append(resp, &section)
	}

	return resp, nil
}

func SectionCreate(ctx context.Context, req RelationHelper) error {

	id := uuid.New().String()

	query := `
		INSERT INTO "section" (
			id,
			"order",
			"column",
			"label",
			"fields",
			"table_id",
			"tab_id"
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := req.Tx.Exec(ctx, query,
		id,
		req.Order,
		"SINGLE",
		"Info",
		req.Fields,
		req.TableID,
		req.TabID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert section: %v", err)
	}

	return nil
}

func SectionFindOneAndUpdate(ctx context.Context, req RelationHelper) error {
	query := `
		UPDATE "section" SET
			"fields" = $2
		WHERE id = $1
	`

	_, err := req.Tx.Exec(ctx, query, req.SectionID, req.Fields)
	if err != nil {
		return fmt.Errorf("failed to update section: %v", err)
	}

	return nil
}

func ViewCreate(ctx context.Context, req RelationHelper) error {
	query := `
	INSERT INTO "view" (
		"id",
		"table_slug",
		"type",
		"group_fields",
		"view_fields",
		"main_field",
		"disable_dates",
		"quick_filters",
		"users",
		"name",
		"columns",
		"calendar_from_slug",
		"calendar_to_slug",
		"time_interval",
		"multiple_insert",
		"status_field_slug",
		"is_editable",
		"relation_table_slug",
		"relation_id",
		"multiple_insert_field",
		"updated_fields",
		"table_label",
		"default_limit",
		"default_editable",
		"order",
		"name_uz",
		"name_en",
		"attributes"
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
	`

	fmt.Printf("View: %v\n", req.View)

	_, err := req.Tx.Exec(ctx, query,
		req.View.Id,
		req.View.TableSlug,
		req.View.Type,
		req.View.GroupFields,
		req.View.ViewFields,
		req.View.MainField,
		req.View.DisableDates,
		req.View.QuickFilters,
		req.View.Users,
		req.View.Name,
		req.View.Columns,
		req.View.CalendarFromSlug,
		req.View.CalendarToSlug,
		req.View.TimeInterval,
		req.View.MultipleInsert,
		req.View.StatusFieldSlug,
		req.View.IsEditable,
		req.View.RelationTableSlug,
		req.View.RelationId,
		req.View.MultipleInsertField,
		req.View.UpdatedFields,
		req.View.TableLabel,
		req.View.DefaultLimit,
		req.View.DefaultEditable,
		req.View.Order,
		req.View.NameUz,
		req.View.NameEn,
		req.View.Attributes,
	)
	if err != nil {
		return fmt.Errorf("failed to insert view: %v", err)
	}

	return nil
}

func RolesFind(ctx context.Context, req RelationHelper) (resp []string, err error) {
	resp = []string{}
	query := `SELECT guid FROM role`

	rows, err := req.Tx.Query(ctx, query)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		roleId := ""

		err = rows.Scan(&roleId)
		if err != nil {
			return resp, err
		}

		resp = append(resp, roleId)
	}

	return resp, nil
}

func RelationFieldPermission(ctx context.Context, req RelationHelper) error {
	query := `
	INSERT INTO "field_permission" (
		guid,
		field_id,
		table_slug,
		view_permission,
		edit_permission,
		role_id,
		label
	) VALUES `

	var values []interface{}
	var placeholders []string

	for _, roleId := range req.RoleIDs {
		id, _ := uuid.NewRandom()
		values = append(values,
			id.String(),
			req.FieldID,
			req.TableSlug,
			true,
			true,
			roleId,
			req.Label,
		)
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", len(values)-6, len(values)-5, len(values)-4, len(values)-3, len(values)-2, len(values)-1, len(values)))
	}

	query += strings.Join(placeholders, ", ")

	_, err := req.Tx.Exec(ctx, query, values...)
	if err != nil {
		return err
	}

	return nil
}

func UpsertField(ctx context.Context, req RelationHelper) (resp *nb.Field, err error) {
	jsonAttr, err := json.Marshal(req.Field.Attributes)
	if err != nil {
		return &nb.Field{}, err
	}

	query := `
		INSERT INTO field (id, table_id, slug, label, type, relation_id, attributes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	resp = &nb.Field{}
	err = req.Tx.QueryRow(ctx, query,
		req.Field.Id,
		req.Field.TableId,
		req.Field.Slug,
		req.Field.Label,
		req.Field.Type,
		req.Field.RelationId,
		jsonAttr,
	).Scan(&resp.Id)

	if err != nil {
		return nil, fmt.Errorf("failed to insert field: %v", err)
	}

	return resp, nil
}

func TabFind(ctx context.Context, req RelationHelper) (resp []*nb.TabResponse, err error) {
	resp = []*nb.TabResponse{}
	query := `SELECT 
		id
	FROM "tab" 
	WHERE layout_id = $1`

	rows, err := req.Tx.Query(ctx, query, req.LayoutID)
	if err != nil {
		log.Println("Error while finding tabs for relation", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tabID string
		if err := rows.Scan(&tabID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		resp = append(resp, &nb.TabResponse{Id: tabID})
	}

	return resp, nil
}

func ViewRelationPermission(ctx context.Context, req RelationHelper) error {
	insertManyRelationPermissions := []ViewRelationModel{}

	for _, roleId := range req.RoleIDs {
		query := `
			SELECT role_id, table_slug, relation_id
			FROM view_relation_permission
			WHERE role_id = $1 AND table_slug = $2 AND relation_id = $3
			LIMIT 1
		`

		var permission ViewRelationModel
		err := req.Tx.QueryRow(context.Background(), query, roleId, req.TableSlug, req.RelationID).
			Scan(&permission.RoleID, &permission.TableSlug, &permission.RelationID)
		if err != nil && err != pgx.ErrNoRows {
			log.Fatalf("Error fetching permission: %v", err)
			return err
		}

		// If permission doesn't exist, add to the slice for bulk insertion
		if err == pgx.ErrNoRows {
			insertManyRelationPermissions = append(insertManyRelationPermissions, ViewRelationModel{
				RoleID:           roleId,
				TableSlug:        req.TableSlug,
				RelationID:       req.RelationID,
				ViewPermission:   true,
				CreatePermission: true,
				EditPermission:   true,
				DeletePermission: true,
			})
		}
	}

	if len(insertManyRelationPermissions) > 0 {
		stmt := `
			INSERT INTO view_relation_permission (role_id, table_slug, relation_id, view_permission, create_permission, edit_permission, delete_permission)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err := req.Tx.Prepare(context.Background(), "insertManyRelationPermissions", stmt)
		if err != nil {
			return fmt.Errorf("failed to prepare insert statement: %v", err)
		}

		for _, permission := range insertManyRelationPermissions {
			_, err := req.Tx.Exec(context.Background(), "insertManyRelationPermissions", permission.RoleID, permission.TableSlug, permission.RelationID, permission.ViewPermission, permission.CreatePermission, permission.EditPermission, permission.DeletePermission)
			if err != nil {
				return fmt.Errorf("failed to insert relation permission: %v", err)
			}
		}

	}

	return nil
}

func ExecRelation(ctx context.Context, req RelationHelper) error {
	var (
		alterTableSQL, addConstraintSQL string
	)
	switch req.RelationType {
	case config.MANY2ONE:
		alterTableSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN  %s UUID;`, req.TableFrom, req.FieldFrom)
		addConstraintSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES "%s"(guid);
    `, req.TableFrom, req.TableFrom, req.FieldFrom, req.FieldFrom, req.TableTo)
	case config.MANY2MANY:
		alterTableSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN  %s VARCHAR[]`, req.TableFrom, req.FieldFrom)
		addConstraintSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN  %s VARCHAR[]`, req.TableTo, req.FieldTo)
	case config.RECURSIVE:
		alterTableSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN  %s UUID`, req.TableFrom, req.FieldTo)
		addConstraintSQL = fmt.Sprintf(`ALTER TABLE "%s" ADD CONSTRAINT fk_%s_%s_id FOREIGN KEY (%s) REFERENCES "%s"(guid) ON DELETE CASCADE
		`, req.TableFrom, req.TableFrom, req.TableFrom, req.FieldTo, req.TableFrom)
	}

	if _, err := req.Tx.Exec(ctx, alterTableSQL); err != nil {
		return err
	}
	if _, err := req.Tx.Exec(ctx, addConstraintSQL); err != nil {
		return err
	}

	return nil
}

func ViewFindOne(ctx context.Context, req RelationHelper) (resp *nb.View, err error) {
	resp = &nb.View{}
	query := `
		SELECT 
			"id",
			"table_slug",
			"type",
			"name",
			"main_field",
			"disable_dates",
			"columns",
			"quick_filters",
			"users",
			"view_fields",
			"group_fields",
			"calendar_from_slug",
			"calendar_to_slug",
			"time_interval",
			"multiple_insert",
			"status_field_slug",
			"is_editable",
			"relation_table_slug",
			"relation_id",
			"multiple_insert_field",
			"updated_fields",
			"app_id",
			"table_label",
			"default_limit",
			"default_editable",
			"navigate",
			"function_path",
			"order",
			"name_uz",
			"name_en",
			"attributes"
		FROM "view" WHERE relation_id = $1 LIMIT 1`

	var (
		attributes          []byte
		TableSlug           sql.NullString
		Type                sql.NullString
		Name                sql.NullString
		MainField           sql.NullString
		CalendarFromSlug    sql.NullString
		CalendarToSlug      sql.NullString
		TimeInterval        sql.NullInt32
		StatusFieldSlug     sql.NullString
		RelationTableSlug   sql.NullString
		RelationId          sql.NullString
		MultipleInsertField sql.NullString
		AppId               sql.NullString
		TableLabel          sql.NullString
		DefaultLimit        sql.NullString
		FunctionPath        sql.NullString
		Order               sql.NullInt32
		NameUz              sql.NullString
		NameEn              sql.NullString
	)

	err = req.Conn.QueryRow(ctx, query, req.RelationID).Scan(
		&resp.Id,
		&TableSlug,
		&Type,
		&Name,
		&MainField,
		&resp.DisableDates,
		&resp.Columns,
		&resp.QuickFilters,
		&resp.Users,
		&resp.ViewFields,
		&resp.GroupFields,
		&CalendarFromSlug,
		&CalendarToSlug,
		&TimeInterval,
		&resp.MultipleInsert,
		&StatusFieldSlug,
		&resp.IsEditable,
		&RelationTableSlug,
		&RelationId,
		&MultipleInsertField,
		&resp.UpdatedFields,
		&AppId,
		&TableLabel,
		&DefaultLimit,
		&resp.DefaultEditable,
		&resp.Navigate,
		&FunctionPath,
		&Order,
		&NameUz,
		&NameEn,
		&attributes,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return nil, err
	}

	resp = &nb.View{
		Id:                  resp.Id,
		TableSlug:           TableSlug.String,
		Type:                Type.String,
		Name:                Name.String,
		MainField:           MainField.String,
		DisableDates:        resp.DisableDates,
		Columns:             resp.Columns,
		QuickFilters:        resp.QuickFilters,
		Users:               resp.Users,
		ViewFields:          resp.ViewFields,
		GroupFields:         resp.GroupFields,
		CalendarFromSlug:    CalendarFromSlug.String,
		CalendarToSlug:      CalendarToSlug.String,
		TimeInterval:        TimeInterval.Int32,
		MultipleInsert:      resp.MultipleInsert,
		StatusFieldSlug:     StatusFieldSlug.String,
		IsEditable:          resp.IsEditable,
		RelationTableSlug:   RelationTableSlug.String,
		RelationId:          RelationId.String,
		MultipleInsertField: MultipleInsertField.String,
		UpdatedFields:       resp.UpdatedFields,
		AppId:               AppId.String,
		TableLabel:          TableLabel.String,
		DefaultLimit:        DefaultLimit.String,
		DefaultEditable:     resp.DefaultEditable,
		Navigate:            resp.Navigate,
		FunctionPath:        FunctionPath.String,
		Order:               Order.Int32,
		NameUz:              NameUz.String,
		NameEn:              NameEn.String,
		Attributes:          resp.Attributes,
	}
	return resp, nil
}

func FieldFindOne(ctx context.Context, req RelationHelper) (resp *nb.Field, err error) {
	resp = &nb.Field{}
	query := `SELECT 
		id,
		slug,
		label
	FROM "field" WHERE relation_id = $1 LIMIT 1`

	err = req.Tx.QueryRow(ctx, query, req.RelationID).Scan(
		&resp.Id,
		&resp.Slug,
		&resp.Label,
	)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func ViewFindOneByTableSlug(ctx context.Context, req RelationHelper) (resp *nb.View, err error) {
	resp = &nb.View{}
	query := `
		SELECT
			id,
			table_slug,
			type,
			"columns"
		FROM "view"
		WHERE table_slug = $1
		LIMIT 1
	`

	err = req.Tx.QueryRow(ctx, query, req.TableSlug).Scan(
		&resp.Id,
		&resp.TableSlug,
		&resp.Type,
		&resp.Columns,
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func TabDeleteMany(ctx context.Context, req RelationHelper) error {
	query := `DELETE FROM "tab" WHERE relation_id = $1`
	_, err := req.Tx.Exec(ctx, query, req.RelationID)
	if err != nil {
		return err
	}

	return nil
}

func FieldFindOneDelete(ctx context.Context, req RelationHelper) error {

	query := `
		DELETE FROM "field" WHERE 
			relation_id = $1 AND table_id = $2 AND slug = $3`
	_, err := req.Tx.Exec(ctx, query, req.RelationID, req.TableID, req.FieldName)
	if err != nil {
		return err
	}

	return nil
}

func RemoveRelation(ctx context.Context, req RelationHelper) error {
	switch req.RelationType {
	case config.MANY2ONE:
		query := fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, req.TableFrom, req.FieldName)
		if _, err := req.Tx.Exec(ctx, query); err != nil {
			return err
		}
	case config.MANY2MANY:
		query := fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, req.TableFrom, req.FieldFrom)
		if _, err := req.Tx.Exec(ctx, query); err != nil {
			return err
		}

		query = fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, req.TableTo, req.FieldTo)
		if _, err := req.Tx.Exec(ctx, query); err != nil {
			return err
		}
	case config.RECURSIVE:
		query := fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN %s`, req.TableFrom, req.FieldName)
		if _, err := req.Tx.Exec(ctx, query); err != nil {
			return err
		}
	}

	return nil
}

func ViewFindOneTx(ctx context.Context, req RelationHelper) (resp *nb.View, err error) {
	resp = &nb.View{}
	query := `
		SELECT 
			"id",
			"table_slug",
			"type",
			"name",
			"main_field",
			"disable_dates",
			"columns",
			"quick_filters",
			"users",
			"view_fields",
			"group_fields",
			"calendar_from_slug",
			"calendar_to_slug",
			"time_interval",
			"multiple_insert",
			"status_field_slug",
			"is_editable",
			"relation_table_slug",
			"relation_id",
			"multiple_insert_field",
			"updated_fields",
			"app_id",
			"table_label",
			"default_limit",
			"default_editable",
			"navigate",
			"function_path",
			"order",
			"name_uz",
			"name_en",
			"attributes"
		FROM "view" WHERE relation_id = $1 LIMIT 1`

	var (
		attributes          []byte
		TableSlug           sql.NullString
		Type                sql.NullString
		Name                sql.NullString
		MainField           sql.NullString
		CalendarFromSlug    sql.NullString
		CalendarToSlug      sql.NullString
		TimeInterval        sql.NullInt32
		StatusFieldSlug     sql.NullString
		RelationTableSlug   sql.NullString
		RelationId          sql.NullString
		MultipleInsertField sql.NullString
		AppId               sql.NullString
		TableLabel          sql.NullString
		DefaultLimit        sql.NullString
		FunctionPath        sql.NullString
		Order               sql.NullInt32
		NameUz              sql.NullString
		NameEn              sql.NullString
	)

	err = req.Tx.QueryRow(ctx, query, req.RelationID).Scan(
		&resp.Id,
		&TableSlug,
		&Type,
		&Name,
		&MainField,
		&resp.DisableDates,
		&resp.Columns,
		&resp.QuickFilters,
		&resp.Users,
		&resp.ViewFields,
		&resp.GroupFields,
		&CalendarFromSlug,
		&CalendarToSlug,
		&TimeInterval,
		&resp.MultipleInsert,
		&StatusFieldSlug,
		&resp.IsEditable,
		&RelationTableSlug,
		&RelationId,
		&MultipleInsertField,
		&resp.UpdatedFields,
		&AppId,
		&TableLabel,
		&DefaultLimit,
		&resp.DefaultEditable,
		&resp.Navigate,
		&FunctionPath,
		&Order,
		&NameUz,
		&NameEn,
		&attributes,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(attributes, &resp.Attributes); err != nil {
		return nil, err
	}

	resp = &nb.View{
		Id:                  resp.Id,
		TableSlug:           TableSlug.String,
		Type:                Type.String,
		Name:                Name.String,
		MainField:           MainField.String,
		DisableDates:        resp.DisableDates,
		Columns:             resp.Columns,
		QuickFilters:        resp.QuickFilters,
		Users:               resp.Users,
		ViewFields:          resp.ViewFields,
		GroupFields:         resp.GroupFields,
		CalendarFromSlug:    CalendarFromSlug.String,
		CalendarToSlug:      CalendarToSlug.String,
		TimeInterval:        TimeInterval.Int32,
		MultipleInsert:      resp.MultipleInsert,
		StatusFieldSlug:     StatusFieldSlug.String,
		IsEditable:          resp.IsEditable,
		RelationTableSlug:   RelationTableSlug.String,
		RelationId:          RelationId.String,
		MultipleInsertField: MultipleInsertField.String,
		UpdatedFields:       resp.UpdatedFields,
		AppId:               AppId.String,
		TableLabel:          TableLabel.String,
		DefaultLimit:        DefaultLimit.String,
		DefaultEditable:     resp.DefaultEditable,
		Navigate:            resp.Navigate,
		FunctionPath:        FunctionPath.String,
		Order:               Order.Int32,
		NameUz:              NameUz.String,
		NameEn:              NameEn.String,
		Attributes:          resp.Attributes,
	}
	return resp, nil
}

func RemoveFromLayout(ctx context.Context, req RelationLayout) error {

	tx := req.Tx

	newField := make(map[string]interface{})

	query := `SELECT s.id, s.fields FROM "section" s JOIN "tab" t ON t.id = s.tab_id JOIN "layout" l ON l.id = t.layout_id WHERE l.table_id = $1`

	rows, err := tx.Query(ctx, query, req.TableId)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		field := []map[string]interface{}{}
		newFields := []map[string]interface{}{}
		id := ""
		fieldBody := []byte{}

		err := rows.Scan(
			&id,
			&fieldBody,
		)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(fieldBody, &field); err != nil {
			return err
		}

		fieldLen := 0

		for _, f := range field {
			if strings.Contains(cast.ToString(f["id"]), "#") {

				relationId := strings.Split(cast.ToString(f["id"]), "#")[1]
				if req.RelationId != relationId {

					newFields = append(newFields, f)
					fieldLen++
				}
			} else {

				newFields = append(newFields, f)
				fieldLen++
			}
		}

		newFieldBody, err := json.Marshal(newFields)
		if err != nil {
			return err
		}

		if fieldLen != len(field) {
			newField[id] = newFieldBody
		}
	}

	query = `UPDATE "section" SET fields = $2 WHERE id = $1`

	for id, fields := range newField {
		_, err = tx.Exec(ctx, query, id, fields)
		if err != nil {
			return err
		}
	}

	return nil
}
