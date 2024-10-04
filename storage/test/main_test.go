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
	cfg.PostgresPassword = "oka"
	cfg.PostgresHost = "65.109.239.69"
	cfg.PostgresPort = 5432
	cfg.PostgresDatabase = "udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs"
	cfg.PostgresUser = "udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs"

	strg, err = postgres.NewPostgres(context.Background(), cfg, nil)

	fakeData, _ = faker.New("en")

	os.Exit(m.Run())
}
