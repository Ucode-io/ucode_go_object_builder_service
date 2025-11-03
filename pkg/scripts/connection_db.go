package scripts

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectToCompanyService() (*pgxpool.Pool, error) {

	var dbURL = "postgresql://company_service:uVah9foo@95.217.155.57:30034/company_service"

	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
