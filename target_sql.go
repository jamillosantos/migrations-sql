package migrationsql

import (
	"database/sql"
	"fmt"

	"github.com/jamillosantos/migrations"
)

type DB interface {
	DBExecer
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Target struct {
	source    migrations.Source
	db        DB
	tableName string
}

type Option func(target *Target) error

func NewTarget(source migrations.Source, db DB, options ...Option) (*Target, error) {
	target := &Target{
		source,
		db,
		"_migrations",
	}
	for _, opt := range options {
		err := opt(target)
		if err != nil {
			return nil, err
		}
	}
	return target, nil
}

func Table(tableName string) Option {
	return func(target *Target) error {
		target.tableName = tableName
		return nil
	}
}

func (target *Target) Create() error {
	_, err := target.db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id BIGINT PRIMARY KEY)", target.tableName))
	return err
}

func (target *Target) Destroy() error {
	_, err := target.db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", target.tableName))
	return err
}

func (target *Target) Current() (migrations.Migration, error) {
	list, err := target.Done()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, migrations.ErrNoCurrentMigration
	}
	return list[len(list)-1], nil
}

func (target *Target) Done() ([]migrations.Migration, error) {
	rs, err := target.db.Query(fmt.Sprintf("SELECT id FROM %s ORDER BY id ASC", target.tableName))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rs.Close()
	}()

	var id string
	result := make([]migrations.Migration, 0)
	for rs.Next() {
		err := rs.Scan(&id)
		if err != nil {
			return nil, err
		}
		idDt := id
		migration, err := target.source.ByID(idDt)
		if err != nil {
			return nil, err
		}
		result = append(result, migration)
	}
	return result, nil
}

func (target *Target) Add(migration migrations.Migration) error {
	_, err := target.db.Exec(fmt.Sprintf("INSERT INTO %s (id) VALUES ($1)", target.tableName), migration.ID())
	return err
}

func (target *Target) Remove(migration migrations.Migration) error {
	_, err := target.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = $1", target.tableName), migration.ID())
	return err
}
