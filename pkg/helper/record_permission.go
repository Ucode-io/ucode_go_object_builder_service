package helper

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GetRecordPermissionRequest struct {
	Conn      *pgxpool.Pool
	TableSlug string
	RoleId    string
}

type GetRecordPermissionResponse struct {
	Guid            string
	RoleId          string
	TableSlug       string
	Read            string
	Write           string
	Update          string
	Delete          string
	IsPublic        bool
	IsHaveCondition bool
}

func GetRecordPermission(ctx context.Context, req GetRecordPermissionRequest) (*GetRecordPermissionResponse, error) {
	recordPermission := GetRecordPermissionResponse{}

	query := `
		SELECT
			"guid",
			"role_id",
			"table_slug",
			"read",
			"write",
			"update",
			"delete",
			"is_public",
			"is_have_condition"
		FROM "record_permission"
		WHERE table_slug = $1 AND role_id = $2
	`

	err := req.Conn.QueryRow(ctx, query, req.TableSlug, req.RoleId).Scan(
		&recordPermission.Guid,
		&recordPermission.RoleId,
		&recordPermission.TableSlug,
		&recordPermission.Read,
		&recordPermission.Write,
		&recordPermission.Update,
		&recordPermission.Delete,
		&recordPermission.IsPublic,
		&recordPermission.IsHaveCondition,
	)
	if err != nil {
		return &GetRecordPermissionResponse{}, err
	}

	return &recordPermission, nil
}
