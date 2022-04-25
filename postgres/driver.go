package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jamillosantos/migrations"
	"github.com/spaolacci/murmur3"

	"github.com/jamillosantos/migrations-sql/target"
)

type TXExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Commit() error
	Rollback() error
}

type Postgres struct {
	db           target.DB
	databaseName string
	tableName    string
}

func New(source migrations.Source, db target.DB, databaseName, tableName string) (*target.Target, error) {
	return target.NewTarget(
		source,
		db,
		&Postgres{
			db:           db,
			databaseName: databaseName,
			tableName:    tableName,
		},
		target.Table(tableName),
	)
}

// pgLocker is the migrations.Locker implementation for Postgres database. Its job is to block other instances of the
// migration system to run at the same time. In other to achieve this, it uses the database and table name to create a
// unique key that is hashed (using murmur3) to a bigint. Then, an advisory lock is created using that key.
type pgLocker struct {
	db   TXExecer
	code int64
}

func (p *pgLocker) Unlock() error {
	_, err := p.db.Exec("SELECT pg_advisory_unlock($1)", p.code)
	if err != nil {
		_ = p.db.Rollback()
		return fmt.Errorf("failed unlocking migration: %w", err)
	}
	_ = p.db.Commit()
	return nil
}

func (p *Postgres) Lock() (migrations.Unlocker, error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed starting transaction for locking: %w", err)
	}

	advisoryLockID, err := p.generateLockID()
	if err != nil {
		return nil, fmt.Errorf("failed locking database: %w", err)
	}
	_, err = tx.Exec("SELECT pg_advisory_lock($1)", advisoryLockID)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed locking database: %w", err)
	}
	return &pgLocker{db: tx, code: advisoryLockID}, nil
}

func (p *Postgres) Add(migration migrations.Migration) error {
	_, err := p.db.Exec(fmt.Sprintf("INSERT INTO %s (id) VALUES ($1)", p.tableName), migration.ID())
	if err != nil {
		return fmt.Errorf("failed adding migration to the executed list: %w", err)
	}
	return nil
}

func (p *Postgres) Remove(migration migrations.Migration) error {
	_, err := p.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = $1", p.tableName), migration.ID())
	if err != nil {
		return fmt.Errorf("failed removing migration from the executed list: %w", err)
	}
	return nil
}

func (p *Postgres) generateLockID() (int64, error) {
	hasher := murmur3.New64()
	if _, err := hasher.Write([]byte(p.databaseName)); err != nil {
		return 0, err
	}
	if _, err := hasher.Write([]byte("|||")); err != nil {
		return 0, err
	}
	if _, err := hasher.Write([]byte(p.tableName)); err != nil {
		return 0, err
	}
	return int64(hasher.Sum64()), nil
}
