package helper

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xtgo/uuid"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
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

func TableFindOne(ctx context.Context, conn *pgxpool.Pool, id string) (resp *nb.Table, err error) {
	var (
		filter string = "id = $1"
	)
	resp = &nb.Table{
		IncrementId: &nb.IncrementID{},
	}

	_, err = uuid.Parse(id)
	if err != nil {
		filter = "slug = $1"
	}

	query := `SELECT

		"id",
		"slug",
		"label",
		"section_column_count"
	FROM "table" WHERE ` + filter

	err = conn.QueryRow(ctx, query, id).Scan(
		&resp.Id,
		&resp.Slug,
		&resp.Label,
		&resp.SectionColumnCount,
	)
	if err != nil {
		log.Println("Error while finding single table", err)
		return nil, err
	}
	return resp, nil
}
