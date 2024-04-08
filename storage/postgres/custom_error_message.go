package postgres

import (
	"context"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type customErrorMessageRepo struct {
	db *pgxpool.Pool
}

func NewCustomErrorMessageRepo(db *pgxpool.Pool) customErrorMessageRepo {
	return customErrorMessageRepo{
		db: db,
	}
}

func (c customErrorMessageRepo) Create(ctx context.Context, req *nb.CreateCustomErrorMessage) (resp *nb.CustomErrorMessage, err error) {
	resp = &nb.CustomErrorMessage{}

	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	cus_id := uuid.NewString()

	query := `INSERT INTO "custom_error_message" (
		"id",
		"table_id",
		"message",
		"error_id",
		"code", 
		"language_id",
		"action_type"
	) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = conn.Exec(ctx, query,
		cus_id,
		req.TableId,
		req.Message,
		req.ErrorId,
		req.Code,
		req.LanguageId,
		req.ActionType,
	)
	if err != nil {
		return &nb.CustomErrorMessage{}, err
	}

	updateTable :=
		`UPDATE "table" SET 
			"is_changed" = true, 
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err = conn.Exec(ctx, updateTable, req.TableId)
	if err != nil {
		return nil, err
	}

	return c.GetSingle(ctx, &nb.CustomErrorMessagePK{Id: cus_id, ProjectId: req.ProjectId})
}

func (c customErrorMessageRepo) GetSingle(ctx context.Context, req *nb.CustomErrorMessagePK) (resp *nb.CustomErrorMessage, err error) {
	resp = &nb.CustomErrorMessage{}
	conn := psqlpool.Get(req.ProjectId)

	defer conn.Close()
	query := `SELECT 
				"id",
				"table_id",
				"message",
				"error_id",
				"code", 
				"language_id",
				"action_type"
				FROM "custom_error_message" WHERE id = $1`

	err = conn.QueryRow(ctx, query, req.Id).Scan(
		&resp.Id,
		&resp.TableId,
		&resp.Message,
		&resp.ErrorId,
		&resp.Code,
		&resp.LanguageId,
		&resp.ActionType,
	)
	if err != nil {
		return &nb.CustomErrorMessage{}, err
	}

	return resp, nil
}
func (c customErrorMessageRepo) GetList(ctx context.Context, req *nb.GetCustomErrorMessageListRequest) (resp *nb.GetCustomErrorMessageListResponse, err error) {
	resp = &nb.GetCustomErrorMessageListResponse{}
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `SELECT 
				COUNT(*) OVER(),
				"id",
				"table_id",
				"message",
				"error_id",
				"code", 
				"language_id",
				"action_type"
				FROM "custom_error_message" WHERE "table_id" = $1
				ORDER BY "created_at" DESC`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		row := &nb.CustomErrorMessage{}

		err = rows.Scan(
			&resp.Count,
			&row.Id,
			&row.TableId,
			&row.Message,
			&row.ErrorId,
			&row.Code,
			&row.LanguageId,
			&row.ActionType,
		)
		if err != nil {
			return resp, err
		}

		resp.CustomErrorMessages = append(resp.CustomErrorMessages, row)
	}

	return resp, nil
}

func (c customErrorMessageRepo) Update(ctx context.Context, req *nb.CustomErrorMessage) error {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `UPDATE "custom_error_message" SET
				"table_id" = $2,
				"message" = $3,
				"error_id" = $4,
				"code" = $5 , 
				"language_id" =  $6,
				"action_type" = $7,
				"updated_at" = CURRENT_TIMESTAMP
			WHERE id = $1
	`

	_, err := conn.Exec(ctx, query,
		req.Id,
		req.TableId,
		req.Message,
		req.ErrorId,
		req.Code,
		req.LanguageId,
		req.ActionType,
	)
	if err != nil {
		return err
	}

	updateTable :=
		`UPDATE "table" SET 
				"is_changed" = true, 
				updated_at = CURRENT_TIMESTAMP 
		WHERE id = $1`
	_, err = conn.Exec(ctx, updateTable, req.TableId)
	if err != nil {
		return err
	}

	return nil
}

func (c customErrorMessageRepo) Delete(ctx context.Context, req *nb.CustomErrorMessagePK) error {
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `DELETE FROM "custom_error_message" WHERE id = $1`

	_, err := conn.Exec(ctx, query, req.Id)
	if err != nil {
		return err
	}

	return nil
}

func (c customErrorMessageRepo) GetListForObject(ctx context.Context, req *nb.GetListForObjectRequest) (resp *nb.GetCustomErrorMessageListResponse, err error) {
	resp = &nb.GetCustomErrorMessageListResponse{}
	conn := psqlpool.Get(req.ProjectId)
	defer conn.Close()

	query := `SELECT 
				COUNT(*) OVER(),
				"id",
				"table_id",
				"message",
				"error_id",
				"code", 
				"language_id",
				"action_type"
				FROM "custom_error_message"  WHERE "table_id" = $1
				ORDER BY "created_at" DESC`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		row := &nb.CustomErrorMessage{}

		err = rows.Scan(
			&resp.Count,
			&row.Id,
			&row.TableId,
			&row.Message,
			&row.ErrorId,
			&row.Code,
			&row.LanguageId,
			&row.ActionType,
		)
		if err != nil {
			return resp, err
		}

		resp.CustomErrorMessages = append(resp.CustomErrorMessages, row)
	}

	return resp, nil
}
