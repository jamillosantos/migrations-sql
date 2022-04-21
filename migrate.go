package migrationsql

import (
	"context"
	"io/fs"

	"github.com/jamillosantos/migrations"
)

func Migrate(ctx context.Context, dirFS fs.ReadDirFS, folder string, db DB, reporter migrations.RunnerReporter) error {
	source, err := NewSourceSQLFromDir(dirFS, folder)
	if err != nil {
		return err
	}

	target, err := NewTarget(source, db)
	if err != nil {
		return err
	}

	ctx = ContextWithDB(ctx, db)

	runner := migrations.NewRunner(source, target)
	_, err = migrations.Migrate(ctx, runner, reporter)
	if err != nil {
		return err
	}

	return nil
}
