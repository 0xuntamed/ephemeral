package environment

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

type Status string

const (
	StatusRequested     Status = "REQUESTED"
	StatusPlanning      Status = "PLANNING"
	StatusProvisioning  Status = "PROVISIONING"
	StatusHealthCheck   Status = "HEALTH_CHECK"
	StatusReady         Status = "READY"
	StatusDestroying    Status = "DESTROYING"
	StatusDestroyed     Status = "DESTROYED"
	StatusFailed        Status = "FAILED"
	StatusDestroyFailed Status = "DESTROY_FAILED"
	StatusOrphaned      Status = "ORPHANED"
	StatusFailedCleaned Status = "FAILED_CLEANED"
)

type DesiredState string

const (
	DesiredReady     DesiredState = "READY"
	DesiredDestroyed DesiredState = "DESTROYED"
)

type DatabaseMode string

const (
	DatabaseSharedSchema DatabaseMode = "shared-schema"
	DatabaseNone         DatabaseMode = "none"
)

type Resources struct {
	CPU          int          `json:"cpu"`
	Memory       int          `json:"memory"`
	DatabaseMode DatabaseMode `json:"database_mode"`
}

type CloudResource struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type Event struct {
	At      time.Time `json:"at"`
	Status  Status    `json:"status"`
	Message string    `json:"message"`
}

type Environment struct {
	ID                  string          `json:"id"`
	IdempotencyKey      string          `json:"idempotency_key"`
	Repository          string          `json:"repository"`
	Branch              string          `json:"branch"`
	CommitSHA           string          `json:"commit_sha"`
	PullRequest         int             `json:"pull_request"`
	Owner               string          `json:"owner"`
	Image               string          `json:"image"`
	Status              Status          `json:"status"`
	DesiredState        DesiredState    `json:"desired_state"`
	StatusMessage       string          `json:"status_message"`
	TerraformWorkspace  string          `json:"terraform_workspace"`
	PreviewURL          string          `json:"preview_url,omitempty"`
	DatabaseSchema      string          `json:"database_schema,omitempty"`
	RoutingPriority     int             `json:"routing_priority"`
	Resources           Resources       `json:"resources"`
	CloudResources      []CloudResource `json:"cloud_resources"`
	EstimatedHourlyCost float64         `json:"estimated_hourly_cost"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
	ExpiresAt           time.Time       `json:"expires_at"`
	ReadyAt             *time.Time      `json:"ready_at,omitempty"`
	DestroyedAt         *time.Time      `json:"destroyed_at,omitempty"`
	LastHeartbeat       *time.Time      `json:"last_heartbeat,omitempty"`
	NextAttemptAt       time.Time       `json:"next_attempt_at"`
	ProvisionAttempts   int             `json:"provision_attempts"`
	DestroyAttempts     int             `json:"destroy_attempts"`
	LastError           string          `json:"last_error,omitempty"`
	Events              []Event         `json:"events"`
}

func (e *Environment) Clone() *Environment {
	if e == nil {
		return nil
	}
	clone := *e
	clone.CloudResources = slices.Clone(e.CloudResources)
	clone.Events = slices.Clone(e.Events)
	return &clone
}

func (e *Environment) AddEvent(status Status, message string, now time.Time) {
	e.Events = append(e.Events, Event{At: now.UTC(), Status: status, Message: strings.TrimSpace(message)})
	if len(e.Events) > 100 {
		e.Events = slices.Clone(e.Events[len(e.Events)-100:])
	}
}

func (e *Environment) SetStatus(next Status, message string, now time.Time) error {
	if e.Status != next && !CanTransition(e.Status, next) {
		return fmt.Errorf("invalid lifecycle transition %s -> %s", e.Status, next)
	}
	e.Status = next
	e.StatusMessage = strings.TrimSpace(message)
	e.UpdatedAt = now.UTC()
	e.AddEvent(next, message, now)
	return nil
}

func CanTransition(from, to Status) bool {
	if from == to {
		return true
	}
	allowed := map[Status][]Status{
		StatusRequested:     {StatusPlanning, StatusDestroying, StatusFailed},
		StatusPlanning:      {StatusProvisioning, StatusFailed, StatusDestroying, StatusOrphaned},
		StatusProvisioning:  {StatusHealthCheck, StatusFailed, StatusDestroying, StatusOrphaned},
		StatusHealthCheck:   {StatusReady, StatusFailed, StatusDestroying, StatusOrphaned},
		StatusReady:         {StatusDestroying, StatusOrphaned},
		StatusFailed:        {StatusRequested, StatusDestroying, StatusFailedCleaned},
		StatusOrphaned:      {StatusDestroying, StatusFailedCleaned},
		StatusDestroyFailed: {StatusDestroying, StatusOrphaned},
		StatusDestroying:    {StatusDestroyed, StatusDestroyFailed, StatusOrphaned, StatusFailedCleaned},
		StatusFailedCleaned: {StatusDestroyed},
	}
	return slices.Contains(allowed[from], to)
}

func (e *Environment) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("environment id is required")
	}
	if strings.TrimSpace(e.IdempotencyKey) == "" {
		return errors.New("idempotency key is required")
	}
	if e.CreatedAt.IsZero() || e.ExpiresAt.IsZero() {
		return errors.New("created_at and expires_at are required")
	}
	return nil
}
