package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/example/ephemeral-platform/internal/app"
	"github.com/example/ephemeral-platform/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	if cfg.DatabaseURL == "" {
		logger.Error("standalone worker requires DATABASE_URL so API and workers share atomic state")
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	service, repository, err := app.Build(ctx, cfg, logger)
	if err != nil {
		logger.Error("initialize worker", "error", err)
		os.Exit(1)
	}
	defer repository.Close()
	service.RunWorker(ctx, cfg.WorkerInterval)
}
