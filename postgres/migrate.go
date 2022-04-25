package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jamillosantos/migrations"

	"github.com/jamillosantos/migrations-sql/target"
)

var (
	ErrMissingDatabaseName = errors.New("missing database name")
)

type MigrateRequest struct {
	DB           target.DB
	Source       migrations.Source
	DatabaseName string
	TableName    string
	Reporter     migrations.RunnerReporter
}

func Migrate(ctx context.Context, req MigrateRequest) error {
	if req.TableName == "" {
		req.TableName = target.DefaultMigrationsTableName
	}

	if req.DatabaseName == "" {
		return ErrMissingDatabaseName
	}

	t, err := New(req.Source, req.DB, req.DatabaseName, req.TableName)
	if err != nil {
		return fmt.Errorf("error initializing postgres driver: %w", err)
	}

	runner := migrations.NewRunner(req.Source, t)

	ctx = target.ContextWithDB(ctx, req.DB)

	_, err = migrations.Migrate(ctx, runner, req.Reporter)
	if err != nil {
		return fmt.Errorf("migration process failed: %w", err)
	}

	return nil
}
