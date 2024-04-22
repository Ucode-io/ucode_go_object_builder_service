package helper

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TableVerReq struct {
	Conn *pgxpool.Pool
	Id   string
	Slug string
}

type GetTableByIdSlugReq struct {
	Conn *pgxpool.Pool
	Id   string
	Slug string
}

func TableVer(ctx context.Context, req TableVerReq) (map[string]interface{}, error) {

	query := `SELECT 
			"id",
			"slug"
	 FROM "table" WHERE `

	value := req.Id

	if req.Id != "" {
		query += ` "id" = $1`
	} else if req.Slug != "" {
		query += ` "slug" = $1`
		value = req.Slug
	}

	err := req.Conn.QueryRow(ctx, query, value).Scan(&req.Id, &req.Slug)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":   req.Id,
		"slug": req.Slug,
	}, nil

}

func GetTableByIdSlug(ctx context.Context, req GetTableByIdSlugReq) (map[string]interface{}, error) {

	query := `SELECT id, slug, label FROM "table" WHERE `

	value := req.Id

	var label string

	if req.Id != "" {
		query += ` id = $1`
	} else if req.Slug != "" {
		query += ` slug = $1`
		value = req.Slug
	}

	err := req.Conn.QueryRow(ctx, query, value).Scan(&req.Id, &req.Slug, &label)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":    req.Id,
		"slug":  req.Slug,
		"label": label,
	}, nil
}
