package models

import (
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/jackc/pgx/v5"
)

type TableVerReq struct {
	Tx   pgx.Tx
	Id   string
	Slug string
	Conn *psqlpool.Pool
}

type GetTableByIdSlugReq struct {
	Conn *psqlpool.Pool
	Id   string
	Slug string
}
