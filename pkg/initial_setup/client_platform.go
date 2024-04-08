package initialsetup

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDefaultClientPlatform(conn *pgxpool.Pool, clientPlatformId, clientTypeId, projectId string) error {
	query := `INSERT INTO client_platform (guid, project_id, name, subdomain, client_type_ids) 
        VALUES ($1, $2, 'Ucode', 'ucode', $3)`

	_, err := conn.Exec(context.Background(), query, clientPlatformId, projectId, []string{clientTypeId})
	if err != nil {
		return err
	}
	
	return nil
}
