package postgres

import (
	"context"
	"fmt"
	"log"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type fileRepo struct {
	db *pgxpool.Pool
}

func NewFileRepo(db *pgxpool.Pool) fileRepo {
	return fileRepo{
		db: db,
	}
}

func (f *fileRepo) Create(ctx context.Context, req *nb.CreateFileRequest) (resp *nb.File, err error) {

	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	fileId := uuid.NewString()

	query := `INSERT INTO "file" (
		"id",
		"title",
		"description",
		"tags",
		"storage",
		"file_name_disk", 
		"file_name_download",
		"link",
		"file_size"
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = conn.Exec(ctx, query,
		fileId,
		req.Title,
		req.Description,
		req.Tags,
		req.Storage,
		req.FileNameDisk,
		req.FileNameDownload,
		req.Link,
		req.FileSize,
	)
	if err != nil {
		return &nb.File{}, err
	}

	return f.GetSingle(ctx, &nb.FilePrimaryKey{Id: fileId, ProjectId: req.ProjectId})
}

func (f *fileRepo) GetSingle(ctx context.Context, req *nb.FilePrimaryKey) (resp *nb.File, err error) {

	resp = &nb.File{}
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `SELECT 
				"id",
				"title",
				"description",
				"tags",
				"storage",
				"file_name_disk", 
				"file_name_download",
				"link",
				"file_size"
		FROM "file" WHERE id = $1`

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.Title,
		&resp.Description,
		&resp.Tags,
		&resp.Storage,
		&resp.FileNameDisk,
		&resp.FileNameDownload,
		&resp.Link,
		&resp.FileSize,
	)
	if err != nil {
		return &nb.File{}, err
	}

	return resp, nil
}

func (f *fileRepo) GetList(ctx context.Context, req *nb.GetAllFilesRequest) (resp *nb.GetAllFilesResponse, err error) {
	resp = &nb.GetAllFilesResponse{}

	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `SELECT 
				COUNT(*) OVER(),
				"id",
				"title",
				"description",
				"tags",
				"storage",
				"file_name_disk", 
				"file_name_download",
				"link",
				"file_size"
			FROM "file"  
			WHERE 1=1
	`

	if req.Search != "" {
		query += fmt.Sprintf(` AND title ~* '%s'`, req.Search)
	}

	if req.Sort != "" {
		query += ` ORDER BY created_at ` + req.GetSort()
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		row := &nb.File{}

		err = rows.Scan(
			&resp.Count,
			&row.Id,
			&row.Title,
			&row.Description,
			&row.Tags,
			&row.Storage,
			&row.FileNameDisk,
			&row.FileNameDownload,
			&row.Link,
			&row.FileSize,
		)
		if err != nil {
			return resp, err
		}

		resp.Files = append(resp.Files, row)
	}

	return resp, nil
}

func (f *fileRepo) Update(ctx context.Context, req *nb.File) error {
	conn := psqlpool.Get(req.ProjectId)

	defer conn.Close()

	query := `UPDATE "file" SET
				"title" = $2,
				"description" = $3,
				"tags" = $4,
				"storage" = $5 , 
				"file_name_disk" =  $6,
				"file_name_download" = $7,
				"link" = $8,
				"file_size" = $9,
				"updated_at" = CURRENT_TIMESTAMP
			WHERE id = $1
	`

	_, err := conn.Exec(ctx, query,
		req.Id,
		req.Title,
		req.Description,
		req.Tags,
		req.Storage,
		req.FileNameDisk,
		req.FileNameDownload,
		req.Link,
		req.FileSize,
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *fileRepo) Delete(ctx context.Context, req *nb.FileDeleteRequest) error {

	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `
        DELETE FROM "file"
        WHERE id = ANY($1)
    `
	ids := make([]interface{}, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = id
	}

	_, err := conn.Exec(ctx, query, ids)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
