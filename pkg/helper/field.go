package helper

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GetFieldBySlugReq struct {
	Conn    *pgxpool.Pool
	Slug    string
	TableId string
}

func GetFieldBySlug(ctx context.Context, req GetFieldBySlugReq) (map[string]interface{}, error) {

	query := `SELECT id, type, attributes FROM "field" WHERE slug = $1 AND table_id = $2`

	var (
		id, ftype  string
		attributes []byte
	)

	err := req.Conn.QueryRow(ctx, query, req.Slug, req.TableId).Scan(&id, &req.Slug)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":         id,
		"type":       ftype,
		"attributes": attributes,
	}, nil
}
