package postgres

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

	args = append(args, req.GetId())

	var query = fmt.Sprintf(`
		UPDATE mcp_project
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), argIndex)

	log.Println("ARGS:", args)

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
		return nil, err
	}

	var (
		baseQuery = `FROM mcp_project WHERE 1=1`
		args      = []any{}
		argIndex  = 1
	)

	if req.GetTitle() != "" {
		baseQuery += fmt.Sprintf(" AND title ILIKE $%d", argIndex)
		args = append(args, "%"+req.GetTitle()+"%")
		argIndex++
	}

	var (
		dataQuery = fmt.Sprintf(`
       SELECT id, title, description, project_env, created_at, updated_at
       %s
       ORDER BY created_at DESC
       LIMIT $%d OFFSET $%d
    `, baseQuery, argIndex, argIndex+1)

		projects []*nb.McpProject
	)

	args = append(args, req.GetLimit(), req.GetOffset())

	rows, err := conn.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			project    nb.McpProject
			projectEnv = make(map[string]any)
			createdAt  time.Time
			updatedAt  time.Time
		)

		err = rows.Scan(
			&project.Id,
			&project.Title,
			&project.Description,
			&projectEnv,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		project.CreatedAt = createdAt.Format(time.RFC3339)
		project.UpdatedAt = updatedAt.Format(time.RFC3339)

		fileGraphStruct, err := structpb.NewStruct(projectEnv)
		if err != nil {
			return nil, fmt.Errorf("failed to convert file_graph to struct: %w", err)
		}

		project.ProjectEnv = fileGraphStruct
		projects = append(projects, &project)
	}

	return &nb.McpProjectList{
		ResourceEnvId: req.GetResourceEnvId(),
		Projects:      projects,
	}, nil
}

func (m *mcpProjectRepo) GetMcpProjectFiles(ctx context.Context, req *nb.McpProjectId) (*nb.McpProject, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "mcp_project.GetMcpProjectFiles")
	defer dbSpan.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		projectQuery = `
       SELECT id, title, description, project_env, created_at, updated_at
       FROM mcp_project
       WHERE id = $1
    `

		project nb.McpProject

		createdAt time.Time
		updatedAt time.Time

		projectEnv = make(map[string]any)
	)

	err = conn.QueryRow(ctx, projectQuery, req.GetId()).Scan(
		&project.Id,
		&project.Title,
		&project.Description,
		&projectEnv,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	projectEnvStruct, err := structpb.NewStruct(projectEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to convert file_graph to struct: %w", err)
	}

	project.ProjectEnv = projectEnvStruct

	project.CreatedAt = createdAt.Format(time.RFC3339)
	project.UpdatedAt = updatedAt.Format(time.RFC3339)

	var (
		filesQuery = `
          SELECT id, project_id, file_path, content, file_graph, created_at, updated_at
          FROM project_files
          WHERE project_id = $1
          ORDER BY created_at ASC
    `
		files []*nb.McpProjectFiles
	)

	rows, err := conn.Query(ctx, filesQuery, req.GetId())
	if err != nil {
		return nil, fmt.Errorf("failed to query project files: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			file      nb.McpProjectFiles
			fileGraph map[string]interface{}

			fCreatedAt time.Time
			fUpdatedAt time.Time
		)

		err = rows.Scan(
			&file.Id,
			&file.ProjectId,
			&file.Path,
			&file.Content,
			&fileGraph,
			&fCreatedAt,
			&fUpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan project file: %w", err)
		}

		file.CreatedAt = fCreatedAt.Format(time.RFC3339)
		file.UpdatedAt = fUpdatedAt.Format(time.RFC3339)

		fileGraphStruct, err := structpb.NewStruct(fileGraph)
		if err != nil {
			return nil, fmt.Errorf("failed to convert file_graph to struct: %w", err)
		}

		file.FileGraph = fileGraphStruct

		files = append(files, &file)
	}

	project.ProjectFiles = files

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
