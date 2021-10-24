//go:build integration
// +build integration

package migrationsql_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jamillosantos/migrations"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	migrationszap "github.com/jamillosantos/migrations/zap"

	migrationsql "github.com/jamillosantos/migrations-sql"
)

const (
	base  = "tests/migrations"
	case1 = base + "/case1"
	case2 = base + "/case2"
	case3 = base + "/case3"
)

func migrate(t *testing.T, db *sql.DB, migrationCase string) error {
	t.Helper()

	dirFS := afero.NewIOFS(afero.NewBasePathFs(afero.NewOsFs(), migrationCase))

	ctx := context.Background()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	return migrationsql.Migrate(ctx, dirFS, db, migrationszap.NewRunnerReport(logger))
}

func TestMigrate_Integration(t *testing.T) {
	t.Run("should run case 1", func(t *testing.T) {
		db := createDBConnection(t)
		err := migrate(t, db, case1)
		require.NoError(t, err)
	})

	t.Run("should run case 1 + case 2", func(t *testing.T) {
		db := createDBConnection(t)
		err := migrate(t, db, case1)
		require.NoError(t, err)
		err = migrate(t, db, case2)
		require.NoError(t, err)
	})

	t.Run("should fail running the case 3 after case 1 and case 2", func(t *testing.T) {
		db := createDBConnection(t)
		err := migrate(t, db, case1)
		require.NoError(t, err)
		err = migrate(t, db, case2)
		require.NoError(t, err)
		err = migrate(t, db, case3)
		require.ErrorIs(t, err, migrations.ErrStaleMigrationDetected)
	})
}

func createDBConnection(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Error(err)
		t.FailNow()
		return nil
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}
