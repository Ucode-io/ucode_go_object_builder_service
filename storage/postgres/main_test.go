package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/pkg/logger"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"
	"ucode/ucode_go_object_builder_service/storage/postgres"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jaswdr/faker/v2"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

const (
	roleId    = "5381a752-0652-4da2-acfc-0dea5082a21e"
	projectId = "a95e3cba-bfc0-46c8-aeaa-ef75ee9986ed"
)

var (
	err      error
	cfg      config.Config
	strg     storage.StorageI
	fakeData faker.Faker
)

// POSTGRES_HOST="65.109.239.69"
// POSTGRES_PORT=5432
// POSTGRES_DATABASE="company_service"
// POSTGRES_USER="company_service"
// POSTGRES_PASSWORD="fgd4dfFFDJFSd23o"

func CreateRandomId(t *testing.T) string {
	id, err := uuid.NewRandom()
	assert.NoError(t, err)
	return id.String()
}

// the code should take the config from the environment
func TestMain(m *testing.M) {
	// cfg.PostgresDatabase = "airbyte_367933c14b1d47da8185b5a92b3e4f75_p_postgres_svcs"
	// cfg.PostgresHost = "95.217.155.57"
	// cfg.PostgresUser = "airbyte_367933c14b1d47da8185b5a92b3e4f75_p_postgres_svcs"
	// cfg.PostgresPassword = "g644bblsP3"
	// cfg.PostgresPort = 30034

	var (
		loggerLevel string
		cfg         = config.Load()
	)

	cfg.PostgresUser = "company_service"
	cfg.PostgresPassword = "uVah9foo"
	cfg.PostgresHost = "postgresql01.u-code.io"
	cfg.PostgresPort = 30032
	cfg.PostgresDatabase = "test"
	cfg.PostgresMaxConnections = 10

	switch cfg.Environment {
	case config.DebugMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.DebugMode)
	case config.TestMode:
		loggerLevel = logger.LevelDebug
		gin.SetMode(gin.TestMode)
	default:
		loggerLevel = logger.LevelInfo
		gin.SetMode(gin.ReleaseMode)
	}

	log := logger.NewLogger(cfg.ServiceName, loggerLevel)
	defer func() {
		_ = logger.Cleanup(log)
	}()

	strg, err = postgres.NewPostgres(context.Background(), cfg, nil, log)
	if err != nil {
		panic(err)
	}

	dbURL := fmt.Sprintf(
		"postgres://%v:%v@%v:%v/%v?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
	)

	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		panic(err)
	}

	config.MaxConns = cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	psqlpool.Add(projectId, &psqlpool.Pool{Db: pool})

	fakeData = faker.New()

	os.Exit(m.Run())
}
