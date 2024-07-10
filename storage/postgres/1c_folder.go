package postgres

import (
	"context"
	"database/sql"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type folderGroupRepo struct {
	db *pgxpool.Pool
}

func NewFolderGroupRepo(db *pgxpool.Pool) storage.FolderGroupRepoI {
	return &folderGroupRepo{
		db: db,
	}
}

func (f *folderGroupRepo) Create(ctx context.Context, req *nb.CreateFolderGroupRequest) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.GetProjectId())

	folderGroupId := uuid.NewString()

	query := `INSERT INTO "folder_group" (
		id,
		table_id,
		name,
		comment,
		code
	) VALUES ($1, $2, $3, $4, $5)`

	_, err := conn.Exec(ctx, query, folderGroupId, req.TableId, req.Name, req.Comment, req.Code)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return f.GetByID(ctx, &nb.FolderGroupPrimaryKey{Id: folderGroupId, ProjectId: req.ProjectId})
}

func (f *folderGroupRepo) GetByID(ctx context.Context, req *nb.FolderGroupPrimaryKey) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.ProjectId)

	var (
		id      sql.NullString
		tableId sql.NullString
		name    sql.NullString
		comment sql.NullString
		code    sql.NullString
	)

	query := `
		SELECT
			id,
			table_id,
			name,
			comment,
			code
		FROM folder_group fg
		WHERE fg.id = $1
	`

	err := conn.QueryRow(ctx, query, req.Id).Scan(
		&id,
		&tableId,
		&name,
		&comment,
		&code,
	)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return &nb.FolderGroup{
		Id:      id.String,
		TableId: tableId.String,
		Name:    name.String,
		Comment: comment.String,
		Code:    code.String,
	}, nil
}

func (f *folderGroupRepo) GetAll(ctx context.Context, req *nb.GetAllFolderGroupRequest) (*nb.GetAllFolderGroupResponse, error) {
	var (
		conn = psqlpool.Get(req.GetProjectId())
		resp = &nb.GetAllFolderGroupResponse{}

		query string
	)

	query = `
		SELECT
			id,
			table_id,
			name,
			comment,
			code
		FROM folder_group fg
		WHERE table_id = $1
		OFFSET $2 LIMIT $3
	`

	if req.Limit == 0 {
		req.Limit = 10
	}

	rows, err := conn.Query(ctx, query, req.TableId, req.Offset, req.Limit)
	if err != nil {
		return &nb.GetAllFolderGroupResponse{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id      sql.NullString
			tableId sql.NullString
			name    sql.NullString
			comment sql.NullString
			code    sql.NullString
		)

		err := rows.Scan(
			&id,
			&tableId,
			&name,
			&comment,
			&code,
		)
		if err != nil {
			return &nb.GetAllFolderGroupResponse{}, err
		}

		resp.FolderGroups = append(resp.FolderGroups, &nb.FolderGroup{
			Id:      id.String,
			TableId: tableId.String,
			Name:    name.String,
			Comment: comment.String,
			Code:    code.String,
		})
	}

	return resp, nil
}

func (f *folderGroupRepo) Update(ctx context.Context, req *nb.UpdateFolderGroupRequest) (*nb.FolderGroup, error) {
	conn := psqlpool.Get(req.GetProjectId())

	query := `
		UPDATE folder_group SET
			table_id = $1,
			name = $2,
			comment = $3,
			code = $4,
			updated_at = now()
		WHERE id = $5
	`

	_, err := conn.Exec(ctx, query, req.TableId, req.Name, req.Comment, req.Code, req.Id)
	if err != nil {
		return &nb.FolderGroup{}, err
	}

	return &nb.FolderGroup{
		Id:      req.Id,
		TableId: req.TableId,
		Name:    req.Name,
		Comment: req.Comment,
		Code:    req.Code,
	}, nil
}

func (f *folderGroupRepo) Delete(ctx context.Context, req *nb.FolderGroupPrimaryKey) error {
	conn := psqlpool.Get(req.GetProjectId())

	query := `DELETE FROM folder_group WHERE id = $1`

	_, err := conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}
