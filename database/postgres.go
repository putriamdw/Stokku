package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stokku/config"
)

// NewPool membuat pool koneksi PostgreSQL dan memastikan database terjangkau
func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
	url := config.GetEnv("DATABASE_URL", "")
	if url == "" {
		return nil, errors.New("DATABASE_URL belum diisi")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}