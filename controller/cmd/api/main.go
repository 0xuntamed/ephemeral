package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/ephemeral-platform/internal/app"
	"github.com/example/ephemeral-platform/internal/config"
	"github.com/example/ephemeral-platform/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	cfg, err := config.Load()
	if err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	if err := cfg.ValidateAPI(); err != nil {
		logger.Error("load configuration", "error", err)
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	service, repository, err := app.Build(ctx, cfg, logger)
	if err != nil {
		logger.Error("initialize control plane", "error", err)
		os.Exit(1)
	}
	defer repository.Close()
	if cfg.WorkerEnabled {
		go service.RunWorker(ctx, cfg.WorkerInterval)
	}
	server := &http.Server{
		Addr: cfg.ListenAddress, Handler: httpapi.New(service, cfg.APIKey, logger),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 30 * time.Minute, IdleTimeout: 60 * time.Second,
	}
	go func() {
		logger.Info("api listening", "address", cfg.ListenAddress, "provider", cfg.Provider)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("api server", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown", "error", err)
	}
}
