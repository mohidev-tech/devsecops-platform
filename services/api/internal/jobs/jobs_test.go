package jobs

import (
	"context"
	"testing"
)

func TestFakeRepoRoundtrip(t *testing.T) {
	ctx := context.Background()
	r := NewFakeRepo()
	if n, _ := r.Count(ctx, StatusPending); n != 0 {
		t.Fatalf("expected 0 pending, got %d", n)
	}
	id, err := r.Enqueue(ctx, []byte(`{"x":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Fatal("empty id")
	}
	if n, _ := r.Count(ctx, StatusPending); n != 1 {
		t.Fatalf("expected 1 pending, got %d", n)
	}
}
