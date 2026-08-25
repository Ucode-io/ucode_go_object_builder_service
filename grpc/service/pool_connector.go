package service

import (
	"context"
	"fmt"
	"time"

	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/genproto/company_service"
	"ucode/ucode_go_object_builder_service/grpc/client"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	psqlpool "ucode/ucode_go_object_builder_service/pool"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPoolConnector returns the lazy reconnect hook for psqlpool.Get: on a pool
// miss it fetches the tenant credentials from the company service (Vault) by
// resource environment id and opens a verified pool, so a tenant skipped by
// startup AutoConnect heals on first request instead of failing with
// "connection not found" until someone calls ReconnectResource manually.
func NewPoolConnector(cfg config.Config, log logger.LoggerI, svcs client.ServiceManagerI) psqlpool.Connector {
	return func(projectId string) (*psqlpool.Pool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		res, err := svcs.ResourceService().AutoConnectByProjectId(ctx, &company_service.AutoConnectByProjectIdRequest{
			ProjectId: projectId,
			NodeType:  cfg.NodeType,
		})
		if err != nil {
			log.Error("!!!PoolConnector->AutoConnectByProjectId", logger.String("resource_env_id", projectId), logger.Error(err))
			return nil, err
		}

		env := res.GetRes()
		if env.GetResourceType() != company_service.ResourceType_POSTGRESQL {
			return nil, fmt.Errorf("unsupported resource type %s for resource environment %s", env.GetResourceType(), projectId)
		}

		creds := env.GetCredentials()
		dbURL := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			creds.GetUsername(),
			creds.GetPassword(),
			creds.GetHost(),
			creds.GetPort(),
			creds.GetDatabase(),
		)

		poolCfg, err := pgxpool.ParseConfig(dbURL)
		if err != nil {
			log.Error("!!!PoolConnector->ParseConfig", logger.String("resource_env_id", projectId), logger.Error(err))
			return nil, err
		}

		poolCfg.MaxConns = cfg.PostgresMaxConnections
		// Disable server-side named prepared statements: they are cached per pgx
		// conn and collide across a connection pooler's shared backends, causing
		// "prepared statement name is already in use" (SQLSTATE 08P01) on tenant DBs.
		poolCfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

		pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
		if err != nil {
			log.Error("!!!PoolConnector->CreatePool", logger.String("resource_env_id", projectId), logger.Error(err))
			return nil, err
		}

		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			log.Error("!!!PoolConnector->Ping", logger.String("resource_env_id", projectId), logger.Error(err))
			return nil, err
		}

		log.Info("!!!PoolConnector--->pool re-established on demand", logger.String("resource_env_id", projectId))

		return &psqlpool.Pool{Db: pool, Logger: log}, nil
	}
}
