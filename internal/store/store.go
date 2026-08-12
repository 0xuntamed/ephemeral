package store

import (
	"context"
	"errors"

	"github.com/example/ephemeral-platform/internal/environment"
)

var (
	ErrNotFound = errors.New("environment not found")
	ErrConflict = errors.New("environment update conflict")
)

type Filter struct {
	Status     environment.Status
	Repository string
	Owner      string
}

type Store interface {
	Create(ctx context.Context, candidate *environment.Environment) (saved *environment.Environment, created bool, err error)
	Get(ctx context.Context, id string) (*environment.Environment, error)
	List(ctx context.Context, filter Filter) ([]*environment.Environment, error)
	Update(ctx context.Context, id string, mutate func(*environment.Environment) error) (*environment.Environment, error)
	Close() error
}
