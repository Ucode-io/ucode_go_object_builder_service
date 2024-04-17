package postgres

// import (
// 	"context"
// 	nb "ucode/ucode_go_object_builder_service/genproto/new_object_builder_service"
// 	"ucode/ucode_go_object_builder_service/storage"

// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// type LayoutRepo struct {
// 	db *pgxpool.Pool
// }

// func NewLayoutRepo(db *pgxpool.Pool) storage.LayoutRepoI {
// 	return &LayoutRepo{
// 		db: db,
// 	}
// }

// func (r *LayoutRepo) CreateAll(ctx context.Context, req *nb.LayoutRequest) (resp *nb.GetListLayoutResponse, err error) {

// 	tx, err := conn.Begin(ctx)
// 	if err != nil {
// 		return &nb.GetListLayoutResponse{}, err
// 	}

// 	_, err = tx.Exec(ctx, "UPDATE table SET is_changed = true WHERE id = $1", req.)
// 	if err != nil {
// 		tx.Rollback(ctx)
// 		return nil, err
// 	}


// 	return &nb.GetListLayoutResponse{}, nil
// }
