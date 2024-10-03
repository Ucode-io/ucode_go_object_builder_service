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

	// err = b.strg.BuilderProject().Register(ctx, req)
	// if err != nil {
	// 	b.log.Error("!!!RegisterProjectErrorBuilder--->", logger.Error(err))
	// 	return resp, err
	// }

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
		b.log.Error("!!!RegisterProject->ParseResourceCredentials", logger.Error(err))
		return resp, err
	}

	config.MaxConns = b.cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		b.log.Error("!!!RegisterProject->CreatePool", logger.Error(err))
		return resp, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		b.log.Error("!!!RegisterProject->Ping", logger.Error(err))
		return resp, err
	}

	dbInstance, err := sql.Open("postgres", dbUrl)
	if err != nil {
		b.log.Error("!!!RegisterProject->OpenDB", logger.Error(err))
		return resp, err
	}

	db, err := postgres.WithInstance(dbInstance, &postgres.Config{})
	if err != nil {
		b.log.Error("!!!RegisterProject->WithInstance", logger.Error(err))
		return resp, err
	}

	migrationsDir := "file://migrations/postgres"
	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		db,
	)
	if err != nil {
		b.log.Error("!!!RegisterProject->NewWithDatabaseInstance", logger.Error(err))
		return resp, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		b.log.Error("!!!RegisterProject->MigrateUp", logger.Error(err))
		return resp, err
	}

	b.log.Info("::::::::::::::::Migration completed successfully::::::::::::::::")

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

func (b *builderProjectService) Reconnect(ctx context.Context, req *nb.RegisterProjectRequest) (resp *emptypb.Empty, err error) {
	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		req.Credentials.GetUsername(),
		req.Credentials.GetPassword(),
		req.Credentials.GetHost(),
		req.Credentials.GetPort(),
		req.Credentials.GetDatabase(),
	)

	b.log.Info("!!!Reconnect--->", logger.Any("dbURL", dbURL))

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		b.log.Error("!!!Reconnect->ParseResourceCredentials", logger.Error(err))
		return resp, err
	}

	config.MaxConns = b.cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		b.log.Error("!!!Reconnect->CreatePool", logger.Error(err))
		return resp, err
	}

	dbInstance, err := sql.Open("postgres", dbURL)
	if err != nil {
		b.log.Error("!!!RegisterProject->OpenDB", logger.Error(err))
		return resp, err
	}

	db, err := postgres.WithInstance(dbInstance, &postgres.Config{})
	if err != nil {
		b.log.Error("!!!ReconnectProject->WithInstance", logger.Error(err))
		return resp, nil
	}

	migrationsDir := "file://migrations/postgres"
	m, err := migrate.NewWithDatabaseInstance(
		migrationsDir,
		"postgres",
		db,
	)
	if err != nil {
		b.log.Error("!!!RegisterProject->NewWithDatabaseInstance", logger.Error(err))
		return resp, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		b.log.Error("!!!RegisterProject->MigrateUp", logger.Error(err))
		return resp, err
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
		if resource.ResourceType != company_service.ResourceType_POSTGRESQL || resource.GetCredentials().GetDatabase() == "час_05f2e3c68af248c68e25d5bdd8954bf5_p_postgres_svcs" {
			continue
		}

		// if resource.Credentials.Username == "ligth_c1240467739b4c07b8c86c49546dbf87_p_postgres_svcs" {
		b.log.Info(
			fmt.Sprintf(
				"postgresql://%v:%v@%v:%v/%v?sslmode=disable",
				resource.Credentials.Database,
				resource.Credentials.Password,
				resource.Credentials.Host,
				resource.Credentials.Port,
				resource.Credentials.Username,
			),
		)

		_, err = b.Reconnect(ctx, &nb.RegisterProjectRequest{
			Credentials: &nb.RegisterProjectRequest_Credentials{
				Host:     resource.GetCredentials().GetHost(),
				Port:     resource.GetCredentials().GetPort(),
				Database: resource.GetCredentials().GetDatabase(),
				Password: resource.GetCredentials().GetPassword(),
				Username: resource.GetCredentials().GetUsername(),
			},
			ProjectId: resource.GetProjectId(),
			// K8SNamespace: resource,
		})
		if err != nil {
			b.log.Error("!!!AutoConnect-->Reconnect", logger.Error(err))
			return err
		}
		// }
	}

	return nil
}

//
