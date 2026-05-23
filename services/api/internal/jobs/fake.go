package jobs

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// FakeRepo is an in-memory Repo for tests and the no-DB local smoke path.
type FakeRepo struct {
	mu   sync.Mutex
	next int
	jobs map[string]*Job
}

func NewFakeRepo() *FakeRepo {
	return &FakeRepo{jobs: map[string]*Job{}}
}

func (f *FakeRepo) Enqueue(_ context.Context, payload []byte) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	id := strconv.Itoa(f.next)
	f.jobs[id] = &Job{
		ID:        id,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}
	return id, nil
}

func (f *FakeRepo) Count(_ context.Context, status Status) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, j := range f.jobs {
		if j.Status == status {
			n++
		}
	}
	return n, nil
}
