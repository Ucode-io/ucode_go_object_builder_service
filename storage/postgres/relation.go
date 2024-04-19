package postgres

import (
	"context"
	"fmt"

	"ucode/ucode_go_object_builder_service/config"
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
	// conn := psqlpool.Get(req.ProjectId)
	// defer conn.Close()
	var (
		fieldFrom, fieldTo string
	)

	pool, err := pgxpool.ParseConfig("postgres://udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs:oka@65.109.239.69:5432/udevs123_b52a2924bcbe4ab1b6b89f748a2fc500_p_postgres_svcs?sslmode=disable")
	if err != nil {
		return nil, err
	}

	conn, err := pgxpool.NewWithConfig(ctx, pool)
	if err != nil {
		return nil, err
	}

	switch req.Type {
	case config.MANY2DYNAMIC:
	case config.MANY2MANY:
	case config.MANY2ONE:
	case config.ONE2ONE:
	}

	query := `
		INSERT INTO "relation" (
			"id", 
			"table_from", 
			"table_to", 
			"field_from", 
			"field_to", 
			"type",
			"view_fields", 
			"relation_field_slug", 
			"dynamic_tables", 
			"editable",
			"is_user_id_default", 
			"cascadings", 
			"is_system", 
			"object_id_from_jwt",
			"cascading_tree_table_slug", 
			"cascading_tree_field_slug" 
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	_, err = conn.Exec(ctx, query,
		req.ID, req.TableFrom, req.TableTo, req.FieldFrom, req.FieldTo,
		req.Type, req.ViewFields, req.reqFieldSlug, req.DynamicTables,
		req.Editable, req.IsUserIDDefault, req.Cascadings, req.IsSystem,
		req.ObjectIDFromJWT, req.CascadingTreeTableSlug, req.CascadingTreeFieldSlug,
		req.CreatedAt, req.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert relation: %v", err)
	}

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
