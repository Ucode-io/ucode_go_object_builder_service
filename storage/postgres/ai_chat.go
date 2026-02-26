package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"ucode/ucode_go_object_builder_service/config"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/structpb"
)

type aiChatRepo struct {
	db *psqlpool.Pool
}

func NewAiChatRepo(db *psqlpool.Pool) storage.AiChatRepoI {
	return &aiChatRepo{
		db: db,
	}
}

// ==================== Chats ====================

func (r *aiChatRepo) CreateChat(ctx context.Context, req *nb.CreateChatRequest) (*nb.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.CreateChat")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		id  = uuid.NewString()
		now = time.Now()

		chat                 nb.Chat
		description          sql.NullString
		createdAt, updatedAt time.Time

		model = req.GetModel()
	)

	if model == "" {
		model = "claude-sonnet-4-5"
	}

	var query = `
		INSERT INTO chats (id, project_id, title, description, model, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, project_id, title, description, model, total_tokens, created_at, updated_at
	`

	err = conn.QueryRow(ctx, query, id, req.GetProjectId(),
		req.GetTitle(), nullString(req.GetDescription()),
		model, now,
	).Scan(
		&chat.Id, &chat.ProjectId, &chat.Title, &description,
		&chat.Model, &chat.TotalTokens, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	if description.Valid {
		chat.Description = description.String
	}

	chat.CreatedAt = createdAt.Format(time.RFC3339)
	chat.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &chat, nil
}

func (r *aiChatRepo) GetChatById(ctx context.Context, req *nb.ChatPrimaryKey) (*nb.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetChatById")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		chat                 nb.Chat
		description          sql.NullString
		createdAt, updatedAt time.Time

		query = `
			SELECT id, project_id, title, description, model, total_tokens, created_at, updated_at
			FROM chats
			WHERE id = $1
		`
	)

	err = conn.QueryRow(ctx, query, req.GetId()).Scan(
		&chat.Id, &chat.ProjectId, &chat.Title, &description,
		&chat.Model, &chat.TotalTokens, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat: %w", err)
	}

	if description.Valid {
		chat.Description = description.String
	}

	chat.CreatedAt = createdAt.Format(time.RFC3339)
	chat.UpdatedAt = updatedAt.Format(time.RFC3339)

	if req.GetWithMessages() {
		messages, err := r.GetMessages(ctx, &nb.GetMessagesRequest{
			ResourceEnvId: req.GetResourceEnvId(),
			ChatId:        chat.Id,
		})
		if err != nil {
			return nil, err
		}
		chat.Messages = messages.GetMessages()
	}

	return &chat, nil
}

func (r *aiChatRepo) GetChatByProjectId(ctx context.Context, req *nb.ChatByProjectIdRequest) (*nb.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetChatByProjectId")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		chat                 nb.Chat
		description          sql.NullString
		createdAt, updatedAt time.Time

		query = `
			SELECT id, project_id, title, description, model, total_tokens, created_at, updated_at
			FROM chats
			WHERE project_id = $1
		`
	)

	err = conn.QueryRow(ctx, query, req.GetProjectId()).Scan(
		&chat.Id, &chat.ProjectId, &chat.Title, &description,
		&chat.Model, &chat.TotalTokens, &createdAt, &updatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get chat by project_id: %w", err)
	}

	if description.Valid {
		chat.Description = description.String
	}

	chat.CreatedAt = createdAt.Format(time.RFC3339)
	chat.UpdatedAt = updatedAt.Format(time.RFC3339)

	if req.GetWithMessages() {
		messages, err := r.GetMessages(ctx, &nb.GetMessagesRequest{
			ResourceEnvId: req.GetResourceEnvId(),
			ChatId:        chat.Id,
		})
		if err != nil {
			return nil, err
		}
		chat.Messages = messages.GetMessages()
	}

	return &chat, nil
}

func (r *aiChatRepo) GetAllChats(ctx context.Context, req *nb.GetAllChatsRequest) (*nb.GetAllChatsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetAllChats")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		queryBuilder strings.Builder
		countBuilder strings.Builder
		args         = make([]any, 0)
		chats        = make([]*nb.Chat, 0)

		count       int32
		orderDir    = "DESC"
		orderColumn = "c.created_at"
	)

	queryBuilder.WriteString(`
		SELECT c.id, c.project_id, c.title, c.description, c.model, c.total_tokens, c.created_at, c.updated_at
		FROM chats c
		WHERE 1=1
	`)
	countBuilder.WriteString(`SELECT COUNT(*) FROM chats c WHERE 1=1`)

	if req.GetTitle() != "" {
		args = append(args, "%"+req.GetTitle()+"%")
		queryBuilder.WriteString(fmt.Sprintf(" AND c.title ILIKE $%d", len(args)))
		countBuilder.WriteString(fmt.Sprintf(" AND c.title ILIKE $%d", len(args)))
	}

	if req.GetModel() != "" {
		args = append(args, req.GetModel())
		queryBuilder.WriteString(fmt.Sprintf(" AND c.model = $%d", len(args)))
		countBuilder.WriteString(fmt.Sprintf(" AND c.model = $%d", len(args)))
	}

	if req.GetProjectId() != "" {
		args = append(args, req.GetProjectId())
		queryBuilder.WriteString(fmt.Sprintf(" AND c.project_id = $%d", len(args)))
		countBuilder.WriteString(fmt.Sprintf(" AND c.project_id = $%d", len(args)))
	}

	err = conn.QueryRow(ctx, countBuilder.String(), args...).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count chats: %w", err)
	}

	if col, ok := config.ChatAllowedOrder[req.GetOrderBy()]; ok {
		orderColumn = col
	}

	if req.GetOrderDirection() == "asc" {
		orderDir = "ASC"
	}

	queryBuilder.WriteString(fmt.Sprintf(" ORDER BY %s %s", orderColumn, orderDir))

	if req.GetLimit() > 0 {
		args = append(args, req.GetLimit())
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	}

	if req.GetOffset() > 0 {
		args = append(args, req.GetOffset())
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	rows, err := conn.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			chat                 nb.Chat
			description          sql.NullString
			createdAt, updatedAt time.Time
		)

		err = rows.Scan(
			&chat.Id, &chat.ProjectId, &chat.Title, &description,
			&chat.Model, &chat.TotalTokens, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chat: %w", err)
		}

		if description.Valid {
			chat.Description = description.String
		}

		chat.CreatedAt = createdAt.Format(time.RFC3339)
		chat.UpdatedAt = updatedAt.Format(time.RFC3339)

		chats = append(chats, &chat)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &nb.GetAllChatsResponse{
		Chats: chats,
		Count: count,
	}, nil
}

func (r *aiChatRepo) UpdateChat(ctx context.Context, req *nb.UpdateChatRequest) (*nb.Chat, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.UpdateChat")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		setClauses = []string{"updated_at = NOW()"}
		args       []any
		argIndex   = 1
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

	if req.GetModel() != "" {
		setClauses = append(setClauses, fmt.Sprintf("model = $%d", argIndex))
		args = append(args, req.GetModel())
		argIndex++
	}

	if req.GetTotalTokens() > 0 {
		setClauses = append(setClauses, fmt.Sprintf("total_tokens = $%d", argIndex))
		args = append(args, req.GetTotalTokens())
		argIndex++
	}

	args = append(args, req.GetId())

	var (
		chat                 nb.Chat
		description          sql.NullString
		createdAt, updatedAt time.Time

		query = fmt.Sprintf(`
			UPDATE chats SET %s
			WHERE id = $%d
				RETURNING id, project_id, title, description,
			    model, total_tokens, created_at, updated_at
		`,
			strings.Join(setClauses, ", "), argIndex)
	)

	err = conn.QueryRow(ctx, query, args...).Scan(
		&chat.Id, &chat.ProjectId, &chat.Title, &description,
		&chat.Model, &chat.TotalTokens, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update chat: %w", err)
	}

	if description.Valid {
		chat.Description = description.String
	}

	chat.CreatedAt = createdAt.Format(time.RFC3339)
	chat.UpdatedAt = updatedAt.Format(time.RFC3339)

	return &chat, nil
}

func (r *aiChatRepo) DeleteChat(ctx context.Context, req *nb.ChatPrimaryKey) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.DeleteChat")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return err
	}

	var query = `DELETE FROM chats WHERE id = $1`
	res, err := conn.Exec(ctx, query, req.GetId())
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("chat not found: %s", req.GetId())
	}

	return nil
}

// ==================== Messages ====================

func (r *aiChatRepo) CreateMessage(ctx context.Context, req *nb.CreateMessageRequest) (*nb.Message, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.CreateMessage")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		id  = uuid.NewString()
		msg nb.Message

		tokensUsed sql.NullInt32
		createdAt  time.Time

		tokensArg any

		images = req.GetImages()
	)

	if images == nil {
		images = []string{}
	}

	var query = `
		INSERT INTO messages (id, chat_id, role, content, images, has_files, tokens_used)
		VALUES ($1, $2, $3::message_role, $4, $5, $6, $7)
		RETURNING id, chat_id, role, content, images, has_files, tokens_used, created_at
	`

	if req.GetTokensUsed() > 0 {
		tokensArg = req.GetTokensUsed()
	}

	err = conn.QueryRow(ctx, query, id,
		req.GetChatId(), req.GetRole(), req.GetContent(),
		images, req.GetHasFiles(), tokensArg,
	).Scan(
		&msg.Id, &msg.ChatId, &msg.Role, &msg.Content,
		&msg.Images, &msg.HasFiles, &tokensUsed, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	if tokensUsed.Valid {
		msg.TokensUsed = tokensUsed.Int32
	}

	msg.CreatedAt = createdAt.Format(time.RFC3339)

	if req.GetTokensUsed() > 0 {
		var updateQuery = `UPDATE chats SET total_tokens = total_tokens + $1, updated_at = NOW() WHERE id = $2`
		_, err = conn.Exec(ctx, updateQuery, req.GetTokensUsed(), req.GetChatId())
		if err != nil {
			return nil, fmt.Errorf("failed to update chat total_tokens: %w", err)
		}
	}

	return &msg, nil
}

func (r *aiChatRepo) GetMessages(ctx context.Context, req *nb.GetMessagesRequest) (*nb.GetMessagesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetMessages")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		queryBuilder strings.Builder
		args         = []any{req.GetChatId()}
		messages     = make([]*nb.Message, 0)

		count int32
	)

	queryBuilder.WriteString(`
		SELECT id, chat_id, role, content, images, has_files, tokens_used, created_at
		FROM messages
		WHERE chat_id = $1
	`)

	var countQuery = `SELECT COUNT(*) FROM messages WHERE chat_id = $1`

	err = conn.QueryRow(ctx, countQuery, req.GetChatId()).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	queryBuilder.WriteString(" ORDER BY created_at ASC")

	if req.GetLimit() > 0 {
		args = append(args, req.GetLimit())
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	}

	if req.GetOffset() > 0 {
		args = append(args, req.GetOffset())
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	rows, err := conn.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			msg        nb.Message
			tokensUsed sql.NullInt32
			createdAt  time.Time
		)

		err = rows.Scan(
			&msg.Id, &msg.ChatId, &msg.Role, &msg.Content,
			&msg.Images, &msg.HasFiles, &tokensUsed, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if tokensUsed.Valid {
			msg.TokensUsed = tokensUsed.Int32
		}
		msg.CreatedAt = createdAt.Format(time.RFC3339)

		messages = append(messages, &msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &nb.GetMessagesResponse{
		Messages: messages,
		Count:    count,
	}, nil
}

func (r *aiChatRepo) DeleteMessage(ctx context.Context, req *nb.MessagePrimaryKey) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.DeleteMessage")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return err
	}

	var query = `DELETE FROM messages WHERE id = $1`
	res, err := conn.Exec(ctx, query, req.GetId())
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("message not found: %s", req.GetId())
	}

	return nil
}

// ==================== File Versions ====================

func (r *aiChatRepo) CreateFileVersion(ctx context.Context, req *nb.CreateFileVersionRequest) (*nb.FileVersion, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.CreateFileVersion")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		id = uuid.NewString()
		fv nb.FileVersion

		changeSummary sql.NullString
		createdAt     time.Time
	)

	var fileGraphBytes []byte
	if req.GetFileGraph() != nil {
		fileGraphBytes, err = json.Marshal(req.GetFileGraph().AsMap())
		if err != nil {
			fileGraphBytes = []byte("{}")
		}
	} else {
		fileGraphBytes = []byte("{}")
	}

	query := `
		INSERT INTO file_versions (id, file_id, message_id, version_num, content, file_graph, change_summary)
		VALUES ($1, $2, $3, COALESCE((SELECT MAX(version_num) FROM file_versions WHERE file_id = $2), 0) + 1, $4, $5, $6)
		RETURNING id, file_id, message_id, version_num, content, file_graph, change_summary, created_at
	`

	var retFileGraph []byte

	err = conn.QueryRow(ctx, query,
		id, req.GetFileId(), req.GetMessageId(), req.GetContent(),
		fileGraphBytes, nullString(req.GetChangeSummary()),
	).Scan(
		&fv.Id, &fv.FileId, &fv.MessageId, &fv.VersionNum,
		&fv.Content, &retFileGraph, &changeSummary, &createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create file_version: %w", err)
	}

	if len(retFileGraph) > 0 {
		var raw map[string]any
		if err = json.Unmarshal(retFileGraph, &raw); err == nil {
			fv.FileGraph, _ = structpb.NewStruct(raw)
		}
	}

	if changeSummary.Valid {
		fv.ChangeSummary = changeSummary.String
	}

	fv.CreatedAt = createdAt.Format(time.RFC3339)

	return &fv, nil
}

func (r *aiChatRepo) GetFileVersions(ctx context.Context, req *nb.GetFileVersionsRequest) (*nb.GetFileVersionsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetFileVersions")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		queryBuilder strings.Builder
		args         = []any{req.GetFileId()}
		versions     = make([]*nb.FileVersion, 0)

		count int32
	)

	queryBuilder.WriteString(`
		SELECT id, file_id, message_id, version_num, content, file_graph, change_summary, created_at
		FROM file_versions
		WHERE file_id = $1
	`)

	var countQuery = `SELECT COUNT(*) FROM file_versions WHERE file_id = $1`
	err = conn.QueryRow(ctx, countQuery, req.GetFileId()).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed to count file_versions: %w", err)
	}

	queryBuilder.WriteString(" ORDER BY version_num DESC")

	if req.GetLimit() > 0 {
		args = append(args, req.GetLimit())
		queryBuilder.WriteString(fmt.Sprintf(" LIMIT $%d", len(args)))
	}

	if req.GetOffset() > 0 {
		args = append(args, req.GetOffset())
		queryBuilder.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))
	}

	rows, err := conn.Query(ctx, queryBuilder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query file_versions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			fv             nb.FileVersion
			fileGraphBytes []byte
			changeSummary  sql.NullString
			createdAt      time.Time
		)

		err = rows.Scan(
			&fv.Id, &fv.FileId, &fv.MessageId, &fv.VersionNum,
			&fv.Content, &fileGraphBytes, &changeSummary, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file_version: %w", err)
		}

		if len(fileGraphBytes) > 0 {
			var raw map[string]any
			if err = json.Unmarshal(fileGraphBytes, &raw); err == nil {
				fv.FileGraph, _ = structpb.NewStruct(raw)
			}
		}
		if changeSummary.Valid {
			fv.ChangeSummary = changeSummary.String
		}
		fv.CreatedAt = createdAt.Format(time.RFC3339)

		versions = append(versions, &fv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &nb.GetFileVersionsResponse{
		FileVersions: versions,
		Count:        count,
	}, nil
}

func (r *aiChatRepo) GetFileVersionsByMessage(ctx context.Context, req *nb.GetFileVersionsByMessageRequest) (*nb.GetFileVersionsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "aiChatRepo.GetFileVersionsByMessage")
	defer span.Finish()

	conn, err := psqlpool.Get(req.GetResourceEnvId())
	if err != nil {
		return nil, err
	}

	var (
		versions = make([]*nb.FileVersion, 0)

		query = `
			SELECT id, file_id, message_id, version_num, content, file_graph, change_summary, created_at
			FROM file_versions
			WHERE message_id = $1
			ORDER BY created_at ASC
	`
	)

	rows, err := conn.Query(ctx, query, req.GetMessageId())
	if err != nil {
		return nil, fmt.Errorf("failed to query file_versions by message: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			fv             nb.FileVersion
			fileGraphBytes []byte
			changeSummary  sql.NullString
			createdAt      time.Time
		)

		err = rows.Scan(
			&fv.Id, &fv.FileId, &fv.MessageId, &fv.VersionNum,
			&fv.Content, &fileGraphBytes, &changeSummary, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file_version: %w", err)
		}

		if len(fileGraphBytes) > 0 {
			var raw map[string]any
			if err = json.Unmarshal(fileGraphBytes, &raw); err == nil {
				fv.FileGraph, _ = structpb.NewStruct(raw)
			}
		}
		if changeSummary.Valid {
			fv.ChangeSummary = changeSummary.String
		}

		fv.CreatedAt = createdAt.Format(time.RFC3339)

		versions = append(versions, &fv)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &nb.GetFileVersionsResponse{
		FileVersions: versions,
		Count:        int32(len(versions)),
	}, nil
}
