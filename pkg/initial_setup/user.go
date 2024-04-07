package initialsetup

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(conn *pgxpool.Pool, userId, roleId, clientTypeId, clientPlatformId, projectId string) error {
	query := `INSERT INTO "user" ("guid", "role_id", "client_type_id", "client_platform_id", "project_id", "active") 
	VALUES 
	($1, $2, $3, $4, $5, 1);`

	_, err := conn.Exec(context.Background(), query, userId, roleId, clientTypeId, clientPlatformId, projectId)
	if err != nil {
		return err
	}

	return nil
}
