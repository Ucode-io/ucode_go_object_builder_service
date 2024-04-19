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
			"slug",
			"label",
			"description",
			"show_in_menu",
			"subtitle_field_slug",
			"is_cached",
			"with_increment_id",
			"soft_delete",
			"digit_number"
	 FROM "table" WHERE `

	value := req.Id

	var (
		label             string
		description       string
		showInMenu        bool
		subtitleFieldSlug string
		isCached          bool
		withIncrementId   bool
		softDelete        bool
		digitNumber       int32
	)

	if req.Id != "" {
		query += ` "id" = $1`
	} else if req.Slug != "" {
		query += ` "slug" = $1`
		value = req.Slug
	}

	err := req.Conn.QueryRow(ctx, query, value).Scan(&req.Id, &req.Slug, &label)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return map[string]interface{}{
		"id":                  req.Id,
		"slug":                req.Slug,
		"label":               label,
		"description":         description,
		"show_in_menu":        showInMenu,
		"subtitle_field_slug": subtitleFieldSlug,
		"is_cached":           isCached,
		"with_increment_id":   withIncrementId,
		"soft_delete":         softDelete,
		"digit_number":        digitNumber,
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
