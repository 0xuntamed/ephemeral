package controlplane

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/fnv"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
	"github.com/example/ephemeral-platform/internal/policy"
	"github.com/example/ephemeral-platform/internal/provider"
	"github.com/example/ephemeral-platform/internal/store"
)

var (
	ErrInvalidInput   = errors.New("invalid input")
	ErrInvalidState   = errors.New("invalid environment state")
	errNotCandidate   = errors.New("not a lifecycle candidate")
	repositoryPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)
	commitPattern     = regexp.MustCompile(`^[A-Fa-f0-9]{7,64}$`)
)

type Config struct {
	PreviewDomain     string
	DefaultImage      string
	Policy            policy.Policy
	OperationTimeout  time.Duration
	WorkerConcurrency int
}

type CreateInput struct {
	Repository     string
	Branch         string
	CommitSHA      string
	PullRequest    int
	Owner          string
	Image          string
	TTL            time.Duration
	Resources      environment.Resources
	IdempotencyKey string
}

type Counters struct {
	created            atomic.Uint64
	idempotentHits     atomic.Uint64
	provisionSucceeded atomic.Uint64
	provisionFailed    atomic.Uint64
	destroySucceeded   atomic.Uint64
	destroyFailed      atomic.Uint64
}

type Service struct {
	store     store.Store
	provider  provider.Provider
	config    Config
	logger    *slog.Logger
	now       func() time.Time
	wake      chan struct{}
	processMu sync.Mutex
	counters  Counters
}

func New(repository store.Store, infrastructure provider.Provider, cfg Config, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.WorkerConcurrency < 1 {
		cfg.WorkerConcurrency = 1
	}
	return &Service{
		store: repository, provider: infrastructure, config: cfg, logger: logger,
		now: time.Now, wake: make(chan struct{}, 1),
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*environment.Environment, bool, error) {
	input.Repository = strings.TrimSpace(input.Repository)
	input.Branch = strings.TrimSpace(input.Branch)
	input.CommitSHA = strings.ToLower(strings.TrimSpace(input.CommitSHA))
	input.Owner = strings.TrimSpace(input.Owner)
	input.Image = strings.TrimSpace(input.Image)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if !repositoryPattern.MatchString(input.Repository) {
		return nil, false, fmt.Errorf("%w: repository must use owner/name format", ErrInvalidInput)
	}
	if input.Branch == "" || len(input.Branch) > 200 {
		return nil, false, fmt.Errorf("%w: branch is required and must be at most 200 characters", ErrInvalidInput)
	}
	if !commitPattern.MatchString(input.CommitSHA) {
		return nil, false, fmt.Errorf("%w: commit_sha must be a 7-64 character hexadecimal SHA", ErrInvalidInput)
	}
	if input.PullRequest <= 0 {
		return nil, false, fmt.Errorf("%w: pull_request must be positive", ErrInvalidInput)
	}
	if input.Owner == "" {
		return nil, false, fmt.Errorf("%w: owner is required", ErrInvalidInput)
	}
	if input.Image == "" {
		input.Image = s.config.DefaultImage
	}
	if input.TTL == 0 {
		input.TTL = s.config.Policy.DefaultTTL
	}
	if input.Resources.CPU == 0 {
		input.Resources.CPU = 512
	}
	if input.Resources.Memory == 0 {
		input.Resources.Memory = 1024
	}
	if input.Resources.DatabaseMode == "" {
		input.Resources.DatabaseMode = environment.DatabaseSharedSchema
	}
	if err := s.config.Policy.ValidateRequest(input.TTL, input.Resources); err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if input.IdempotencyKey == "" {
		input.IdempotencyKey = fmt.Sprintf("%s:%d:%s", strings.ToLower(input.Repository), input.PullRequest, input.CommitSHA)
	}
	// The idempotency lookup belongs before capacity enforcement: a duplicate webhook
	// must return its original environment even when the owner is already at quota.
	existing, err := s.store.List(ctx, store.Filter{})
	if err != nil {
		return nil, false, err
	}
	for _, item := range existing {
		if item.IdempotencyKey == input.IdempotencyKey {
			s.counters.idempotentHits.Add(1)
			return item, false, nil
		}
	}
	if err := s.config.Policy.ValidateCapacity(existing, input.Owner, input.Repository); err != nil {
		return nil, false, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	now := s.now().UTC()
	id, err := environmentID(input.Repository, input.PullRequest, input.CommitSHA)
	if err != nil {
		return nil, false, err
	}
	databaseSchema := ""
	if input.Resources.DatabaseMode == environment.DatabaseSharedSchema {
		databaseSchema = strings.ReplaceAll("preview_"+id, "-", "_")
		if len(databaseSchema) > 63 {
			databaseSchema = strings.TrimRight(databaseSchema[:63], "_")
		}
	}
	item := &environment.Environment{
		ID: id, IdempotencyKey: input.IdempotencyKey, Repository: input.Repository,
		Branch: input.Branch, CommitSHA: input.CommitSHA, PullRequest: input.PullRequest,
		Owner: input.Owner, Image: input.Image, Status: environment.StatusRequested,
		DesiredState: environment.DesiredReady, StatusMessage: "Environment request accepted",
		TerraformWorkspace: id, DatabaseSchema: databaseSchema,
		RoutingPriority: routingPriority(input.IdempotencyKey), Resources: input.Resources,
		CloudResources:      make([]environment.CloudResource, 0),
		EstimatedHourlyCost: estimateHourlyCost(input.Resources), CreatedAt: now,
		UpdatedAt: now, ExpiresAt: now.Add(input.TTL), NextAttemptAt: now,
	}
	item.AddEvent(item.Status, item.StatusMessage, now)
	saved, created, err := s.store.Create(ctx, item)
	if err != nil {
		return nil, false, err
	}
	if created {
		s.counters.created.Add(1)
		s.signalWorker()
	} else {
		s.counters.idempotentHits.Add(1)
	}
	return saved, created, nil
}

func (s *Service) Get(ctx context.Context, id string) (*environment.Environment, error) {
	return s.store.Get(ctx, id)
}

func (s *Service) List(ctx context.Context, filter store.Filter) ([]*environment.Environment, error) {
	return s.store.List(ctx, filter)
}

func (s *Service) RequestDestroy(ctx context.Context, id, reason string) (*environment.Environment, error) {
	now := s.now().UTC()
	item, err := s.store.Update(ctx, id, func(item *environment.Environment) error {
		if item.Status == environment.StatusDestroyed || item.Status == environment.StatusFailedCleaned {
			return nil
		}
		item.DesiredState = environment.DesiredDestroyed
		item.NextAttemptAt = now
		item.UpdatedAt = now
		if strings.TrimSpace(reason) == "" {
			reason = "Destruction requested"
		}
		item.AddEvent(item.Status, reason, now)
		return nil
	})
	if err == nil {
		s.signalWorker()
	}
	return item, err
}

func (s *Service) Extend(ctx context.Context, id string, extension time.Duration) (*environment.Environment, error) {
	if extension <= 0 {
		return nil, fmt.Errorf("%w: extension must be positive", ErrInvalidInput)
	}
	now := s.now().UTC()
	return s.store.Update(ctx, id, func(item *environment.Environment) error {
		if item.DesiredState != environment.DesiredReady || item.Status == environment.StatusDestroyed {
			return fmt.Errorf("%w: only active environments can be extended", ErrInvalidState)
		}
		maximum := item.CreatedAt.Add(s.config.Policy.MaxTTL)
		candidate := item.ExpiresAt.Add(extension)
		if candidate.After(maximum) {
			return fmt.Errorf("%w: extension exceeds the absolute maximum lifetime of %s", ErrInvalidInput, s.config.Policy.MaxTTL)
		}
		item.ExpiresAt = candidate
		item.UpdatedAt = now
		item.AddEvent(item.Status, "TTL extended by "+extension.String(), now)
		return nil
	})
}

func (s *Service) Retry(ctx context.Context, id string) (*environment.Environment, error) {
	now := s.now().UTC()
	item, err := s.store.Update(ctx, id, func(item *environment.Environment) error {
		switch item.Status {
		case environment.StatusFailed:
			if err := item.SetStatus(environment.StatusRequested, "Provisioning retry requested", now); err != nil {
				return err
			}
			item.DesiredState = environment.DesiredReady
		case environment.StatusDestroyFailed, environment.StatusOrphaned:
			item.DesiredState = environment.DesiredDestroyed
			item.NextAttemptAt = now
			item.UpdatedAt = now
			item.AddEvent(item.Status, "Cleanup retry requested", now)
		default:
			return fmt.Errorf("%w: status %s cannot be retried", ErrInvalidState, item.Status)
		}
		item.LastError = ""
		item.NextAttemptAt = now
		return nil
	})
	if err == nil {
		s.signalWorker()
	}
	return item, err
}

func (s *Service) RunWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.logger.Info("lifecycle worker started", "provider", s.provider.Name(), "interval", interval)
	for {
		if err := s.ProcessOnce(ctx); err != nil && !errors.Is(err, context.Canceled) {
			s.logger.Error("lifecycle reconciliation failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.wake:
		}
	}
}

func (s *Service) TriggerReconcile() {
	s.signalWorker()
}

func (s *Service) ProcessOnce(ctx context.Context) error {
	if !s.processMu.TryLock() {
		return nil
	}
	defer s.processMu.Unlock()
	items, err := s.store.List(ctx, store.Filter{})
	if err != nil {
		return err
	}
	now := s.now().UTC()
	for _, item := range items {
		if item.DesiredState == environment.DesiredReady && item.Status != environment.StatusDestroyed && !item.ExpiresAt.After(now) {
			if _, err := s.RequestDestroy(ctx, item.ID, "TTL expired; cleanup scheduled by reconciler"); err != nil {
				s.logger.Error("schedule expired environment cleanup", "environment_id", item.ID, "error", err)
			}
			continue
		}
		if isInFlight(item.Status) && now.Sub(item.UpdatedAt) > s.config.OperationTimeout {
			_, err := s.store.Update(ctx, item.ID, func(current *environment.Environment) error {
				if !isInFlight(current.Status) || now.Sub(current.UpdatedAt) <= s.config.OperationTimeout {
					return errNotCandidate
				}
				if err := current.SetStatus(environment.StatusOrphaned, "Operation lease expired; cleanup scheduled", now); err != nil {
					return err
				}
				current.DesiredState = environment.DesiredDestroyed
				current.NextAttemptAt = now
				return nil
			})
			if err != nil && !errors.Is(err, errNotCandidate) {
				s.logger.Error("mark stale environment orphaned", "environment_id", item.ID, "error", err)
			}
		}
	}
	items, err = s.store.List(ctx, store.Filter{})
	if err != nil {
		return err
	}
	sem := make(chan struct{}, s.config.WorkerConcurrency)
	var wg sync.WaitGroup
	for _, item := range items {
		if item.NextAttemptAt.After(now) {
			continue
		}
		shouldProvision := item.DesiredState == environment.DesiredReady && item.Status == environment.StatusRequested
		shouldDestroy := item.DesiredState == environment.DesiredDestroyed && item.Status != environment.StatusDestroyed && item.Status != environment.StatusDestroying && item.Status != environment.StatusFailedCleaned
		if !shouldProvision && !shouldDestroy {
			continue
		}
		item := item
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}
			var operationErr error
			if shouldProvision {
				operationErr = s.processProvision(ctx, item.ID)
			} else {
				operationErr = s.processDestroy(ctx, item.ID)
			}
			if operationErr != nil && !errors.Is(operationErr, errNotCandidate) && !errors.Is(operationErr, context.Canceled) {
				s.logger.Error("environment operation failed", "environment_id", item.ID, "error", operationErr)
			}
		}()
	}
	wg.Wait()
	return ctx.Err()
}

func (s *Service) processProvision(parent context.Context, id string) error {
	now := s.now().UTC()
	claimed, err := s.store.Update(parent, id, func(item *environment.Environment) error {
		if item.Status != environment.StatusRequested || item.DesiredState != environment.DesiredReady || item.NextAttemptAt.After(now) {
			return errNotCandidate
		}
		if err := item.SetStatus(environment.StatusPlanning, "Validating policy and preparing Terraform state", now); err != nil {
			return err
		}
		item.ProvisionAttempts++
		return nil
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, s.config.OperationTimeout)
	defer cancel()
	started := time.Now()
	reporter := func(status environment.Status, message string) error {
		_, err := s.store.Update(ctx, id, func(item *environment.Environment) error {
			return item.SetStatus(status, message, s.now().UTC())
		})
		return err
	}
	result, err := s.provider.Provision(ctx, claimed, reporter)
	if err != nil {
		s.counters.provisionFailed.Add(1)
		_, updateErr := s.store.Update(parent, id, func(item *environment.Environment) error {
			now := s.now().UTC()
			if transitionErr := item.SetStatus(environment.StatusFailed, "Provisioning failed", now); transitionErr != nil {
				return transitionErr
			}
			item.LastError = truncate(err.Error(), 4000)
			return nil
		})
		if updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	_, err = s.store.Update(parent, id, func(item *environment.Environment) error {
		now := s.now().UTC()
		if err := item.SetStatus(environment.StatusReady, "Preview environment is ready", now); err != nil {
			return err
		}
		item.PreviewURL = result.PreviewURL
		item.CloudResources = result.CloudResources
		item.ReadyAt = &now
		item.LastHeartbeat = &now
		item.LastError = ""
		return nil
	})
	if err == nil {
		s.counters.provisionSucceeded.Add(1)
		s.logger.Info("environment ready", "environment_id", id, "duration", time.Since(started))
	}
	return err
}

func (s *Service) processDestroy(parent context.Context, id string) error {
	now := s.now().UTC()
	claimed, err := s.store.Update(parent, id, func(item *environment.Environment) error {
		if item.DesiredState != environment.DesiredDestroyed || item.Status == environment.StatusDestroyed || item.Status == environment.StatusDestroying || item.NextAttemptAt.After(now) {
			return errNotCandidate
		}
		if err := item.SetStatus(environment.StatusDestroying, "Destroying cloud resources and isolated data", now); err != nil {
			return err
		}
		item.DestroyAttempts++
		return nil
	})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(parent, s.config.OperationTimeout)
	defer cancel()
	if err := s.provider.Destroy(ctx, claimed, func(status environment.Status, message string) error { return nil }); err != nil {
		s.counters.destroyFailed.Add(1)
		_, updateErr := s.store.Update(parent, id, func(item *environment.Environment) error {
			now := s.now().UTC()
			if transitionErr := item.SetStatus(environment.StatusDestroyFailed, "Cleanup failed and will be retried", now); transitionErr != nil {
				return transitionErr
			}
			item.LastError = truncate(err.Error(), 4000)
			item.NextAttemptAt = now.Add(retryBackoff(item.DestroyAttempts))
			return nil
		})
		if updateErr != nil {
			return errors.Join(err, updateErr)
		}
		return err
	}
	_, err = s.store.Update(parent, id, func(item *environment.Environment) error {
		now := s.now().UTC()
		if err := item.SetStatus(environment.StatusDestroyed, "Environment and associated resources destroyed", now); err != nil {
			return err
		}
		item.DestroyedAt = &now
		item.LastError = ""
		return nil
	})
	if err == nil {
		s.counters.destroySucceeded.Add(1)
	}
	return err
}

func (s *Service) signalWorker() {
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Service) Metrics(ctx context.Context) (string, error) {
	items, err := s.store.List(ctx, store.Filter{})
	if err != nil {
		return "", err
	}
	counts := make(map[environment.Status]int)
	active := 0
	estimatedCost := 0.0
	for _, item := range items {
		counts[item.Status]++
		if item.Status != environment.StatusDestroyed && item.Status != environment.StatusFailedCleaned {
			active++
			estimatedCost += item.EstimatedHourlyCost
		}
	}
	statuses := make([]string, 0, len(counts))
	for status := range counts {
		statuses = append(statuses, string(status))
	}
	sort.Strings(statuses)
	var b strings.Builder
	fmt.Fprintf(&b, "# HELP ephemeral_environments_active Current non-destroyed environments.\n# TYPE ephemeral_environments_active gauge\nephemeral_environments_active %d\n", active)
	fmt.Fprintf(&b, "# HELP ephemeral_environment_estimated_hourly_cost Estimated hourly cost of active environments.\n# TYPE ephemeral_environment_estimated_hourly_cost gauge\nephemeral_environment_estimated_hourly_cost %.4f\n", estimatedCost)
	for _, status := range statuses {
		fmt.Fprintf(&b, "ephemeral_environments_status{status=%q} %d\n", status, counts[environment.Status(status)])
	}
	fmt.Fprintf(&b, "ephemeral_environments_created_total %d\n", s.counters.created.Load())
	fmt.Fprintf(&b, "ephemeral_idempotent_requests_total %d\n", s.counters.idempotentHits.Load())
	fmt.Fprintf(&b, "ephemeral_provision_succeeded_total %d\n", s.counters.provisionSucceeded.Load())
	fmt.Fprintf(&b, "ephemeral_provision_failed_total %d\n", s.counters.provisionFailed.Load())
	fmt.Fprintf(&b, "ephemeral_destroy_succeeded_total %d\n", s.counters.destroySucceeded.Load())
	fmt.Fprintf(&b, "ephemeral_destroy_failed_total %d\n", s.counters.destroyFailed.Load())
	return b.String(), nil
}

func environmentID(repository string, pullRequest int, commitSHA string) (string, error) {
	repoName := repository[strings.LastIndex(repository, "/")+1:]
	repoName = sanitizeSlug(repoName)
	sha := commitSHA
	if len(sha) > 7 {
		sha = sha[:7]
	}
	base := fmt.Sprintf("%s-pr-%d-%s", repoName, pullRequest, sha)
	if len(base) > 56 {
		base = base[:56]
	}
	random := make([]byte, 2)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate environment id: %w", err)
	}
	return strings.Trim(base, "-") + "-" + hex.EncodeToString(random), nil
}

func sanitizeSlug(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	lastDash := false
	for _, character := range value {
		valid := character >= 'a' && character <= 'z' || character >= '0' && character <= '9'
		if valid {
			b.WriteRune(character)
			lastDash = false
		} else if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func routingPriority(key string) int {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(key))
	return 1000 + int(hash.Sum32()%48000)
}

func estimateHourlyCost(resources environment.Resources) float64 {
	// Deliberately labeled as an estimate: tune these coefficients for the target AWS region.
	compute := (float64(resources.CPU) / 1024) * 0.04048
	memory := (float64(resources.Memory) / 1024) * 0.004445
	return compute + memory
}

func retryBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 6 {
		attempt = 6
	}
	return time.Duration(1<<attempt) * time.Second
}

func isInFlight(status environment.Status) bool {
	return status == environment.StatusPlanning || status == environment.StatusProvisioning || status == environment.StatusHealthCheck || status == environment.StatusDestroying
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
