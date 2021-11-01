package migrationsql

import (
	"database/sql"
	"fmt"

	"github.com/jamillosantos/migrations"
)

type DB interface {
	DBExecer

	Conn() (*sql.Conn, error)
}

type Target struct {
	source migrations.Source
	driver Driver
}

type Option func(target *Target) error

func NewTarget(source migrations.Source, driver Driver) (*Target, error) {
	target := &Target{
		source,
		driver,
	}
	return target, nil
}

func (target *Target) Create() error {
	return target.driver.CreateMigrationsTable()
}

func (target *Target) Destroy() error {
	return target.driver.DropMigrationsTable()
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
	rs, err := target.driver.Query(fmt.Sprintf("SELECT id FROM %s ORDER BY id ASC", target.tableName))
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
	_, err := target.driver.Exec(fmt.Sprintf("INSERT INTO %s (id) VALUES (?)", target.tableName), migration.ID())
	return err
}

func (target *Target) Remove(migration migrations.Migration) error {
	_, err := target.driver.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = ?", target.tableName), migration.ID())
	return err
}
