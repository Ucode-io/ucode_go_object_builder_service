package postgres

import (
	"context"
	"fmt"
	"log"
	"ucode/ucode_go_object_builder_service/config"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	db             *pgxpool.Pool
	builderProject storage.BuilderProjectRepoI
	field          storage.FieldRepoI
	function       storage.FunctionRepoI
	file           storage.FieldRepoI
}

func NewPostgres(ctx context.Context, cfg config.Config) (storage.StorageI, error) {
	config, err := pgxpool.ParseConfig(fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDatabase,
	))
	if err != nil {
		return nil, err
	}

	config.MaxConns = cfg.PostgresMaxConnections

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	return &Store{
		db: pool,
	}, err
}

func (s *Store) CloseDB() {
	s.db.Close()
}

func (l *Store) Log(ctx context.Context, msg string, data map[string]interface{}) {
	args := make([]interface{}, 0, len(data)+2) // making space for arguments + msg
	args = append(args, msg)
	for k, v := range data {
		args = append(args, fmt.Sprintf("%s=%v", k, v))
	}
	log.Println(args...)
}

func (s *Store) BuilderProject() storage.BuilderProjectRepoI {
	if s.builderProject == nil {
		s.builderProject = NewBuilderProjectRepo(s.db)
	}

	return s.builderProject
}

func (s *Store) Field() storage.FieldRepoI {
	if s.field == nil {
		s.field = NewFieldRepo(s.db)
	}

	return s.field
}

func (s *Store) Function() storage.FunctionRepoI {
	if s.function == nil {
		s.function = NewFunctionRepo(s.db)
	}

	return s.function
}

func (s *Store) File()
