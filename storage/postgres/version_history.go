package postgres

import (
	"context"
	"fmt"
	"strings"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type versionHistoryRepo struct {
	db *pgxpool.Pool
}

func NewVersionHistoryRepo(db *pgxpool.Pool) storage.VersionHistoryRepoI {
	return &versionHistoryRepo{
		db: db,
	}
}

func (v *versionHistoryRepo) GetById(ctx context.Context, req *nb.VersionHistoryPrimaryKey) (*nb.VersionHistory, error) {
	conn := psqlpool.Get(req.ProjectId)

	query := `
		SELECT 
			id, 
			action_source, 
			action_type, 
			previous, 
			current, 
			date, 
			user_info, 
			request, 
			response, 
			api_key, 
			type, 
			table_slug
		FROM version_history 
		WHERE id = $1
	`

	var (
		history = &nb.VersionHistory{}
	)
	err := conn.QueryRow(ctx, query, req.Id).Scan(
		&history.Id,
		&history.ActionSource,
		&history.ActionType,
		&history.Previus,
		&history.Current,
		&history.Date,
		&history.UserInfo,
		&history.Request,
		&history.Response,
		&history.ApiKey,
		&history.Type,
		&history.TableSlug,
	)
	if err != nil {
		return &nb.VersionHistory{}, err
	}

	return history, nil
}

func (v *versionHistoryRepo) GetAll(ctx context.Context, req *nb.GetAllRquest) (*nb.ListVersionHistory, error) {
	conn := psqlpool.Get(req.GetProjectId())

	query := `
		SELECT 
			id, 
			action_source, 
			action_type, 
			previous, 
			current, 
			date, 
			user_info, 
			request, 
			response, 
			api_key, 
			type, 
			table_slug
			--used_environments
		FROM version_history WHERE true
	`
	args := []interface{}{}
	argIndex := 1

	if req.Type == "DOWN" || req.Type == "UP" {
		query += fmt.Sprintf(" AND action_source IN (%s)", "'RELATION', 'FIELD', 'MENU', 'TABLE', 'LAYOUT', 'VIEW'")
	} else if req.Type != "" {
		query += fmt.Sprintf(" AND type = $%d", argIndex)
		args = append(args, req.Type)
		argIndex++
	}

	if req.FromDate != "" {
		query += fmt.Sprintf(" AND date >= $%d", argIndex)
		args = append(args, req.FromDate)
		argIndex++
	}
	if req.ToDate != "" {
		query += fmt.Sprintf(" AND date <= $%d", argIndex)
		args = append(args, req.ToDate)
		argIndex++
	}
	if req.UserInfo != "" {
		query += fmt.Sprintf(" AND user_info = $%d", argIndex)
		args = append(args, req.UserInfo)
		argIndex++
	}
	if req.ApiKey != "" {
		query += fmt.Sprintf(" AND api_key = $%d", argIndex)
		args = append(args, req.ApiKey)
	}

	sortOrder := "DESC"
	if req.OrderBy {
		sortOrder = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY date %s LIMIT %d OFFSET %d", sortOrder, req.Limit, req.Offset)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		histories = []*nb.VersionHistory{}
	)
	for rows.Next() {
		var history nb.VersionHistory
		if err := rows.Scan(
			&history.Id,
			&history.ActionSource,
			&history.ActionType,
			&history.Previus,
			&history.Current,
			&history.Date,
			&history.UserInfo,
			&history.Request,
			&history.Response,
			&history.ApiKey,
			&history.Type,
			&history.TableSlug,
		); err != nil {
			return nil, err
		}

		histories = append(histories, &history)
	}

	resp := &nb.ListVersionHistory{}
	resp.Histories = histories

	countQuery := `SELECT COUNT(*) FROM version_history`
	err = conn.QueryRow(ctx, countQuery).Scan(&resp.Count)
	if err != nil {
		return &nb.ListVersionHistory{}, err
	}

	return resp, nil
}

func (v *versionHistoryRepo) Update(ctx context.Context, req *nb.UsedForEnvRequest) error {
	conn := psqlpool.Get(req.ProjectId)

	query := `
		UPDATE object_builder_service.version_history
		SET used_environments = jsonb_set(
			used_environments, 
			ARRAY[$1], 
			'true'::jsonb, 
			true
		)
		WHERE id = ANY($2)
	`

	_, err := conn.Exec(ctx, query, req.EnvId, req.Ids)
	if err != nil {
		return err
	}

	return nil
}

func (v *versionHistoryRepo) Create(ctx context.Context, req *nb.CreateVersionHistoryRequest) (err error) {
	conn := psqlpool.Get(req.ProjectId)

	versionH := `INSERT INTO version_history (
		action_source,
		action_type,
		previous,
		current,
		date,
		user_info,
		request,
		response,
		api_key,
		type,
		table_slug
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	query := `SELECT label FROM "table" `
	tableLabel := ""

	if err := uuid.Validate(req.TableSlug); err != nil {
		query += fmt.Sprintf(`WHERE slug = '%s'`, req.TableSlug)
	} else {
		query += fmt.Sprintf(`WHERE id = '%s'`, req.TableSlug)
	}

	err = conn.QueryRow(ctx, query).Scan(&tableLabel)
	if err != nil && !strings.Contains(err.Error(), "no rows") {
		return err
	}

	if tableLabel != "" {
		req.TableSlug = tableLabel
	}

	if req.Type == "" {
		req.Type = "GLOBAL"
	}

	_, err = conn.Exec(ctx, versionH,
		req.ActionSource,
		req.ActionType,
		[]byte(req.Previus),
		[]byte(req.Current),
		req.Date,
		req.UserInfo,
		[]byte(req.Request),
		[]byte(req.Response),
		req.ApiKey,
		req.Type,
		req.TableSlug,
	)
	if err != nil {
		return err
	}

	return nil
}
