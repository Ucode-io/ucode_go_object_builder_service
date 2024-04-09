package postgres

import (
	"context"
	"fmt"
	"time"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type functionRepo struct {
	db *pgxpool.Pool
}

func NewFunctionRepo(db *pgxpool.Pool) storage.FunctionRepoI {
	return &functionRepo{
		db: db,
	}
}

func (f *functionRepo) Create(ctx context.Context, req *nb.CreateFunctionRequest) (resp *nb.Function, err error) {

	conn := psqlpool.Get(req.ProjectId)

	functionId := uuid.NewString()

	query := `INSERT INTO "function" (
		id,
		name,
		path,
		type,
		description,
		project_id,
		environment_id,
		function_folder_id,
		url,
		password,
		ssh_url,
		gitlab_id,
		gitlab_group_id,
		request_time
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err = conn.Exec(ctx, query,
		functionId,
		req.Name,
		req.Path,
		req.Type,
		req.Description,
		req.ProjectId,
		req.EnvironmentId,
		req.FunctionFolderId,
		req.Url,
		req.Password,
		req.SshUrl,
		req.GitlabId,
		req.GitlabGroupId,
		time.Now().Format(time.RFC3339),
	)
	if err != nil {
		return &nb.Function{}, err
	}

	return f.GetSingle(ctx, &nb.FunctionPrimaryKey{Id: functionId, ProjectId: req.ProjectId})
}

func (f *functionRepo) GetList(ctx context.Context, req *nb.GetAllFunctionsRequest) (resp *nb.GetAllFunctionsResponse, err error) {

	conn := psqlpool.Get(req.ProjectId)

	query := fmt.Sprintf(`SELECT 
		id,
		name,
		path,
		type,
		description,
		project_id,
		environment_id,
		function_folder_id,
		url,
		password,
		ssh_url,
		gitlab_id,
		gitlab_group_id
	FROM "function" WHERE type = '%s' 
	`, req.Type)

	if req.Search != "" {
		query += fmt.Sprintf(` AND name ~* '%s'`, req.Search)
	}

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return &nb.GetAllFunctionsResponse{}, err
	}

	for rows.Next() {
		row := &nb.Function{}

		err = rows.Scan(
			&row.Id,
			&row.Name,
			&row.Path,
			&row.Type,
			&row.Description,
			&row.ProjectId,
			&row.EnvironmentId,
			&row.FunctionFolderId,
			&row.Url,
			&row.Password,
			&row.SshUrl,
			&row.GitlabId,
			&row.GitlabGroupId,
		)
		if err != nil {
			return &nb.GetAllFunctionsResponse{}, err
		}

		resp.Functions = append(resp.Functions, row)
	}

	return resp, nil
}

func (f *functionRepo) GetSingle(ctx context.Context, req *nb.FunctionPrimaryKey) (resp *nb.Function, err error) {

	conn := psqlpool.Get(req.ProjectId)

	query := `SELECT 
		id,
		name,
		path,
		type,
		description,
		project_id,
		environment_id,
		function_folder_id,
		url,
		password,
		ssh_url,
		gitlab_id,
		gitlab_group_id
	FROM "function" WHERE id = $1`

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.Name,
		&resp.Path,
		&resp.Type,
		&resp.Description,
		&resp.ProjectId,
		&resp.EnvironmentId,
		&resp.FunctionFolderId,
		&resp.Url,
		&resp.Password,
		&resp.SshUrl,
		&resp.GitlabId,
		&resp.GitlabGroupId,
	)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (f *functionRepo) Update(ctx context.Context, req *nb.Function) error {
	conn := psqlpool.Get(req.ProjectId)

	query := `UPDATE "function" SET
		name = $2,
		path = $3,
		type = $4,
		description = $5,
		project_id = $6,
		environment_id = $7,
		function_folder_id = $8,
		url = $9,
		password = $10,
		ssh_url = $11,
		gitlab_id = $12,
		gitlab_group_id = $13
	WHERE id = $1
	`

	_, err := conn.Exec(ctx, query,
		req.Id,
		req.Name,
		req.Path,
		req.Type,
		req.Description,
		req.ProjectId,
		req.EnvironmentId,
		req.FunctionFolderId,
		req.Url,
		req.Password,
		req.SshUrl,
		req.GitlabId,
		req.GitlabGroupId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *functionRepo) Delete(ctx context.Context, req *nb.FunctionPrimaryKey) error {

	conn := psqlpool.Get(req.ProjectId)

	query := `DELETE FROM "function" WHERE id = $1`

	_, err := conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}
