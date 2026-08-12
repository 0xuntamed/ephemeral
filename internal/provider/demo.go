package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
)

type Demo struct {
	PreviewDomain  string
	PreviewBaseURL string
	StageDelay     time.Duration
}

func (d Demo) Name() string { return "demo" }

func (d Demo) Provision(ctx context.Context, item *environment.Environment, report Reporter) (ProvisionResult, error) {
	if err := wait(ctx, d.StageDelay); err != nil {
		return ProvisionResult{}, err
	}
	if err := report(environment.StatusProvisioning, "Creating isolated ECS workload and ALB route"); err != nil {
		return ProvisionResult{}, err
	}
	if err := wait(ctx, d.StageDelay); err != nil {
		return ProvisionResult{}, err
	}
	if err := report(environment.StatusHealthCheck, "Waiting for the preview health check"); err != nil {
		return ProvisionResult{}, err
	}
	if err := wait(ctx, d.StageDelay); err != nil {
		return ProvisionResult{}, err
	}
	domain := strings.TrimPrefix(strings.TrimSpace(d.PreviewDomain), ".")
	previewURL := fmt.Sprintf("https://%s.%s", item.ID, domain)
	if strings.TrimSpace(d.PreviewBaseURL) != "" {
		previewURL = strings.TrimRight(d.PreviewBaseURL, "/") + "/" + item.ID
	}
	return ProvisionResult{
		PreviewURL: previewURL,
		CloudResources: []environment.CloudResource{
			{Kind: "ecs-service", ID: "demo/" + item.ID},
			{Kind: "target-group", ID: "demo/" + item.ID},
			{Kind: "alb-rule", ID: fmt.Sprintf("priority/%d", item.RoutingPriority)},
		},
	}, nil
}

func (d Demo) Destroy(ctx context.Context, _ *environment.Environment, _ Reporter) error {
	return wait(ctx, d.StageDelay)
}

func wait(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
