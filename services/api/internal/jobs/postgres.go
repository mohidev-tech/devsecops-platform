package jobs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS jobs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payload      BYTEA NOT NULL,
    status       TEXT NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS jobs_status_created_idx
    ON jobs (status, created_at)
    WHERE status = 'pending';
`

type PostgresRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresRepo(ctx context.Context, dsn string) (*PostgresRepo, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &PostgresRepo{pool: pool}, nil
}

func (p *PostgresRepo) Close() { p.pool.Close() }

func (p *PostgresRepo) Pool() *pgxpool.Pool { return p.pool }

func (p *PostgresRepo) Enqueue(ctx context.Context, payload []byte) (string, error) {
	var id string
	err := p.pool.QueryRow(ctx,
		`INSERT INTO jobs (payload) VALUES ($1) RETURNING id`,
		payload,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("insert: %w", err)
	}
	return id, nil
}

func (p *PostgresRepo) Count(ctx context.Context, status Status) (int, error) {
	var n int
	err := p.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs WHERE status = $1`,
		string(status),
	).Scan(&n)
	return n, err
}
