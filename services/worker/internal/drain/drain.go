package drain

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Once claims up to `batchSize` pending jobs using FOR UPDATE SKIP LOCKED and
// marks them done. Returns the number drained. Safe to run concurrently across
// many worker pods — Postgres serializes the row locks.
func Once(ctx context.Context, pool *pgxpool.Pool, batchSize int) (int, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rows, err := tx.Query(ctx, `
		SELECT id FROM jobs
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, batchSize)
	if err != nil {
		return 0, fmt.Errorf("select: %w", err)
	}

	ids := make([]string, 0, batchSize)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return 0, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return 0, nil
	}

	ct, err := tx.Exec(ctx, `
		UPDATE jobs SET status = 'done', processed_at = now()
		WHERE id = ANY($1::uuid[])
	`, ids)
	if err != nil {
		return 0, fmt.Errorf("update: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return int(ct.RowsAffected()), nil
}
