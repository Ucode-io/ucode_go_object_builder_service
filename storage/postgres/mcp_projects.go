package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"ucode/ucode_go_object_builder_service/pkg/util"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/structpb"
)

type mcpProjectRepo struct {
	db *psqlpool.Pool
}

func NewMcpProjectRepo(db *psqlpool.Pool) storage.McpProjectRepoI {
	return &mcpProjectRepo{
		db: db,
	}
}

func (m *mcpProjectRepo) CreateMcpProject(ctx context.Context, req *nb.CreateMcpProjectReqeust) (*nb.McpProject, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.CreateMcpProject")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var (
		projectId = uuid.NewString()
		now       = time.Now()

		projectQuery = `
			INSERT INTO mcp_project (id, title, description, project_env, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $5)`
	)

	_, err = tx.Exec(ctx, projectQuery, projectId, req.GetTitle(), req.GetDescription(), req.ProjectEnv.AsMap(), now)
	if err != nil {
		return nil, fmt.Errorf("failed to insert mcp_project: %w", err)
	}

	if len(req.GetProjectFiles()) > 0 {
		var (
			valueStrings = make([]string, 0, len(req.GetProjectFiles()))
			valueArgs    = make([]interface{}, 0, len(req.GetProjectFiles())*7)

			argIndex = 1
		)

		for _, file := range req.GetProjectFiles() {
			var (
				fileId       = uuid.NewString()
				fileGraphMap = file.GetFileGraph().AsMap()
			)

			valueStrings = append(valueStrings, fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				argIndex, argIndex+1, argIndex+2, argIndex+3, argIndex+4, argIndex+5, argIndex+6,
			))

			valueArgs = append(valueArgs,
				fileId,
				projectId,
				file.GetPath(),
				file.GetContent(),
				fileGraphMap,
				now,
				now,
			)

			argIndex += 7
		}

		fileQuery := fmt.Sprintf(`
			INSERT INTO project_files (id, project_id, file_path, content, file_graph, created_at, updated_at)
			VALUES %s
		`, strings.Join(valueStrings, ", "))

		_, err = tx.Exec(ctx, fileQuery, valueArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to insert project_files: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	result, err := m.GetMcpProjectFiles(ctx,
		&nb.McpProjectId{
			ResourceEnvId: req.GetResourceEnvId(),
			Id:            projectId,
		},
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *mcpProjectRepo) UpdateMcpProject(ctx context.Context, req *nb.McpProject) (*nb.McpProject, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.UpdateMcpProject")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err = m.updateProjectFields(ctx, tx, req); err != nil {
		return nil, err
	}

	if err = m.upsertProjectFiles(ctx, tx, req.GetId(), req.GetProjectFiles()); err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return m.GetMcpProjectFiles(ctx, &nb.McpProjectId{
		ResourceEnvId: req.GetResourceEnvId(),
		Id:            req.GetId(),
	})
}

func (m *mcpProjectRepo) updateProjectFields(ctx context.Context, tx pgx.Tx, req *nb.McpProject) error {
	var (
		setClauses = []string{"updated_at = $1"}
		args       = []any{time.Now()}
		argIndex   = 2
	)

	if req.GetTitle() != "" {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, req.GetTitle())
		argIndex++
	}

	if req.GetDescription() != "" {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.GetDescription())
		argIndex++
	}

	if util.IsValidUUID(req.GetFunctionId()) {
		setClauses = append(setClauses, fmt.Sprintf("function_id = $%d", argIndex))
		args = append(args, req.GetFunctionId())
		argIndex++
	}

	args = append(args, req.GetId())

	var query = fmt.Sprintf(`
		UPDATE mcp_project
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update mcp_project: %w", err)
	}

	return nil
}

func (m *mcpProjectRepo) upsertProjectFiles(ctx context.Context, tx pgx.Tx, projectId string, files []*nb.McpProjectFiles) error {
	if len(files) == 0 {
		return nil
	}

	var (
		ids        []string
		filePaths  []string
		contents   []string
		fileGraphs []map[string]interface{}
	)

	for _, file := range files {
		fileId := file.GetId()
		if fileId == "" {
			fileId = uuid.NewString()
		}

		ids = append(ids, fileId)
		filePaths = append(filePaths, file.GetPath())
		contents = append(contents, file.GetContent())
		fileGraphs = append(fileGraphs, file.GetFileGraph().AsMap())
	}

	var query = `
       INSERT INTO project_files (id, project_id, file_path, content, file_graph, created_at, updated_at)
       SELECT 
          unnest($1::uuid[]),
          $2::uuid,
          unnest($3::text[]),
          unnest($4::text[]),
          unnest($5::jsonb[]),
          $6::timestamp,
          $6::timestamp
       
       ON CONFLICT (project_id, file_path) WHERE deleted_at IS NULL
       
       DO UPDATE SET
          content = EXCLUDED.content,
          file_graph = EXCLUDED.file_graph,
          updated_at = EXCLUDED.updated_at
    `

	_, err := tx.Exec(ctx, query,
		ids,
		projectId,
		filePaths,
		contents,
		fileGraphs,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert project_files: %w", err)
	}

	return nil
}

func (m *mcpProjectRepo) GetAllMcpProject(ctx context.Context, req *nb.GetMcpProjectListReq) (*nb.McpProjectList, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.GetAllMcpProject")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var (
		queryBuilder strings.Builder
		args         = make([]any, 0)
		projects     = make([]*nb.McpProject, 0)
	)

	queryBuilder.WriteString(`
        SELECT mp.id, mp.title, mp.description, mp.project_env, mp.created_at, mp.updated_at,
               f.id, f.name, f.path, f.type, f.url, f.branch, f.repo_id,
               f.created_at, f.updated_at
        FROM mcp_project AS mp
        LEFT JOIN function AS f ON mp.function_id = f.id
        WHERE 1=1
    `)

	if req.GetTitle() != "" {
		args = append(args, "%"+req.GetTitle()+"%")
		queryBuilder.WriteString(fmt.Sprintf(" AND mp.title ILIKE $%d", len(args)))
	}

	args = append(args, req.GetLimit(), req.GetOffset())
	queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args)))

	rows, err := conn.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return &nb.McpProjectList{}, nil
		}
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			project              nb.McpProject
			projectEnv           map[string]any
			createdAt, updatedAt time.Time

			fId, fName, fPath, fType, fUrl, fBranch, fRepoId sql.NullString
			fCreatedAt, fUpdatedAt                           sql.NullTime
		)

		project.FunctionData = &nb.FunctionData{}

		err = rows.Scan(
			&project.Id, &project.Title, &project.Description, &projectEnv, &createdAt, &updatedAt,
			&fId, &fName, &fPath, &fType, &fUrl, &fBranch, &fRepoId, &fCreatedAt, &fUpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		project.CreatedAt = createdAt.Format(time.RFC3339)
		project.UpdatedAt = updatedAt.Format(time.RFC3339)

		if projectEnv != nil {
			envStruct, _ := structpb.NewStruct(projectEnv)
			project.ProjectEnv = envStruct
		}

		if fId.Valid {
			project.FunctionId = fId.String
			project.FunctionData.Id = fId.String
			project.FunctionData.Name = fName.String
			project.FunctionData.Path = fPath.String
			project.FunctionData.Type = fType.String
			project.FunctionData.Url = fUrl.String
			project.FunctionData.Branch = fBranch.String
			project.FunctionData.RepoId = fRepoId.String
			if fCreatedAt.Valid {
				project.FunctionData.CreatedAt = fCreatedAt.Time.Format(time.RFC3339)
			}
			if fUpdatedAt.Valid {
				project.FunctionData.UpdatedAt = fUpdatedAt.Time.Format(time.RFC3339)
			}
		}

		projects = append(projects, &project)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &nb.McpProjectList{
		Projects: projects,
	}, nil
}

func (m *mcpProjectRepo) GetMcpProjectFiles(ctx context.Context, req *nb.McpProjectId) (*nb.McpProject, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.GetMcpProjectFiles")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, fmt.Errorf("failed to get connection: %w", err)
	}

	var (
		project                nb.McpProject
		projectEnv             map[string]any
		pCreatedAt, pUpdatedAt time.Time

		fId, fName, fPath, fType, fUrl, fBranch, fRepoId sql.NullString
		fCreatedAt, fUpdatedAt                           sql.NullTime

		projectQuery = `
        	SELECT 
            	mp.id, mp.title, mp.description, mp.project_env, mp.created_at, mp.updated_at,
            	f.id, f.name, f.path, f.type, f.url, f.branch, f.repo_id, f.created_at, f.updated_at
        	FROM mcp_project mp
        	LEFT JOIN function f ON mp.function_id = f.id
        	WHERE mp.id = $1
    `
	)

	err = conn.QueryRow(ctx, projectQuery, req.GetId()).Scan(
		&project.Id, &project.Title, &project.Description, &projectEnv, &pCreatedAt, &pUpdatedAt,
		&fId, &fName, &fPath, &fType, &fUrl, &fBranch, &fRepoId, &fCreatedAt, &fUpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", req.GetId())
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	project.CreatedAt = pCreatedAt.Format(time.RFC3339)
	project.UpdatedAt = pUpdatedAt.Format(time.RFC3339)
	project.ResourceEnvId = req.GetResourceEnvId()

	if projectEnv != nil {
		project.ProjectEnv, _ = structpb.NewStruct(projectEnv)
	}

	if fId.Valid {
		project.FunctionId = fId.String
		project.FunctionData = &nb.FunctionData{
			Id:     fId.String,
			Name:   fName.String,
			Path:   fPath.String,
			Type:   fType.String,
			Url:    fUrl.String,
			Branch: fBranch.String,
			RepoId: fRepoId.String,
		}
		if fCreatedAt.Valid {
			project.FunctionData.CreatedAt = fCreatedAt.Time.Format(time.RFC3339)
		}
		if fUpdatedAt.Valid {
			project.FunctionData.UpdatedAt = fUpdatedAt.Time.Format(time.RFC3339)
		}
	}

	if !req.GetWithoutFiles() {
		var filesQuery = `
            SELECT id, project_id, file_path, content, file_graph, created_at, updated_at
            FROM project_files
            WHERE project_id = $1
            ORDER BY created_at ASC
        `
		rows, err := conn.Query(ctx, filesQuery, req.GetId())
		if err != nil {
			return nil, fmt.Errorf("failed to query project files: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var (
				file                 nb.McpProjectFiles
				fileGraph            map[string]any
				createdAt, updatedAt time.Time
			)

			err = rows.Scan(
				&file.Id, &file.ProjectId, &file.Path, &file.Content, &fileGraph, &createdAt, &updatedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to scan project file: %w", err)
			}

			file.CreatedAt = createdAt.Format(time.RFC3339)
			file.UpdatedAt = updatedAt.Format(time.RFC3339)
			file.ResourceEnvId = req.GetResourceEnvId()

			if fileGraph != nil {
				file.FileGraph, _ = structpb.NewStruct(fileGraph)
			}

			project.ProjectFiles = append(project.ProjectFiles, &file)
		}

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("rows iteration error: %w", err)
		}
	}

	return &project, nil
}

func (m *mcpProjectRepo) DeleteMcpProject(ctx context.Context, req *nb.McpProjectId) error {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.DeleteMcpProject")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var deleteFilesQuery = `DELETE FROM project_files WHERE project_id = $1`
	_, err = tx.Exec(ctx, deleteFilesQuery, req.GetId())
	if err != nil {
		return fmt.Errorf("failed to delete project files: %w", err)
	}

	var deleteProjectQuery = `DELETE FROM mcp_project WHERE id = $1`
	_, err = tx.Exec(ctx, deleteProjectQuery, req.GetId())
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

//func (m *mcpProjectRepo) DeleteProjectFile(ctx context.Context, req *nb.McpProjectId) error {
//	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.DeleteProjectFile")
//	defer dbSpan.Finish()
//
//	conn, err := psqlpool.Get(req.GetResourceEnvId())
//	if err != nil {
//		return err
//	}
//
//	query := `DELETE FROM project_files WHERE project_id = $1`
//	args := []interface{}{req.GetProjectId()}
//	argIndex := 2
//
//	// Можем удалять либо по file_id, либо по file_path
//	if req.GetFileId() != "" {
//		query += fmt.Sprintf(" AND id = $%d", argIndex)
//		args = append(args, req.GetFileId())
//	} else if req.GetPath() != "" {
//		query += fmt.Sprintf(" AND file_path = $%d", argIndex)
//		args = append(args, req.GetPath())
//	} else {
//		return fmt.Errorf("either file_id or file_path must be provided")
//	}
//
//	_, err = conn.Exec(ctx, query, args...)
//	if err != nil {
//		return fmt.Errorf("failed to delete project file: %w", err)
//	}
//
//	return nil
//}
