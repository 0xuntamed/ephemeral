package provider

import (
	"context"

	"github.com/example/ephemeral-platform/internal/environment"
)

type Reporter func(status environment.Status, message string) error

type ProvisionResult struct {
	PreviewURL     string
	CloudResources []environment.CloudResource
}

type Provider interface {
	Name() string
	Provision(ctx context.Context, item *environment.Environment, report Reporter) (ProvisionResult, error)
	Destroy(ctx context.Context, item *environment.Environment, report Reporter) error
}
