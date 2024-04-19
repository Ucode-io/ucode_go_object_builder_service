package helper

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/structpb"
)

type BoardOrder struct {
	Conn      *pgxpool.Pool
	TableSlug string
}

func BoardOrderChecker(ctx context.Context, req BoardOrder) error {
	var (
		tableId      string
		boardOrderId string

		tableSelectQuery string
		fieldSelectQuery string
		fieldInsertQuery string

		now = time.Now()
	)

	tableSelectQuery = `SELECT id FROM "table" WHERE slug = $1`
	err := req.Conn.QueryRow(ctx, tableSelectQuery, req.TableSlug).Scan(&tableId)
	if err != nil {
		return err
	}

	fieldSelectQuery = `SELECT id FROM "field" WHERE table_id = $1 AND "slug" = 'board_order'`
	err = req.Conn.QueryRow(ctx, fieldSelectQuery, tableId).Scan(&boardOrderId)
	if err == pgx.ErrNoRows {
		attributes := &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"icon":        {Kind: &structpb.Value_StringValue{StringValue: ""}},
				"placeholder": {Kind: &structpb.Value_StringValue{StringValue: ""}},
				"showTooltip": {Kind: &structpb.Value_StringValue{StringValue: ""}},
			},
		}
		attributesJson, err := json.Marshal(attributes)
		if err != nil {
			return err
		}

		fieldInsertQuery = `INSERT INTO "field" (id, table_id, required, slug, label, "default", "type", "index", attributes, is_visible, autofill_field, autofill_table, created_at, updated_at)
						  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

		_, err = req.Conn.Exec(
			ctx,
			fieldInsertQuery,
			"93999892-78b0-4674-9e42-6a2a41524ebe",
			tableId,
			false,
			"board_order",
			"BOARD ORDER",
			"",
			"NUMBER",
			"string",
			attributesJson,
			false, "", "", now, now)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}
