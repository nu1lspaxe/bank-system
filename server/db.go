package server

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetPGConn(ctx context.Context, dbLink string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dbLink)
	if err != nil {
		return nil, err
	}

	if connErr := pool.Ping(ctx); connErr != nil {
		return nil, connErr
	}

	return pool, nil
}
