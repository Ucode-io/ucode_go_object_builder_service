package models

import (
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/jackc/pgx/v5"
)

type TableVerReq struct {
	Tx   pgx.Tx `json:"-"`
	Id   string
	Slug string
	Conn *psqlpool.Pool
}

type GetTableByIdSlugReq struct {
	Conn *psqlpool.Pool
	Id   string
	Slug string
}

type TableSchema struct {
	Name      string         `json:"name"`
	Columns   []ColumnInfo   `json:"columns"`
	Relations []RelationInfo `json:"relations"`
}

type ColumnInfo struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
}

type RelationInfo struct {
	TableFrom string `json:"table_from"`
	TableTo   string `json:"table_to"`
}
