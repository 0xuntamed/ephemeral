package controlplane

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
	"github.com/example/ephemeral-platform/internal/policy"
	"github.com/example/ephemeral-platform/internal/provider"
	"github.com/example/ephemeral-platform/internal/store"
)

func TestCreateIsIdempotentAndBecomesReady(t *testing.T) {
	service := newTestService(t)
	input := validCreateInput()
	first, created, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("create first environment: %v", err)
	}
	if !created {
		t.Fatal("first request was not reported as created")
	}
	second, created, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatalf("replay create request: %v", err)
	}
	if created {
		t.Fatal("replayed request created a duplicate")
	}
	if first.ID != second.ID {
		t.Fatalf("idempotent request returned %q, want %q", second.ID, first.ID)
	}
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("process provisioning: %v", err)
	}
	ready, err := service.Get(context.Background(), first.ID)
	if err != nil {
		t.Fatalf("get ready environment: %v", err)
	}
	if ready.Status != environment.StatusReady {
		t.Fatalf("status = %s, want READY; last error: %s", ready.Status, ready.LastError)
	}
	if ready.PreviewURL == "" || ready.ReadyAt == nil {
		t.Fatalf("ready environment lacks preview metadata: %#v", ready)
	}
}

func TestDestroyIsAsynchronousAndIdempotent(t *testing.T) {
	service := newTestService(t)
	item, _, err := service.Create(context.Background(), validCreateInput())
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	requested, err := service.RequestDestroy(context.Background(), item.ID, "test cleanup")
	if err != nil {
		t.Fatal(err)
	}
	if requested.DesiredState != environment.DesiredDestroyed {
		t.Fatalf("desired state = %s, want DESTROYED", requested.DesiredState)
	}
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	destroyed, err := service.Get(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if destroyed.Status != environment.StatusDestroyed || destroyed.DestroyedAt == nil {
		t.Fatalf("environment was not destroyed: %#v", destroyed)
	}
	if _, err := service.RequestDestroy(context.Background(), item.ID, "duplicate cleanup"); err != nil {
		t.Fatalf("duplicate destroy request should be harmless: %v", err)
	}
}

func TestTTLReaperSchedulesCleanup(t *testing.T) {
	service := newTestService(t)
	now := time.Date(2026, 8, 12, 8, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	input := validCreateInput()
	input.TTL = time.Hour
	item, _, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Hour)
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	// The first pass records desired state and the next pass performs cleanup.
	if err := service.ProcessOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	expired, err := service.Get(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if expired.Status != environment.StatusDestroyed {
		t.Fatalf("expired status = %s, want DESTROYED", expired.Status)
	}
}

func TestCapacityPolicyRejectsFourthActiveEnvironment(t *testing.T) {
	service := newTestService(t)
	for pr := 1; pr <= 3; pr++ {
		input := validCreateInput()
		input.PullRequest = pr
		input.CommitSHA = "abcdef" + string(rune('0'+pr))
		if _, _, err := service.Create(context.Background(), input); err != nil {
			t.Fatalf("create environment %d: %v", pr, err)
		}
	}
	input := validCreateInput()
	input.PullRequest = 4
	input.CommitSHA = "abcdef4"
	if _, _, err := service.Create(context.Background(), input); err == nil {
		t.Fatal("fourth active environment was accepted")
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()
	repository, err := store.NewFileStore("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repository.Close() })
	return New(repository, provider.Demo{PreviewDomain: "preview.test", StageDelay: time.Millisecond}, Config{
		PreviewDomain: "preview.test", DefaultImage: "demo:latest",
		Policy: policy.Policy{
			DefaultTTL: 6 * time.Hour, MaxTTL: 24 * time.Hour,
			MaxCPU: 2048, MaxMemory: 4096,
			MaxEnvironmentsPerUser: 3, MaxEnvironmentsPerRepo: 10,
		},
		OperationTimeout: time.Second, WorkerConcurrency: 2,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func validCreateInput() CreateInput {
	return CreateInput{
		Repository: "acme/billing-api", Branch: "feature/checkout",
		CommitSHA: "a81f2aa", PullRequest: 127, Owner: "bilal",
		Resources: environment.Resources{CPU: 512, Memory: 1024, DatabaseMode: environment.DatabaseSharedSchema},
	}
}
