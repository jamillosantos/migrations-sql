package target

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jamillosantos/migrations"
)

var (
	ErrMissingSource = errors.New("source is required")
)

type DBExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type DB interface {
	DBExecer
	Begin() (*sql.Tx, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

type Target struct {
	source    migrations.Source
	db        DB
	tableName string
	driver    Driver
}

type Option func(target *Target) error

type Driver interface {
	migrations.TargetLocker
	Add(migration migrations.Migration) error
	Remove(migration migrations.Migration) error
}

const DefaultMigrationsTableName = "_migrations"

func NewTarget(source migrations.Source, db DB, driver Driver, options ...Option) (*Target, error) {
	if source == nil {
		return nil, ErrMissingSource
	}
	target := &Target{
		source:    source,
		db:        db,
		driver:    driver,
		tableName: DefaultMigrationsTableName,
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
	return target.driver.Add(migration)
}

func (target *Target) Remove(migration migrations.Migration) error {
	return target.driver.Remove(migration)
}
