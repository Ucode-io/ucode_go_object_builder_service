package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"ucode/ucode_go_object_builder_service/config"
	psqlpool "ucode/ucode_go_object_builder_service/pool"
	"ucode/ucode_go_object_builder_service/storage"
	"ucode/ucode_go_object_builder_service/storage/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/manveru/faker"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

var (
	err      error
	cfg      config.Config
	strg     storage.StorageI
	fakeData *faker.Faker
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
	cfg = config.Load()
	cfg.PostgresHost = cast.ToString(getOrReturnDefaultValue("POSTGRES_HOST", ""))
	cfg.PostgresPort = cast.ToInt(getOrReturnDefaultValue("POSTGRES_PORT", 5432))
	cfg.PostgresUser = cast.ToString(getOrReturnDefaultValue("POSTGRES_USER", ""))
	cfg.PostgresPassword = cast.ToString(getOrReturnDefaultValue("POSTGRES_PASSWORD", ""))
	cfg.PostgresDatabase = cast.ToString(getOrReturnDefaultValue("POSTGRES_DATABASE", ""))

	strg, err = postgres.NewPostgres(context.Background(), cfg, nil)
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

	psqlpool.Add("f0259839-c2fc-44e8-af90-1a6aa7ba43f7", &psqlpool.Pool{Db: pool})

	fakeData, err = faker.New("en")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func getOrReturnDefaultValue(key string, defaultValue any) any {
	val, exists := os.LookupEnv(key)

	if exists {
		return val
	}

	return defaultValue
}
