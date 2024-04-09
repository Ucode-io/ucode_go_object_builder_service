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
	if projectId == "" {
		projectId = uuid.NewString()
	}
	if userId == "" {
		userId = uuid.NewString()
	}
	clientPlatformId := uuid.NewString()
	testLoginId := uuid.NewString()

	err := initialsetup.CreateDefaultClientPlatform(conn, clientPlatformId, clientTypeId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientPlatform - %v", err)
	}

	err = initialsetup.CreateDefaultClientType(conn, clientPlatformId, clientTypeId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientType - %v", err)
	}

	err = initialsetup.CreateDefaultRole(conn, roleId, clientPlatformId, clientTypeId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientType - %v", err)
	}

	err = initialsetup.CreateDefaultFieldPermission(conn, roleId)
	if err != nil {
		return fmt.Errorf("CreateDefaultFieldPermission - %v", err)
	}

	err = initialsetup.CreateDefaultGlobalPermission(conn, roleId)
	if err != nil {
		return fmt.Errorf("CreateDefaultClientType - %v", err)
	}

	err = initialsetup.CreateDefaultRecordPermission(conn, roleId)
	if err != nil {
		return fmt.Errorf("CreateDefaultRecordPermission - %v", err)
	}

	err = initialsetup.CreateDefaultTestLogin(conn, testLoginId, clientTypeId)
	if err != nil {
		return fmt.Errorf("CreateDefaultTestLogin - %v", err)
	}

	err = initialsetup.CreateDefaultUser(conn, userId, roleId, clientTypeId, clientPlatformId, projectId)
	if err != nil {
		return fmt.Errorf("CreateDefaultUser - %v", err)
	}

	err = initialsetup.CreateDefaultViewRelationPermission(conn, roleId)
	if err != nil {
		return fmt.Errorf("CreateDefaultViewRelationPermission - %v", err)
	}

	return nil
}
