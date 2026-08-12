package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddress          string
	DatabaseURL            string
	PreviewDatabaseURL     string
	FileStorePath          string
	APIKey                 string
	Provider               string
	PreviewDomain          string
	DefaultImage           string
	DefaultTTL             time.Duration
	MaxTTL                 time.Duration
	MaxEnvironmentsPerUser int
	MaxEnvironmentsPerRepo int
	MaxCPU                 int
	MaxMemory              int
	WorkerEnabled          bool
	WorkerInterval         time.Duration
	WorkerConcurrency      int
	OperationTimeout       time.Duration
	DemoStageDelay         time.Duration
	DemoPreviewBaseURL     string
	TerraformBinary        string
	TerraformDirectory     string
	TerraformDataDirectory string
	TerraformBackendConfig string
}

func Load() (Config, error) {
	cfg := Config{
		ListenAddress:          env("LISTEN_ADDRESS", ":8080"),
		DatabaseURL:            strings.TrimSpace(os.Getenv("DATABASE_URL")),
		PreviewDatabaseURL:     strings.TrimSpace(os.Getenv("PREVIEW_DATABASE_URL")),
		FileStorePath:          env("FILE_STORE_PATH", "data/environments.json"),
		APIKey:                 strings.TrimSpace(os.Getenv("EPHEMERAL_API_KEY")),
		Provider:               strings.ToLower(env("PROVIDER", "demo")),
		PreviewDomain:          env("PREVIEW_DOMAIN", "preview.localhost"),
		DefaultImage:           env("DEFAULT_IMAGE", "demo-service:latest"),
		TerraformBinary:        env("TERRAFORM_BINARY", "terraform"),
		TerraformDirectory:     env("TERRAFORM_DIR", "terraform/live"),
		TerraformDataDirectory: env("TERRAFORM_DATA_DIR", "terraform/runtime-data"),
		TerraformBackendConfig: strings.TrimSpace(os.Getenv("TERRAFORM_BACKEND_CONFIG")),
		DemoPreviewBaseURL:     env("DEMO_PREVIEW_BASE_URL", "http://localhost:8080/preview"),
	}
	var err error
	if cfg.DefaultTTL, err = duration("DEFAULT_TTL", 6*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.MaxTTL, err = duration("MAX_TTL", 24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.WorkerInterval, err = duration("WORKER_INTERVAL", 2*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.OperationTimeout, err = duration("OPERATION_TIMEOUT", 20*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.DemoStageDelay, err = duration("DEMO_STAGE_DELAY", 600*time.Millisecond); err != nil {
		return Config{}, err
	}
	if cfg.MaxEnvironmentsPerUser, err = integer("MAX_ENVIRONMENTS_PER_USER", 3); err != nil {
		return Config{}, err
	}
	if cfg.MaxEnvironmentsPerRepo, err = integer("MAX_ENVIRONMENTS_PER_REPO", 10); err != nil {
		return Config{}, err
	}
	if cfg.MaxCPU, err = integer("MAX_CPU", 2048); err != nil {
		return Config{}, err
	}
	if cfg.MaxMemory, err = integer("MAX_MEMORY", 4096); err != nil {
		return Config{}, err
	}
	if cfg.WorkerConcurrency, err = integer("WORKER_CONCURRENCY", 4); err != nil {
		return Config{}, err
	}
	if cfg.WorkerEnabled, err = boolean("RUN_WORKER", true); err != nil {
		return Config{}, err
	}
	if cfg.DefaultTTL <= 0 || cfg.MaxTTL < cfg.DefaultTTL {
		return Config{}, fmt.Errorf("MAX_TTL must be greater than or equal to DEFAULT_TTL")
	}
	if cfg.Provider != "demo" && cfg.Provider != "terraform" {
		return Config{}, fmt.Errorf("PROVIDER must be demo or terraform")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return value, nil
}

func integer(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func boolean(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false", key)
	}
	return value, nil
}
