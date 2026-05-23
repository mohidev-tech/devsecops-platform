package jobs

import (
	"context"
	"errors"
	"time"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusDone    Status = "done"
	StatusFailed  Status = "failed"
)

type Job struct {
	ID          string
	Payload     []byte
	Status      Status
	CreatedAt   time.Time
	ProcessedAt *time.Time
}

// Repo is the boundary the api and worker depend on. Production: Postgres.
// Tests: in-memory fake.
type Repo interface {
	Enqueue(ctx context.Context, payload []byte) (string, error)
	Count(ctx context.Context, status Status) (int, error)
}

var ErrNotFound = errors.New("job not found")
