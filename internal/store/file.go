package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/example/ephemeral-platform/internal/environment"
)

type FileStore struct {
	mu    sync.RWMutex
	path  string
	items map[string]*environment.Environment
	byKey map[string]string
}

type fileSnapshot struct {
	Environments []*environment.Environment `json:"environments"`
}

func NewFileStore(path string) (*FileStore, error) {
	s := &FileStore{
		path:  strings.TrimSpace(path),
		items: make(map[string]*environment.Environment),
		byKey: make(map[string]string),
	}
	if s.path == "" {
		return s, nil
	}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileStore) load() error {
	contents, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read store: %w", err)
	}
	var snapshot fileSnapshot
	if err := json.Unmarshal(contents, &snapshot); err != nil {
		return fmt.Errorf("decode store: %w", err)
	}
	for _, item := range snapshot.Environments {
		if item == nil {
			continue
		}
		s.items[item.ID] = item.Clone()
		s.byKey[item.IdempotencyKey] = item.ID
	}
	return nil
}

func (s *FileStore) persistLocked() error {
	if s.path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o750); err != nil {
		return fmt.Errorf("create store directory: %w", err)
	}
	items := make([]*environment.Environment, 0, len(s.items))
	for _, item := range s.items {
		items = append(items, item.Clone())
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	contents, err := json.MarshalIndent(fileSnapshot{Environments: items}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode store: %w", err)
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, contents, 0o600); err != nil {
		return fmt.Errorf("write store: %w", err)
	}
	if err := replaceFile(tmp, s.path); err != nil {
		return fmt.Errorf("replace store: %w", err)
	}
	return nil
}

func replaceFile(tmp, destination string) error {
	if err := os.Rename(tmp, destination); err == nil {
		return nil
	}
	contents, err := os.ReadFile(tmp)
	if err != nil {
		return err
	}
	if err := os.WriteFile(destination, contents, 0o600); err != nil {
		return err
	}
	return os.Remove(tmp)
}

func (s *FileStore) Create(_ context.Context, candidate *environment.Environment) (*environment.Environment, bool, error) {
	if err := candidate.Validate(); err != nil {
		return nil, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if id, ok := s.byKey[candidate.IdempotencyKey]; ok {
		return s.items[id].Clone(), false, nil
	}
	if _, ok := s.items[candidate.ID]; ok {
		return nil, false, ErrConflict
	}
	s.items[candidate.ID] = candidate.Clone()
	s.byKey[candidate.IdempotencyKey] = candidate.ID
	if err := s.persistLocked(); err != nil {
		delete(s.items, candidate.ID)
		delete(s.byKey, candidate.IdempotencyKey)
		return nil, false, err
	}
	return candidate.Clone(), true, nil
}

func (s *FileStore) Get(_ context.Context, id string) (*environment.Environment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	return item.Clone(), nil
}

func (s *FileStore) List(_ context.Context, filter Filter) ([]*environment.Environment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]*environment.Environment, 0, len(s.items))
	for _, item := range s.items {
		if filter.Status != "" && item.Status != filter.Status {
			continue
		}
		if filter.Repository != "" && !strings.EqualFold(item.Repository, filter.Repository) {
			continue
		}
		if filter.Owner != "" && !strings.EqualFold(item.Owner, filter.Owner) {
			continue
		}
		items = append(items, item.Clone())
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	return items, nil
}

func (s *FileStore) Update(_ context.Context, id string, mutate func(*environment.Environment) error) (*environment.Environment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	updated := current.Clone()
	if err := mutate(updated); err != nil {
		return nil, err
	}
	if err := updated.Validate(); err != nil {
		return nil, err
	}
	s.items[id] = updated
	if err := s.persistLocked(); err != nil {
		s.items[id] = current
		return nil, err
	}
	return updated.Clone(), nil
}

func (s *FileStore) Close() error { return nil }
