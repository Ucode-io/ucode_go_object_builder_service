package postgres

import (
	"context"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"
)

type relationRepo struct {
	db *pgxpool.Pool
}

func NewRelationRepo(db *pgxpool.Pool) storage.RelationRepoI {
	return &relationRepo{
		db: db,
	}
}

func (r *relationRepo) Create(ctx context.Context, req *nb.CreateRelationRequest) (resp *nb.CreateRelationRequest, err error) {
	return resp, nil
}

func (r *relationRepo) GetByID(ctx context.Context, req *nb.RelationPrimaryKey) (resp *nb.RelationForGetAll, err error) {
	return resp, nil
}

func (r *relationRepo) GetList(ctx context.Context, req *nb.GetAllRelationsRequest) (resp *nb.GetAllRelationsResponse, err error) {
	return resp, nil
}

func (r *relationRepo) Update(ctx context.Context, req *nb.UpdateRelationRequest) (resp *nb.RelationForGetAll, err error) {
	return resp, err
}

func (r *relationRepo) Delete(ctx context.Context, req *nb.RelationPrimaryKey) (err error) {
	return nil
}
