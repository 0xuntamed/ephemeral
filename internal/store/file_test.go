package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
)

func TestFileStorePersistsAndReplaysIdempotencyKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "environments.json")
	first, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	item := &environment.Environment{
		ID: "billing-pr-1-abcdef0", IdempotencyKey: "acme/billing:1:abcdef0",
		Status: environment.StatusRequested, DesiredState: environment.DesiredReady,
		CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(time.Hour), NextAttemptAt: now,
	}
	if _, created, err := first.Create(context.Background(), item); err != nil || !created {
		t.Fatalf("create: created=%v err=%v", created, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := NewFileStore(path)
	if err != nil {
		t.Fatal(err)
	}
	replayed := item.Clone()
	replayed.ID = "different-id"
	saved, created, err := second.Create(context.Background(), replayed)
	if err != nil {
		t.Fatal(err)
	}
	if created || saved.ID != item.ID {
		t.Fatalf("idempotency replay = (%v, %q), want (false, %q)", created, saved.ID, item.ID)
	}
}
