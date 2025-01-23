package models

import (
	pa "ucode/ucode_go_object_builder_service/genproto/auth_service"

	"github.com/jackc/pgx/v5"
)

type CreateSyncWithLoginTableRequest struct {
	Tx                 pgx.Tx
	Guid               string
	UserIdAuth         string
	LoginTableSlug     string
	TableAttributesMap map[string]any
	Data               map[string]any
}

type UpdateSyncWithLoginTableRequest struct {
	Tx        pgx.Tx
	EnvId     string
	ProjectId string
	Guid      string
	Data      map[string]any
}

type DeleteSyncWithLoginTableRequest struct {
	Tx       pgx.Tx
	Id       string
	Response map[string]any
}

type DeleteManySyncWithLoginTableRequest struct {
	Tx    pgx.Tx
	Ids   []string
	Users []*pa.DeleteManyUserRequest_User
	Table *Table
	Data  map[string]any
}
