package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/company_service"
	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/helper"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/types/known/emptypb"
)

type builderProjectService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	nb.UnimplementedBuilderProjectServiceServer
}

func NewBuilderProjectService(strg storage.StorageI, cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) *builderProjectService { // strg storage.StorageI,
	return &builderProjectService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

// runMigrationsForDB opens a temporary sql.DB connection, runs migrations (with dirty-state recovery),
// then closes the connection. This is intentionally separate from the long-lived pgxpool used at runtime.
func runMigrationsForDB(dbURL string) error {
	dbInstance, err := sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer dbInstance.Close()

	dbDriver, err := postgres.WithInstance(dbInstance, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations/postgres", "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate instance: %w", err)
	}

	version, dirty, vErr := m.Version()
	if vErr != nil && !errors.Is(vErr, migrate.ErrNilVersion) {
		return fmt.Errorf("migrate version: %w", vErr)
	}
	if dirty {
		// Pod was killed between dirty=true and dirty=false on the previous run.
		// Step back to the last clean version so Up() can re-apply it safely.
		if forceErr := m.Force(int(version) - 1); forceErr != nil {
			return fmt.Errorf("migrate force: %w", forceErr)
		}
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func (b *builderProjectService) Register(ctx context.Context, req *nb.RegisterProjectRequest) (resp *emptypb.Empty, err error) {
	b.log.Info("!!!RegisterProject--->", logger.Any("req", req))

	if req.UserId == "" {
		err = fmt.Errorf("error user_id is required")
		b.log.Error("!!!RegisterProjectError--->", logger.Error(err))
		return resp, err
	}

	if req.ProjectId == "" {
		err = fmt.Errorf("error project_id is required")
		b.log.Error("!!!RegisterProjectError--->", logger.Error(err))
		return resp, err
	}

	dbUrl := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		req.Credentials.Username,
		req.Credentials.Password,
		req.Credentials.Host,
		req.Credentials.Port,
		req.Credentials.Database,
	)

	if err = runMigrationsForDB(dbUrl); err != nil {
		b.log.Error("!!!RegisterProject->MigrateUp", logger.Error(err))
		return resp, err
	}

	b.log.Info("::::::::::::::::Migration completed successfully::::::::::::::::")

	poolCfg, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		b.log.Error("!!!RegisterProject->ParseResourceCredentials", logger.Error(err))
		return resp, err
	}

	poolCfg.MaxConns = b.cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		b.log.Error("!!!RegisterProject->CreatePool", logger.Error(err))
		return resp, err
	}

	if err = pool.Ping(ctx); err != nil {
		b.log.Error("!!!RegisterProject->Ping", logger.Error(err))
		return resp, err
	}

	resourceEnv, err := b.services.ResourceService().GetResourceEnvironment(ctx, &company_service.GetResourceEnvironmentReq{
		Username: req.Credentials.Username,
	})
	if err != nil {
		b.log.Error("!!!RegisterProject->GetResourceEnvironment", logger.Error(err))
		return resp, err
	}

	err = helper.InsertDatas(pool, req.UserId, req.ProjectId, req.ClientTypeId, req.RoleId, resourceEnv.Id)
	if err != nil {
		b.log.Error("!!!RegisterProject->InsertDatas", logger.Error(err))
		return resp, err
	}

	psqlpool.Add(resourceEnv.Id, &psqlpool.Pool{Db: pool})

	return resp, nil
}

func (b *builderProjectService) RegisterProjects(ctx context.Context, req *nb.RegisterProjectRequest) (resp *emptypb.Empty, err error) {
	return resp, nil
}

func (b *builderProjectService) Deregister(ctx context.Context, req *nb.DeregisterProjectRequest) (resp *emptypb.Empty, err error) {
	return resp, nil
}

// Reconnect creates (or replaces) the runtime pgxpool for a project database.
// It does NOT run migrations — call runMigrationsForDB separately before this.
func (b *builderProjectService) Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) (resp *emptypb.Empty, err error) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		req.Credentials.GetUsername(),
		req.Credentials.GetPassword(),
		req.Credentials.GetHost(),
		req.Credentials.GetPort(),
		req.Credentials.GetDatabase(),
	)

	b.log.Info("!!!Reconnect--->", logger.Any("projectId", req.ProjectId))

	poolCfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		b.log.Error("!!!Reconnect->ParseConfig", logger.Error(err))
		return resp, err
	}

	// Use a small per-tenant connection limit to avoid exhausting PostgreSQL's
	// max_connections when hundreds of project databases are loaded at startup.
	poolCfg.MaxConns = b.cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		b.log.Error("!!!Reconnect->CreatePool", logger.Error(err))
		return resp, err
	}

	if existing, getErr := psqlpool.Get(req.ProjectId); getErr == nil {
		existing.Db.Close()
		psqlpool.Remove(req.ProjectId)
	}
	psqlpool.Add(req.ProjectId, &psqlpool.Pool{Db: pool})

	b.log.Info("::::::::::::::::AUTOCONNECTRED AND SUCCESSFULLY ADDED TO POOL::::::::::::::::")

	return resp, nil
}

func (b *builderProjectService) RegisterMany(ctx context.Context, req *nb.RegisterManyProjectsRequest) (resp *nb.RegisterManyProjectsResponse, err error) {
	return resp, nil
}

func (b *builderProjectService) DeregisterMany(ctx context.Context, req *nb.DeregisterManyProjectsRequest) (resp *nb.DeregisterManyProjectsResponse, err error) {
	return resp, err
}

func (b *builderProjectService) AutoConnect(ctx context.Context) error {
	b.log.Info("!!!AUTOCONNECTING TO RESOURCES--->")

	if b.cfg.K8sNamespace == "" {
		err := errors.New("k8s_namespace is required to get project")
		b.log.Error("!!!AutoConnect--->", logger.Error(err))
		return err
	}

	connect, err := b.services.ResourceService().AutoConnect(
		ctx, &company_service.GetProjectsRequest{
			K8SNamespace: b.cfg.K8sNamespace,
			NodeType:     b.cfg.NodeType,
		},
	)
	if err != nil {
		b.log.Error("!!!AutoConnect--->ResourceAutoConnect", logger.Error(err))
		return err
	}

	b.log.Info("BUILDING PROJECTS---> ", logger.Any("COUNT", len(connect.Res)))

	for _, resource := range connect.Res {
		if resource.ResourceType != company_service.ResourceType_POSTGRESQL {
			continue
		}

		dbURL := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			resource.GetCredentials().GetUsername(),
			resource.GetCredentials().GetPassword(),
			resource.GetCredentials().GetHost(),
			resource.GetCredentials().GetPort(),
			resource.GetCredentials().GetDatabase(),
		)

		// Step 1: run migrations (opens sql.DB, migrates, closes sql.DB).
		// This connection is temporary and freed before the pool is created.
		if err = runMigrationsForDB(dbURL); err != nil {
			b.log.Error("!!!AutoConnect->MigrateUp "+resource.GetProjectId(), logger.Error(err))
			continue
		}

		// Step 2: create runtime pool (lazy connections, no immediate ping).
		_, err = b.Reconnect(ctx, &nb.RegisterProjectRequest{
			Credentials: &nb.RegisterProjectRequest_Credentials{
				Host:     resource.GetCredentials().GetHost(),
				Port:     resource.GetCredentials().GetPort(),
				Database: resource.GetCredentials().GetDatabase(),
				Password: resource.GetCredentials().GetPassword(),
				Username: resource.GetCredentials().GetUsername(),
			},
			ProjectId: resource.GetProjectId(),
		})
		if err != nil {
			b.log.Error("!!!AutoConnect->Reconnect "+resource.GetProjectId(), logger.Error(err))
			continue
		}
	}

	return nil
}
