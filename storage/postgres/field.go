package postgres

import (
	"context"
	"encoding/json"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func Create(ctx context.Context, req *nb.CreateFieldRequest) error {

	conn := psqlpool.Get(req.ProjectId)
	fieldId := uuid.NewString()

	query := `INSERT INTO "field" (
		"required",
		"slug",
		"label",
		"default",
		"type",
		"index",
		"is_visible",
		"table_id",
		"commit_id",
		"attributes",
		id
	) VALUES (
		$1, $2, $3, $4,$5,$6, $7, $8, $9, $10, $11
	)`

	attributes, err := json.Marshal(req.Attributes)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, query,
		false,
		req.Slug,
		req.Label,
		req.Default,
		req.Type,
		req.Index,
		req.IsVisible,
		req.TableId,
		req.CommitId,
		attributes,
		fieldId,
	)
	if err != nil {
		return err
	}

	query = `SELECT is_changed_by_host, slug FROM "table" where id = $1`

	var (
		data          = []byte{}
		tableSlug     string
		layoutId      string
		tabId         string
		sectionId     string
		sectionCount  int32
		sectionFields int32
	)

	err = conn.QueryRow(ctx, query, req.TableId).Scan(&data, &tableSlug)
	if err != nil {
		return err
	}

	data, err = helper.ChangeHostname(data)
	if err != nil {
		return err
	}

	query = `UPDATE "table" SET 
		is_changed = true,
		is_changed_by_host = $1
	`

	_, err = conn.Exec(ctx, query, data)
	if err != nil {
		return err
	}

	query = `SELECT guid FROM "role"`

	row, err := conn.Query(ctx, query)
	if err != nil {
		return err
	}

	query = `INSERT INTO "field_permission" (
		"edit_permission",
		"view_permission",
		"table_slug",
		"field_id",
		"label",
		role_id
	) VALUES (true, true, $1, $2, $3, $4)`

	for row.Next() {
		id := ""

		err := row.Scan(&id)
		if err != nil {
			return err
		}

		_, err = conn.Exec(ctx, query, tableSlug, fieldId, req.Label, id)
		if err != nil {
			return err
		}
	}

	query = `SELECT id FROM "layout" WHERE table_id = $1`
	err = conn.QueryRow(ctx, query, req.TableId).Scan(&layoutId)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	query = `SELECT id FROM "tab" WHERE section_id = $1 and type = 'section'`
	err = conn.QueryRow(ctx, query, layoutId).Scan(&tabId)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	query = `SELECT id FROM "section" WHERE tab_id = $1 ORDER BY created_at DESC LIMIT 1`
	err = conn.QueryRow(ctx, query, tabId).Scan(&sectionId)
	if err != nil {
		return err
	}

	queryCount := `SELECT COUNT(*) FROM "section" WHERE tab_id = $1`
	err = conn.QueryRow(ctx, queryCount, tabId).Scan(&sectionCount)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	query = `SELECT COUNT(*) FROM "section_fields" WHERE section_id = $1`
	err = conn.QueryRow(ctx, queryCount, tabId).Scan(&sectionFields)
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if sectionFields < 3 {
		query := `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

		_, err = conn.Exec(ctx, query, fieldId, sectionFields+1, req.Label, sectionId)
		if err != nil {
			return err
		}
	} else {
		query = `INSERT INTO "section" (id, order, column, label, table_id, tab_id) VALUES ($1, $2, $3, $4, $5, $6)`

		sectionId = uuid.NewString()

		_, err = conn.Exec(ctx, query, sectionId, sectionCount+1, "SINGLE", "Info", req.TableId, tabId)
		if err != nil {
			return err
		}

		query = `INSERT INTO "section_fields" (id, order, field_name, section_id) VALUES ($1, $2, $3, $4)`

		_, err = conn.Exec(ctx, query, fieldId, 1, req.Label, sectionId)
		if err != nil {
			return err
		}
	}

	return nil
}
