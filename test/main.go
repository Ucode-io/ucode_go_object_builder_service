package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
)

type ClientType struct {
	Guid              string   `json:"guid"`
	ProjectId         string   `json:"project_id"`
	Name              string   `json:"name"`
	SelfRegister      bool     `json:"self_register"`
	SelfRecover       bool     `json:"self_recover"`
	ClientPlatformIds []string `json:"client_platform_ids"`
	ConfirmBy         string   `json:"confirm_by"`
	IsSystem          bool     `json:"is_system"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
	TableSlug         string   `json:"table_slug"`
	DefaultPage       string   `json:"default_page"`
}

func main() {
	pool, err := pgxpool.ParseConfig("postgres://login_psql_5e9c087aca884920be1936cb20ca56f9_p_postgres_svcs:oka@65.109.239.69:5432/login_psql_5e9c087aca884920be1936cb20ca56f9_p_postgres_svcs?sslmode=disable")
	if err != nil {
		fmt.Println(err)
		return
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), pool)
	if err != nil {
		fmt.Println(err)
		return
	}

	query := `
		SELECT
			"guid",
			"project_id",
			"name",
			"self_register",
			"self_recover",
			"client_platform_ids",
			"confirm_by",
			"is_system",
			"table_slug",
			"default_page"
		FROM client_type WHERE "guid" = $1 OR "name" = $1::varchar
	`

	var (
		clientType = ClientType{}
		// clientPlatformIds pq.StringArray
	)

	err = conn.QueryRow(context.Background(), query, "b099de17-05dd-4a86-85ed-d0b8076459c7").Scan(
		&clientType.Guid,
		&clientType.ProjectId,
		&clientType.Name,
		&clientType.SelfRegister,
		&clientType.SelfRecover,
		pq.Array(&clientType.ClientPlatformIds),
		&clientType.ConfirmBy,
		&clientType.IsSystem,
		&clientType.TableSlug,
		&clientType.DefaultPage,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

}
