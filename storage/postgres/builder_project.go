package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type builderProjectRepo struct {
	db *pgxpool.Pool
}

func NewBuilderProjectRepo(db *pgxpool.Pool) storage.BuilderProjectRepoI {
	return &builderProjectRepo{
		db: db,
	}
}

func (b *builderProjectRepo) Register(ctx context.Context, req *nb.RegisterProjectRequest) error {

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
		fmt.Println("error parse config")
		return err
	}

	// config.MaxConns = cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		fmt.Println("error create pool")
		return err
	}

	err = pool.Ping(ctx)
	if err != nil {
		fmt.Println("error ping")
		return err
	}

	// fmt.Println("DB URL ", dbUrl)
	// Create init tables ru migration
	// migrations, err := migrate.New(
	// 	"../../migrations/postgres", // path to migrations
	// 	dbUrl,                       // database URL
	// )
	// if err != nil {
	// 	fmt.Println("error create migrations")
	// 	return err
	// }
	// defer migrations.Close()

	// // Run migration UP
	// err = migrations.Up()
	// if err != nil && err != migrate.ErrNoChange {
	// 	fmt.Println("error run migrations")
	// 	return err
	// }

	dbInstance, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return err
	}

	db, err := postgres.WithInstance(dbInstance, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	migrationsDir := "file://migrations/postgres"
	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		db,
	)
	if err != nil {
		log.Fatalf("Failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("An error occurred while syncing the database: %v", err)
	}

	fmt.Println("Migration completed successfully")

	err = helper.InsertDatas(pool, req.UserId, req.ProjectId, req.ClientTypeId, req.RoleId)
	if err != nil {
		return err
	}

	psqlpool.Add(req.ProjectId, pool)

	return nil
}

func (b *builderProjectRepo) RegisterProjects(ctx context.Context, req *nb.RegisterProjectRequest) error {
	return nil
}

func (b *builderProjectRepo) Deregister(ctx context.Context, req *nb.DeregisterProjectRequest) error {
	return nil
}

func (b *builderProjectRepo) Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) error {
	return nil
}

func (b *builderProjectRepo) RegisterMany(ctx context.Context, req *nb.RegisterManyProjectsRequest) (resp *nb.RegisterManyProjectsResponse, err error) {
	return nil, nil
}

func (b *builderProjectRepo) DeregisterMany(ctx context.Context, req *nb.DeregisterManyProjectsRequest) (resp *nb.DeregisterManyProjectsResponse, err error) {
	return nil, nil
}
