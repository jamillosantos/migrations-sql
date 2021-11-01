package postgres

import (
	"database/sql"
	"fmt"
)

type Driver struct {
	TableName string
}

func (p *Driver) DB() *sql.DB {
	panic("implement me")
}

func (p *Driver) Lock() error {
	panic("implement me")
}

func (p *Driver) Unlock() error {
	panic("implement me")
}

func (p *Driver) Exec(query string, args ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (p *Driver) Conn() (*sql.Conn, error) {
	panic("implement me")
}

func (p *Driver) Query(query string, args ...interface{}) (*sql.Rows, error) {
	panic("implement me")
}

func (p *Driver) CreateMigrationsTable() error {
	_, err := p.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id VARCHAR(255) PRIMARY KEY, dirty bool, run_at DATETIME)", p.TableName))
	return err
}

func (p *Driver) DropMigrationsTable() error {
	_, err := p.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", p.TableName))
	return err
}
