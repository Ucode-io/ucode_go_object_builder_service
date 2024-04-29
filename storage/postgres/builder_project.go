package postgres

import (
	"context"
	"ucode/ucode_go_object_builder_service/storage"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"

	"github.com/jackc/pgx/v5/pgxpool"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
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
