package app

import (
	"context"
	"log/slog"

	"github.com/example/ephemeral-platform/internal/config"
	"github.com/example/ephemeral-platform/internal/controlplane"
	"github.com/example/ephemeral-platform/internal/policy"
	"github.com/example/ephemeral-platform/internal/provider"
	"github.com/example/ephemeral-platform/internal/store"
)

func Build(ctx context.Context, cfg config.Config, logger *slog.Logger) (*controlplane.Service, store.Store, error) {
	repository, err := store.New(ctx, cfg.DatabaseURL, cfg.FileStorePath)
	if err != nil {
		return nil, nil, err
	}
	var infrastructure provider.Provider
	if cfg.Provider == "terraform" {
		infrastructure = provider.Terraform{
			Binary: cfg.TerraformBinary, Directory: cfg.TerraformDirectory,
			DataDirectory: cfg.TerraformDataDirectory, PreviewDomain: cfg.PreviewDomain,
			BackendConfig: cfg.TerraformBackendConfig,
		}
	} else {
		infrastructure = provider.Demo{PreviewDomain: cfg.PreviewDomain, PreviewBaseURL: cfg.DemoPreviewBaseURL, StageDelay: cfg.DemoStageDelay}
	}
	if cfg.PreviewDatabaseURL != "" {
		infrastructure = provider.WithDatabaseSchemas(infrastructure, cfg.PreviewDatabaseURL)
	}
	service := controlplane.New(repository, infrastructure, controlplane.Config{
		PreviewDomain: cfg.PreviewDomain, DefaultImage: cfg.DefaultImage,
		Policy: policy.Policy{
			DefaultTTL: cfg.DefaultTTL, MaxTTL: cfg.MaxTTL,
			MaxCPU: cfg.MaxCPU, MaxMemory: cfg.MaxMemory,
			MaxEnvironmentsPerUser: cfg.MaxEnvironmentsPerUser,
			MaxEnvironmentsPerRepo: cfg.MaxEnvironmentsPerRepo,
		},
		OperationTimeout: cfg.OperationTimeout, WorkerConcurrency: cfg.WorkerConcurrency,
	}, logger)
	return service, repository, nil
}
