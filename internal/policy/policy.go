package policy

import (
	"fmt"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
)

type Policy struct {
	DefaultTTL             time.Duration
	MaxTTL                 time.Duration
	MaxCPU                 int
	MaxMemory              int
	MaxEnvironmentsPerUser int
	MaxEnvironmentsPerRepo int
}

func (p Policy) ValidateRequest(ttl time.Duration, resources environment.Resources) error {
	if ttl <= 0 {
		return fmt.Errorf("ttl must be positive")
	}
	if ttl > p.MaxTTL {
		return fmt.Errorf("ttl exceeds the maximum of %s", p.MaxTTL)
	}
	if resources.CPU <= 0 || resources.CPU > p.MaxCPU {
		return fmt.Errorf("cpu must be between 1 and %d units", p.MaxCPU)
	}
	if resources.Memory <= 0 || resources.Memory > p.MaxMemory {
		return fmt.Errorf("memory must be between 1 and %d MiB", p.MaxMemory)
	}
	if resources.DatabaseMode != environment.DatabaseSharedSchema && resources.DatabaseMode != environment.DatabaseNone {
		return fmt.Errorf("database_mode must be %q or %q", environment.DatabaseSharedSchema, environment.DatabaseNone)
	}
	return nil
}

func (p Policy) ValidateCapacity(existing []*environment.Environment, owner, repository string) error {
	ownerCount, repoCount := 0, 0
	for _, item := range existing {
		if item.Status == environment.StatusDestroyed || item.Status == environment.StatusFailedCleaned {
			continue
		}
		if item.Owner == owner {
			ownerCount++
		}
		if item.Repository == repository {
			repoCount++
		}
	}
	if ownerCount >= p.MaxEnvironmentsPerUser {
		return fmt.Errorf("owner %q already has the maximum of %d active environments", owner, p.MaxEnvironmentsPerUser)
	}
	if repoCount >= p.MaxEnvironmentsPerRepo {
		return fmt.Errorf("repository %q already has the maximum of %d active environments", repository, p.MaxEnvironmentsPerRepo)
	}
	return nil
}
