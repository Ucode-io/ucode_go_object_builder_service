package postgres

import (
	"context"
	"fmt"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	psqlpool "ucode/ucode_go_object_builder_service/pkg/pool"
	"ucode/ucode_go_object_builder_service/storage"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/golang-migrate/migrate/v4"
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

	fmt.Println("DB URL ", dbUrl)
	// Create init tables ru migration
	migrations, err := migrate.New(
		"../../migrations/postgres", // path to migrations
		dbUrl,                       // database URL
	)
	if err != nil {
		fmt.Println("error create migrations")
		return err
	}
	defer migrations.Close()

	// Run migration UP
	err = migrations.Up()
	if err != nil && err != migrate.ErrNoChange {
		fmt.Println("error run migrations")
		return err
	}

	// cmd := exec.Command("migrate", "-path", "../../migrations/postgres", "-database", dbUrl, "up")

	// // Run the command asynchronously
	// err = cmd.Start()
	// if err != nil {
	// 	fmt.Println("Error executing command:", err)
	// 	return err
	// }

	// // Wait for the command to complete
	// err = cmd.Wait()
	// if err != nil {
	// 	fmt.Println("Command finished with error:", err)
	// 	return err
	// }

	// After the command completes successfully, continue with other operations
	fmt.Println("Migration completed successfully")

	helper.InsertDatas(pool, req.UserId, req.ProjectId, req.ClientTypeId, req.RoleId)

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
