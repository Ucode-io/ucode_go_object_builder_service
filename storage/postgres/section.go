package postgres

import (
	"context"
	"ucode/ucode_go_object_builder_service/storage"

	"github.com/jackc/pgx/v5/pgxpool"

	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
)

type sectionRepo struct {
	db *pgxpool.Pool
}

func NewSectionRepo(db *pgxpool.Pool) storage.SectionRepoI {
	return &sectionRepo{
		db: db,
	}
}

func (s *sectionRepo) GetViewRelation(ctx context.Context, req *nb.GetAllSectionsRequest) (resp *nb.GetViewRelationResponse, err error) {

	return &nb.GetViewRelationResponse{}, nil
}
