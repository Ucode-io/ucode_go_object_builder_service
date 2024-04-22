package helper

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type RelationHelper struct {
	Tx           pgx.Tx
	FieldName    string
	TableID      string
	LayoutID     string
	TableSlug    string
	TabID        string
	Fields       []*nb.FieldForSection
	SectionOrder int
	SectionID    string
	View         *nb.CreateViewRequest
	Field        *nb.CreateFieldRequest
}

func CheckRelationFieldExists(ctx context.Context, req RelationHelper) (bool, string, error) {
	rows, err := req.Tx.Query(ctx, "SELECT slug FROM field WHERE table_id = $1 AND slug LIKE $2 ORDER BY slug DESC", req.TableID, req.FieldName+"%")
	if err != nil {
		return false, "", fmt.Errorf("failed to query fields: %v", err)
	}
	defer rows.Close()

	var lastField string
	for rows.Next() {
		var fieldSlug string
		err := rows.Scan(&fieldSlug)
		if err != nil {
			return false, "", fmt.Errorf("failed to scan field slug: %v", err)
		}
		lastField = fieldSlug
	}

	// If lastField is not empty, extract the index and increment it
	if lastField != "" {
		// Split the slug to extract the index
		parts := strings.Split(lastField, "_")
		if len(parts) > 1 {
			index, err := strconv.Atoi(parts[len(parts)-1])
			if err != nil {
				return false, "", fmt.Errorf("failed to parse index from last field: %v", err)
			}
			// Increment the index
			index++
			lastField = fmt.Sprintf("%s_%d", req.FieldName, index)
		}
	}

	// Return the existence status and the last field name
	return lastField != "", lastField, nil
}

func GetLayoutByTableId(ctx context.Context, req RelationHelper) (resp *nb.LayoutResponse, err error) {
	resp = &nb.LayoutResponse{}
	query := `SELECT 
		"id"
	FROM "layout" WHERE table_id = $1`

	err = req.Tx.QueryRow(ctx, query, req.TableID).Scan(
		&resp.Id,
	)
	if err != nil {
		log.Println("Error while finding layout by table id for relation", err)
		return nil, err
	}
	return resp, nil
}

func TabFindOne(ctx context.Context, req RelationHelper) (resp *nb.TabResponse, err error) {
	resp = &nb.TabResponse{}
	query := `SELECT 
		id
	FROM "tab" 
	WHERE layout_id = $1 AND type = 'section'`

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

	query := `
		INSERT INTO "tab" (
			"id",
			"order",
			"label",
			"type",
			"table_slug",
			"layout_id"
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING "id"
	`
	err = req.Tx.QueryRow(ctx, query,
		id,
		1,
		"Tab",
		"section",
		req.TableSlug,
		req.LayoutID,
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
		id
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
		var sectionID string
		if err := rows.Scan(&sectionID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		resp = append(resp, &nb.Section{Id: sectionID})
	}

	return resp, nil
}

func SectionCreate(ctx context.Context, req RelationHelper) error {

	id := uuid.New().String()

	query := `
		INSERT INTO "section" (
			id,
			"order",
			column,
			label,
			fields,
			table_id,
			tab_id
		) VALUE($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := req.Tx.Exec(ctx, query,
		id,
		req.SectionOrder,
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
		"app_id",
		"table_label",
		"default_limit",
		"default_editable",
		"order",
		"name_uz",
		"name_en",
		"attributes"
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
	`

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
		req.View.AppId,
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

func RelationRoles(ctx context.Context, req RelationHelper) error {

	query := `SELECT guid FROM role`

	rows, err := req.Tx.Query(ctx, query)
	if err != nil {
		// tx.Rollback(ctx)
		return err
	}
	defer rows.Close()

	batch := pgx.Batch{}

	for rows.Next() {
		roleId := ""

		err = rows.Scan(&roleId)
		if err != nil {
			// tx.Rollback(ctx)
			return err
		}

		batch.Queue(`
			INSERT INTO "field_permission" (

			)

		`)
	}

	return nil
}

func UpsertField(ctx context.Context, req RelationHelper) (resp *nb.Field, err error) {
	query := `
		INSERT INTO fields (id, table_id, slug, label, type, relation_id)
		VALUES ($1, $2, $3, $4, $5, $6)
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
	).Scan(&resp.Id)

	if err != nil {
		return nil, fmt.Errorf("failed to insert field: %v", err)
	}

	return resp, nil
}
