package postgres_test

import (
	"context"
	"os"
	"testing"

	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/storage"
	"ucode/ucode_go_object_builder_service/storage/postgres"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/manveru/faker"
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

func TestMain(m *testing.M) {
	cfg = config.Load()
	cfg.PostgresPassword = "1231"
	cfg.PostgresHost = "localhost"
	cfg.PostgresPort = 5432
	cfg.PostgresDatabase = "go_object_builder"
	cfg.PostgresUser = "postgres"

	strg, err = postgres.NewPostgres(context.Background(), cfg)

	fakeData, _ = faker.New("en")

	os.Exit(m.Run())
}
