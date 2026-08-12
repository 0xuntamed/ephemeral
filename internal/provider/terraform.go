package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
)

var safeEnvironmentID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{2,62}$`)

type Terraform struct {
	Binary        string
	Directory     string
	DataDirectory string
	PreviewDomain string
	BackendConfig string
}

func (t Terraform) Name() string { return "terraform" }

func (t Terraform) Provision(ctx context.Context, item *environment.Environment, report Reporter) (ProvisionResult, error) {
	if err := t.validate(item.ID); err != nil {
		return ProvisionResult{}, err
	}
	if _, err := t.run(ctx, item.ID, t.initArgs()...); err != nil {
		return ProvisionResult{}, fmt.Errorf("terraform init: %w", err)
	}
	if _, err := t.run(ctx, item.ID, "workspace", "select", item.TerraformWorkspace); err != nil {
		if _, createErr := t.run(ctx, item.ID, "workspace", "new", item.TerraformWorkspace); createErr != nil {
			return ProvisionResult{}, fmt.Errorf("select or create workspace: %w", createErr)
		}
	}
	if err := report(environment.StatusProvisioning, "Applying the Terraform environment module"); err != nil {
		return ProvisionResult{}, err
	}
	hostName := fmt.Sprintf("%s.%s", item.ID, strings.TrimPrefix(t.PreviewDomain, "."))
	args := []string{
		"apply", "-auto-approve", "-input=false", "-no-color",
		"-var=environment_id=" + item.ID,
		"-var=image=" + item.Image,
		"-var=cpu=" + strconv.Itoa(item.Resources.CPU),
		"-var=memory=" + strconv.Itoa(item.Resources.Memory),
		"-var=database_schema=" + item.DatabaseSchema,
		"-var=host_name=" + hostName,
		"-var=listener_rule_priority=" + strconv.Itoa(item.RoutingPriority),
	}
	if _, err := t.run(ctx, item.ID, args...); err != nil {
		return ProvisionResult{}, fmt.Errorf("terraform apply: %w", err)
	}
	if err := report(environment.StatusHealthCheck, "Terraform completed; checking the preview endpoint"); err != nil {
		return ProvisionResult{}, err
	}
	output, err := t.run(ctx, item.ID, "output", "-json")
	if err != nil {
		return ProvisionResult{}, fmt.Errorf("terraform output: %w", err)
	}
	values := map[string]struct {
		Value any `json:"value"`
	}{}
	if err := json.Unmarshal([]byte(output), &values); err != nil {
		return ProvisionResult{}, fmt.Errorf("decode terraform output: %w", err)
	}
	previewURL := "https://" + hostName
	if value, ok := values["preview_url"]; ok {
		if parsed, ok := value.Value.(string); ok && parsed != "" {
			previewURL = parsed
		}
	}
	if err := waitForPreviewHealth(ctx, strings.TrimRight(previewURL, "/")+"/healthz"); err != nil {
		return ProvisionResult{}, fmt.Errorf("preview health check: %w", err)
	}
	resources := make([]environment.CloudResource, 0, 3)
	for outputName, kind := range map[string]string{
		"ecs_service_arn":   "ecs-service",
		"target_group_arn":  "target-group",
		"listener_rule_arn": "alb-rule",
	} {
		if value, ok := values[outputName]; ok {
			if id, ok := value.Value.(string); ok && id != "" {
				resources = append(resources, environment.CloudResource{Kind: kind, ID: id})
			}
		}
	}
	return ProvisionResult{PreviewURL: previewURL, CloudResources: resources}, nil
}

func (t Terraform) Destroy(ctx context.Context, item *environment.Environment, _ Reporter) error {
	if err := t.validate(item.ID); err != nil {
		return err
	}
	if _, err := t.run(ctx, item.ID, t.initArgs()...); err != nil {
		return fmt.Errorf("terraform init: %w", err)
	}
	if _, err := t.run(ctx, item.ID, "workspace", "select", item.TerraformWorkspace); err != nil {
		return fmt.Errorf("select workspace: %w", err)
	}
	args := []string{
		"destroy", "-auto-approve", "-input=false", "-no-color",
		"-var=environment_id=" + item.ID,
		"-var=image=" + item.Image,
		"-var=cpu=" + strconv.Itoa(item.Resources.CPU),
		"-var=memory=" + strconv.Itoa(item.Resources.Memory),
		"-var=database_schema=" + item.DatabaseSchema,
		"-var=host_name=" + fmt.Sprintf("%s.%s", item.ID, strings.TrimPrefix(t.PreviewDomain, ".")),
		"-var=listener_rule_priority=" + strconv.Itoa(item.RoutingPriority),
	}
	if _, err := t.run(ctx, item.ID, args...); err != nil {
		return fmt.Errorf("terraform destroy: %w", err)
	}
	_, _ = t.run(ctx, item.ID, "workspace", "select", "default")
	_, _ = t.run(ctx, item.ID, "workspace", "delete", item.TerraformWorkspace)
	return t.removeDataDirectory(item.ID)
}

func (t Terraform) validate(id string) error {
	if !safeEnvironmentID.MatchString(id) {
		return fmt.Errorf("unsafe environment id %q", id)
	}
	if strings.TrimSpace(t.Directory) == "" || strings.TrimSpace(t.DataDirectory) == "" {
		return fmt.Errorf("terraform directory and data directory are required")
	}
	return nil
}

func (t Terraform) initArgs() []string {
	args := []string{"init", "-input=false"}
	if strings.TrimSpace(t.BackendConfig) != "" {
		backendConfig, err := filepath.Abs(t.BackendConfig)
		if err == nil {
			args = append(args, "-backend-config="+backendConfig)
		}
	}
	return args
}

func waitForPreviewHealth(ctx context.Context, url string) error {
	client := &http.Client{Timeout: 8 * time.Second}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var lastErr error
	for {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4<<10))
			_ = response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 400 {
				return nil
			}
			lastErr = fmt.Errorf("health endpoint returned %s", response.Status)
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return errors.Join(ctx.Err(), lastErr)
		case <-ticker.C:
		}
	}
}

func (t Terraform) run(ctx context.Context, id string, args ...string) (string, error) {
	dataDirectory, err := t.environmentDataDirectory(id)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dataDirectory, 0o750); err != nil {
		return "", err
	}
	directory, err := filepath.Abs(t.Directory)
	if err != nil {
		return "", err
	}
	commandArgs := append([]string{"-chdir=" + directory}, args...)
	command := exec.CommandContext(ctx, t.Binary, commandArgs...)
	command.Env = append(os.Environ(),
		"TF_DATA_DIR="+dataDirectory,
		"TF_IN_AUTOMATION=1",
		"TF_INPUT=0",
	)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if len(message) > 4000 {
			message = message[len(message)-4000:]
		}
		return "", fmt.Errorf("%w: %s", err, message)
	}
	return string(output), nil
}

func (t Terraform) environmentDataDirectory(id string) (string, error) {
	base, err := filepath.Abs(t.DataDirectory)
	if err != nil {
		return "", err
	}
	target := filepath.Join(base, id)
	relative, err := filepath.Rel(base, target)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || relative == ".." {
		return "", fmt.Errorf("terraform data directory escapes its configured root")
	}
	return target, nil
}

func (t Terraform) removeDataDirectory(id string) error {
	target, err := t.environmentDataDirectory(id)
	if err != nil {
		return err
	}
	return os.RemoveAll(target)
}
