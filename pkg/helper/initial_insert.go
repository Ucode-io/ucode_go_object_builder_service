package helper

import (
	"fmt"
	initialsetup "ucode/ucode_go_object_builder_service/pkg/initial_setup"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertDatas(conn *pgxpool.Pool, userId, projectId, clientTypeId, roleId string) error {
	if clientTypeId == "" {
		clientTypeId = uuid.NewString()
	}
	if roleId == "" {
		roleId = uuid.NewString()
	}
	clientPlatformId := uuid.NewString()

	err := initialsetup.CreateDefaultClientPlatform(conn, clientPlatformId, clientTypeId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientPlatform - %v", err)
	}

	err = initialsetup.CreateDefaultClientPlatform(conn, clientPlatformId, clientTypeId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientPlatform - %v", err)
	}

	err = initialsetup.CreateDefaultFieldPermission(conn, roleId)
	if err != nil {
		return fmt.Errorf("CreateDefaultFieldPermission - %v", err)
	}

	return nil
}
