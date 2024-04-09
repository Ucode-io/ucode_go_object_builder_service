package initialsetup

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateDefaultGlobalPermission(conn *pgxpool.Pool, roleId string) error {
	query := `INSERT INTO global_permission (guid, role_id, chat, menu_button, settings_button, projects_button, environments_button, api_keys_button, menu_setting_button, redirects_button, profile_settings_button, project_settings_button, project_button, sms_button, version_button) VALUES
	($1, $2, true, true, true, true, true, true, true, true, true, true, true, true, true, true)
	`

	_, err := conn.Exec(context.Background(), query, uuid.NewString(), roleId)
	if err != nil {
		return err
	}
	return nil
}
