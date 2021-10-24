//go:generate go run github.com/golang/mock/mockgen -package migrationsql -destination migration_mock_test.go github.com/jamillosantos/migrations Source,Migration

package migrationsql

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamillosantos/migrations"
)

const (
	sqlListTables = "SELECT name FROM sqlite_master"
)

func newMockMigration(ctrl *gomock.Controller, id string) *MockMigration {
	m := NewMockMigration(ctrl)
	m.EXPECT().ID().Return(id).AnyTimes()
	m.EXPECT().String().Return(fmt.Sprintf("migration %s", id)).AnyTimes()
	return m
}

func createDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func TestNewTarget(t *testing.T) {
	t.Run("should create the target", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		source := NewMockSource(ctrl)
		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.Equal(t, db, target.db)
		assert.Equal(t, source, target.source)
	})

	t.Run("should fail when a option fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		wantErr := errors.New("random error")

		source := NewMockSource(ctrl)
		_, err := NewTarget(source, db, func(target *Target) error {
			return wantErr
		})
		assert.ErrorIs(t, err, wantErr)
	})
}

func Test_targetSQL_Create(t *testing.T) {
	t.Run("should create the _migrations table", func(t *testing.T) {
		db := createDB(t)

		// Prepares scenario
		target, err := NewTarget(nil, db)
		assert.NoError(t, err)

		// Creates table
		require.NoError(t, target.Create())

		// Checks if the table exists
		var tableName string
		assert.NoError(t, db.QueryRow(sqlListTables).Scan(&tableName))
		assert.Equal(t, tableName, "_migrations")
	})

	t.Run("should create the _migrations table", func(t *testing.T) {
		db := createDB(t)

		// Prepares scenario
		wantTableName := "new_migration_table"
		target, err := NewTarget(nil, db, Table(wantTableName))
		assert.NoError(t, err)

		// Creates table
		require.NoError(t, target.Create())

		// Checks if the table exists
		var tableName string
		assert.NoError(t, db.QueryRow(sqlListTables).Scan(&tableName))
		assert.Equal(t, tableName, wantTableName)
	})
}

func Test_targetSQL_Destroy(t *testing.T) {
	db := createDB(t)

	// Prepares scenario
	target, err := NewTarget(nil, db)
	require.NoError(t, err)
	err = target.Create()
	require.NoError(t, err)

	// Destroys table
	err = target.Destroy()
	require.NoError(t, err)

	// Checks if the table exists.
	var tableName string
	err = db.QueryRow(sqlListTables).Scan(&tableName)
	assert.Error(t, err)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func Test_targetSQL_Add(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := createDB(t)

	// Prepares scenarios
	target, err := NewTarget(nil, db)
	require.NoError(t, err)
	require.NoError(t, target.Create())

	// Creates and adds migrations
	m1 := newMockMigration(ctrl, "1")
	m2 := newMockMigration(ctrl, "2")
	m3 := newMockMigration(ctrl, "3")

	assert.NoError(t, target.Add(m3))
	assert.NoError(t, target.Add(m1))
	assert.NoError(t, target.Add(m2))

	// Check the database for the tables
	rs, err := db.Query("SELECT id FROM _migrations ORDER BY id;")
	require.NoError(t, err)
	defer func() {
		_ = rs.Close()
	}()

	var id string

	// finds m1
	assert.True(t, rs.Next())
	assert.NoError(t, rs.Scan(&id))
	assert.Equal(t, m1.ID(), id)

	// finds m2
	assert.True(t, rs.Next())
	assert.NoError(t, rs.Scan(&id))
	assert.Equal(t, m2.ID(), id)

	// finds m3
	assert.True(t, rs.Next())
	assert.NoError(t, rs.Scan(&id))
	assert.Equal(t, m3.ID(), id)

	// EOF
	assert.False(t, rs.Next())
}

func Test_targetSQL_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := createDB(t)

	// Prepare scenario
	m1 := newMockMigration(ctrl, "1")
	m2 := newMockMigration(ctrl, "2")
	m3 := newMockMigration(ctrl, "3")

	target, err := NewTarget(nil, db)
	require.NoError(t, err)

	assert.NoError(t, target.Create())
	assert.NoError(t, target.Add(m3))
	assert.NoError(t, target.Add(m1))
	assert.NoError(t, target.Add(m2))

	// Removes the migration m3
	assert.NoError(t, target.Remove(m3))

	// Check the database for the tables
	rs, err := db.Query("SELECT id FROM _migrations ORDER BY id;")
	assert.NoError(t, err)
	defer func() {
		_ = rs.Close()
	}()

	var gotID string

	// Find m1
	assert.True(t, rs.Next())
	assert.NoError(t, rs.Scan(&gotID))
	assert.Equal(t, m1.ID(), gotID)

	// Find m2
	assert.True(t, rs.Next())
	assert.NoError(t, rs.Scan(&gotID))
	assert.Equal(t, m2.ID(), gotID)

	// EOF
	assert.False(t, rs.Next())
}

func Test_targetSQL_Current(t *testing.T) {
	t.Run("should return the most recent migration", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		// Prepare scenario
		source := NewMockSource(ctrl)

		m1 := newMockMigration(ctrl, "0")
		m2 := newMockMigration(ctrl, "2")

		source.EXPECT().ByID(m1.ID()).Return(m1, nil).AnyTimes()
		source.EXPECT().ByID(m2.ID()).Return(m2, nil).AnyTimes()

		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.NoError(t, target.Create())
		assert.NoError(t, target.Add(m1))
		assert.NoError(t, target.Add(m2))

		// Get the current migration
		currentMigration, err := target.Current()
		assert.NoError(t, err)
		assert.Equal(t, m2, currentMigration)
	})

	t.Run("should fail when there is no migration applied", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		// Prepare scenario
		source := NewMockSource(ctrl)

		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.NoError(t, target.Create())

		// Get the current migration
		_, err = target.Current()
		assert.ErrorIs(t, err, migrations.ErrNoCurrentMigration)
	})

	t.Run("should fail when Done fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		// Prepare scenario
		source := NewMockSource(ctrl)

		m1 := newMockMigration(ctrl, "1")
		wantErr := errors.New("random error")

		source.EXPECT().ByID(gomock.Any()).Return(nil, wantErr)

		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.NoError(t, target.Create())
		assert.NoError(t, target.Add(m1))

		// Get the current migration
		_, err = target.Current()
		assert.ErrorIs(t, err, wantErr)
	})
}

func Test_targetSQL_Done(t *testing.T) {
	t.Run("should list of the executed migrations", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		// Prepare scenario
		source := NewMockSource(ctrl)

		m1 := newMockMigration(ctrl, "1")
		m2 := newMockMigration(ctrl, "2")
		m3 := newMockMigration(ctrl, "3")

		source.EXPECT().ByID(m1.ID()).Return(m1, nil)
		source.EXPECT().ByID(m2.ID()).Return(m2, nil)
		source.EXPECT().ByID(m3.ID()).Return(m3, nil)

		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.NoError(t, target.Create())
		assert.NoError(t, target.Add(m1))
		assert.NoError(t, target.Add(m2))
		assert.NoError(t, target.Add(m3))

		// List executed migrations
		list, err := target.Done()
		assert.NoError(t, err)
		require.Len(t, list, 3)

		assert.Equal(t, m1, list[0])
		assert.Equal(t, m2, list[1])
		assert.Equal(t, m3, list[2])
	})

	t.Run("should fail listing a migration that is not in the source", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		db := createDB(t)

		// Prepare scenario
		source := NewMockSource(ctrl)

		m1 := newMockMigration(ctrl, "1")
		m2 := newMockMigration(ctrl, "2")
		m3 := newMockMigration(ctrl, "3")

		source.EXPECT().ByID(m1.ID()).Return(m1, nil)
		source.EXPECT().ByID(m2.ID()).Return(nil, migrations.ErrMigrationNotFound)
		// source.EXPECT().ByID(m3.ID()).Return(m3, nil)

		target, err := NewTarget(source, db)
		assert.NoError(t, err)

		assert.NoError(t, target.Create())
		assert.NoError(t, target.Add(m1))
		assert.NoError(t, target.Add(m2))
		assert.NoError(t, target.Add(m3))

		// List executed migrations
		_, err = target.Done()
		assert.ErrorIs(t, err, migrations.ErrMigrationNotFound)
	})
}
