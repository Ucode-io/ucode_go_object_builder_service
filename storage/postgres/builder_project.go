package postgres

import (
	"context"
	"fmt"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func Register(ctx context.Context, req *nb.RegisterProjectRequest) error {

	if req.UserId == "" {
		return fmt.Errorf("error user_id is required")
	}

	if req.ProjectId == "" {
		return fmt.Errorf("error project_id is required")
	}

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		req.Credentials.Username,
		req.Credentials.Password,
		req.Credentials.Host,
		req.Credentials.Port,
		req.Credentials.Database,
	)

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return err
	}

	// config.MaxConns = cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return err
	}

	// Create init tables ru migration
	m, err := migrate.New(
		"./migrations", // path to migrations
		dbUrl,          // database URL
	)
	if err != nil {
		return err
	}

	defer m.Close()

	// Run migration UP
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	psqlpool.Add(req.ProjectId, pool)

	return nil
}

func RegisterProjects(ctx context.Context, req *nb.RegisterProjectRequest) error

func Deregister(ctx context.Context, req *nb.DeregisterProjectRequest) error

func Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) error

func RegisterMany(ctx context.Context, req *nb.RegisterManyProjectsRequest) (resp *nb.RegisterManyProjectsResponse, err error)

func DeregisterMany(ctx context.Context, req *nb.DeregisterManyProjectsRequest) (resp *nb.DeregisterManyProjectsResponse, err error)
