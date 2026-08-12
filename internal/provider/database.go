package provider

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"

	"github.com/example/ephemeral-platform/internal/environment"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var safeSchemaName = regexp.MustCompile(`^[a-z][a-z0-9_]{2,62}$`)

type databaseSchemas struct {
	provider    Provider
	databaseURL string
}

func WithDatabaseSchemas(infrastructure Provider, databaseURL string) Provider {
	return &databaseSchemas{provider: infrastructure, databaseURL: databaseURL}
}

func (d *databaseSchemas) Name() string { return d.provider.Name() + "+postgres-schema" }

func (d *databaseSchemas) Provision(ctx context.Context, item *environment.Environment, report Reporter) (ProvisionResult, error) {
	if item.Resources.DatabaseMode == environment.DatabaseSharedSchema {
		if err := d.create(ctx, item.DatabaseSchema); err != nil {
			return ProvisionResult{}, fmt.Errorf("create preview database schema: %w", err)
		}
	}
	result, err := d.provider.Provision(ctx, item, report)
	if err != nil && item.Resources.DatabaseMode == environment.DatabaseSharedSchema {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = d.drop(cleanupCtx, item.DatabaseSchema)
	}
	return result, err
}

func (d *databaseSchemas) Destroy(ctx context.Context, item *environment.Environment, report Reporter) error {
	if err := d.provider.Destroy(ctx, item, report); err != nil {
		return err
	}
	if item.Resources.DatabaseMode == environment.DatabaseSharedSchema {
		if err := d.drop(ctx, item.DatabaseSchema); err != nil {
			return fmt.Errorf("drop preview database schema: %w", err)
		}
	}
	return nil
}

func (d *databaseSchemas) create(ctx context.Context, schema string) error {
	if !safeSchemaName.MatchString(schema) {
		return fmt.Errorf("unsafe schema name %q", schema)
	}
	return d.execute(ctx, `CREATE SCHEMA IF NOT EXISTS "`+schema+`"`)
}

func (d *databaseSchemas) drop(ctx context.Context, schema string) error {
	if !safeSchemaName.MatchString(schema) {
		return fmt.Errorf("unsafe schema name %q", schema)
	}
	return d.execute(ctx, `DROP SCHEMA IF EXISTS "`+schema+`" CASCADE`)
}

func (d *databaseSchemas) execute(ctx context.Context, statement string) error {
	db, err := sql.Open("pgx", d.databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.ExecContext(ctx, statement)
	return err
}
