package migrationsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"strings"

	"github.com/jamillosantos/migrations"

	"github.com/jamillosantos/migrations-sql/target"
)

var (
	ErrInvalidMigrationDirection = errors.New("invalid migration direction")

	migrationFileNameRegexp = regexp.MustCompile(`^(\d+)_(.*?)(\.(do|undo|down|up))?\.sql$`)
)

type DBExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

// parseSQLFile checks if the entry is valid and returns its id, description and type.
func parseSQLFile(entry fs.DirEntry) (id, description, t string) {
	if entry.IsDir() {
		return
	}
	m := migrationFileNameRegexp.FindStringSubmatch(entry.Name())
	if len(m) == 0 {
		return
	}
	id, description, t = m[1], strings.ReplaceAll(m[2], "_", " "), m[4]
	return
}

type migration struct {
	description string
	doFile      string
	undoFile    string
}

func NewSourceSQLFromDir(fs fs.ReadDirFS, folder string, opts ...Option) (migrations.Source, error) {
	options := defaultOptions()
	entries, err := fs.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("failed listing migrations files: %w", err)
	}
	migrationSet := make(map[string]*migration)
	for _, entry := range entries {
		id, description, t := parseSQLFile(entry)
		if id == "" { // Does not match
			continue
		}

		migrationEntry := migrationSet[id]
		if migrationEntry == nil {
			migrationEntry = &migration{
				description: description,
			}
			migrationSet[id] = migrationEntry
		}

		switch t {
		case "", "up", "do":
			if migrationEntry.doFile != "" {
				// TODO: Improve this error
				return nil, fmt.Errorf("migration %s already defined by %s", entry.Name(), migrationEntry.doFile)
			}
			migrationEntry.doFile = path.Join(folder, entry.Name())
		case "down", "undo":
			if migrationEntry.undoFile != "" {
				// TODO: Improve this error
				return nil, fmt.Errorf("migration %s already defined by %s", entry.Name(), migrationEntry.doFile)
			}
			migrationEntry.undoFile = path.Join(folder, entry.Name())
		default:
			return nil, fmt.Errorf("%w: %s (%s)", ErrInvalidMigrationDirection, t, entry.Name())
		}
	}
	if options.source == nil {
		options.source = migrations.NewSource()
	}
	return newSourceSQLFromFiles(options.source, fs, migrationSet)
}

func defaultOptions() options {
	return options{}
}

type SourceWithAdd interface {
	migrations.Source
	Add(migration migrations.Migration) error
}

func newSourceSQLFromFiles(source SourceWithAdd, fs fs.ReadDirFS, files map[string]*migration) (migrations.Source, error) {
	for migrationID, migration := range files {
		m, err := source.ByID(migrationID)
		if errors.Is(err, migrations.ErrMigrationNotFound) {
			// If the migration was not added yet, create the instance and add it.
			m = &migrationSQL{
				id:          migrationID,
				description: migration.description,
			}
		} else if err != nil {
			return nil, err
		}

		mSQL := m.(*migrationSQL)

		mSQL.doFile = migration.doFile
		mSQL.doFileContent, err = loadMigrationFile(fs, migration.doFile)
		if err != nil {
			return nil, err
		}
		mSQL.undoFile = migration.undoFile
		mSQL.undoFileContent, err = loadMigrationFile(fs, migration.undoFile)
		if err != nil {
			return nil, err
		}
		err = source.Add(m)
		if err != nil {
			return nil, err
		}
	}
	return source, nil
}

func loadMigrationFile(fs fs.ReadDirFS, file string) (string, error) {
	if file == "" {
		// does not have migration
		return "", nil
	}

	f, err := fs.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	var buf strings.Builder
	_, err = io.Copy(&buf, f)
	if err != nil {
		return "", fmt.Errorf("cannot read migration file: %s: %w", file, err)
	}
	return buf.String(), nil
}

type migrationSQL struct {
	id              string
	description     string
	next            migrations.Migration
	previous        migrations.Migration
	doFile          string
	doFileContent   string
	undoFile        string
	undoFileContent string
}

// ID identifies the migration. Through the ID, all the sorting is done.
func (migration *migrationSQL) ID() string {
	return migration.id
}

// String will return a representation of the migration into a string format
// for user identification.
func (migration *migrationSQL) String() string {
	if migration.CanUndo() {
		return fmt.Sprintf("[%s,%s]", migration.doFile, migration.undoFile)
	}
	return fmt.Sprintf("[%s]", migration.doFile)
}

// Description is the humanized description for the migration.
func (migration *migrationSQL) Description() string {
	return migration.description
}

// Next will link this migration with the next. This link should be created
// by the source while it is being loaded.
func (migration *migrationSQL) Next() migrations.Migration {
	return migration.next
}

// SetNext will set the next migration
func (migration *migrationSQL) SetNext(value migrations.Migration) migrations.Migration {
	migration.next = value
	return migration
}

// Previous will link this migration with the previous. This link should be
// created by the Source while it is being loaded.
func (migration *migrationSQL) Previous() migrations.Migration {
	return migration.previous
}

// SetPrevious will set the previous migration
func (migration *migrationSQL) SetPrevious(value migrations.Migration) migrations.Migration {
	migration.previous = value
	return migration
}

func (migration *migrationSQL) executeSQL(ctx context.Context, sql string) error {
	db, err := target.DBFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = db.Exec(sql)
	if err != nil {
		return migrations.NewQueryError(err, sql)
	}
	return nil
}

// Do will execute the migration.
func (migration *migrationSQL) Do(ctx context.Context) error {
	return migration.executeSQL(ctx, migration.doFileContent)
}

// CanUndo is a flag that mark this flag as undoable.
func (migration *migrationSQL) CanUndo() bool {
	return migration.undoFile != ""
}

// Undo will undo the migration.
func (migration *migrationSQL) Undo(ctx context.Context) error {
	return migration.executeSQL(ctx, migration.undoFileContent)
}
