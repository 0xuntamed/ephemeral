package store

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed schema.sql
var postgresSchema string

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(12)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(30 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := db.ExecContext(ctx, postgresSchema); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate postgres: %w", err)
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Create(ctx context.Context, candidate *environment.Environment) (*environment.Environment, bool, error) {
	if err := candidate.Validate(); err != nil {
		return nil, false, err
	}
	document, err := json.Marshal(candidate)
	if err != nil {
		return nil, false, err
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO environments (
			id, idempotency_key, status, desired_state, repository, owner,
			pull_request, expires_at, created_at, updated_at, document
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (idempotency_key) DO NOTHING`,
		candidate.ID, candidate.IdempotencyKey, candidate.Status, candidate.DesiredState,
		candidate.Repository, candidate.Owner, candidate.PullRequest, candidate.ExpiresAt,
		candidate.CreatedAt, candidate.UpdatedAt, document,
	)
	if err != nil {
		return nil, false, fmt.Errorf("insert environment: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	if rows == 1 {
		return candidate.Clone(), true, nil
	}
	existing, err := s.getByIdempotencyKey(ctx, candidate.IdempotencyKey)
	return existing, false, err
}

func (s *PostgresStore) getByIdempotencyKey(ctx context.Context, key string) (*environment.Environment, error) {
	var document []byte
	err := s.db.QueryRowContext(ctx, `SELECT document FROM environments WHERE idempotency_key=$1`, key).Scan(&document)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return decodeEnvironment(document)
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*environment.Environment, error) {
	var document []byte
	err := s.db.QueryRowContext(ctx, `SELECT document FROM environments WHERE id=$1`, id).Scan(&document)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get environment: %w", err)
	}
	return decodeEnvironment(document)
}

func (s *PostgresStore) List(ctx context.Context, filter Filter) ([]*environment.Environment, error) {
	query := `SELECT document FROM environments WHERE 1=1`
	args := make([]any, 0, 3)
	if filter.Status != "" {
		args = append(args, filter.Status)
		query += fmt.Sprintf(" AND status=$%d", len(args))
	}
	if filter.Repository != "" {
		args = append(args, filter.Repository)
		query += fmt.Sprintf(" AND lower(repository)=lower($%d)", len(args))
	}
	if filter.Owner != "" {
		args = append(args, filter.Owner)
		query += fmt.Sprintf(" AND lower(owner)=lower($%d)", len(args))
	}
	query += ` ORDER BY created_at DESC`
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	defer rows.Close()
	items := make([]*environment.Environment, 0)
	for rows.Next() {
		var document []byte
		if err := rows.Scan(&document); err != nil {
			return nil, err
		}
		item, err := decodeEnvironment(document)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) Update(ctx context.Context, id string, mutate func(*environment.Environment) error) (*environment.Environment, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var document []byte
	if err := tx.QueryRowContext(ctx, `SELECT document FROM environments WHERE id=$1 FOR UPDATE`, id).Scan(&document); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	item, err := decodeEnvironment(document)
	if err != nil {
		return nil, err
	}
	if err := mutate(item); err != nil {
		return nil, err
	}
	if err := item.Validate(); err != nil {
		return nil, err
	}
	document, err = json.Marshal(item)
	if err != nil {
		return nil, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE environments SET status=$2, desired_state=$3, repository=$4, owner=$5,
		pull_request=$6, expires_at=$7, updated_at=$8, document=$9 WHERE id=$1`,
		item.ID, item.Status, item.DesiredState, item.Repository, item.Owner,
		item.PullRequest, item.ExpiresAt, item.UpdatedAt, document,
	)
	if err != nil {
		return nil, fmt.Errorf("update environment: %w", err)
	}
	if rows, _ := result.RowsAffected(); rows != 1 {
		return nil, ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return item.Clone(), nil
}

func decodeEnvironment(document []byte) (*environment.Environment, error) {
	var item environment.Environment
	if err := json.Unmarshal(document, &item); err != nil {
		return nil, fmt.Errorf("decode environment: %w", err)
	}
	return &item, nil
}

func (s *PostgresStore) Close() error { return s.db.Close() }

func New(ctx context.Context, databaseURL, filePath string) (Store, error) {
	if strings.TrimSpace(databaseURL) != "" {
		return NewPostgresStore(ctx, databaseURL)
	}
	return NewFileStore(filePath)
}
