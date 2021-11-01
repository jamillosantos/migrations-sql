package migrationsql

import (
	"context"
	"database/sql"
	"io/fs"

	"github.com/jamillosantos/migrations"
)

type Driver interface {
	DB() *sql.DB

	Lock() error
	Unlock() error

	DBExecer
	Conn() (*sql.Conn, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)

	CreateMigrationsTable() error
	DropMigrationsTable() error
}

type opts struct {
	table    string
	reporter migrations.RunnerReporter
}

type MigrateOption func(*opts)

func defaultOpts() opts {
	return opts{}
}

func Migrate(ctx context.Context, dirFS fs.ReadDirFS, driver Driver, opts ...MigrateOption) error {
	source, err := NewSourceSQLFromDir(dirFS)
	if err != nil {
		return err
	}

	o := defaultOpts()
	for _, optionModifier := range opts {
		optionModifier(&o)
	}

	targetOpts := make([]Option, 0)
	if o.table != "" {
		targetOpts = append(targetOpts, Table(o.table))
	}

	target, err := NewTarget(source, driver, targetOpts...)
	if err != nil {
		return err
	}

	conn, err := driver.Conn()
	if err != nil {
		return err
	}

	ctx = ContextWithConn(ctx, conn)

	runner := migrations.NewRunner(source, target)
	_, err = migrations.Migrate(ctx, runner, o.reporter)
	if err != nil {
		return err
	}

	return nil
}
