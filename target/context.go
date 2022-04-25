package target

import (
	"context"
	"errors"
)

var (
	ErrDBInstanceNotFound = errors.New("db instance not found on the context")
	ErrInvalidDBInstance  = errors.New("context has an invalid db instance")
)

type sourceContextKey string

const (
	dbContextKey sourceContextKey = "db"
)

func (s sourceContextKey) String() string {
	return "migrations_sql_source_" + string(s)
}

// ContextWithDB returns a context with the given db attached.
func ContextWithDB(ctx context.Context, db DB) context.Context {
	return context.WithValue(ctx, dbContextKey, db)
}

func DBFromContext(ctx context.Context) (DBExecer, error) {
	dbInterface := ctx.Value(dbContextKey)
	if dbInterface == nil {
		return nil, ErrDBInstanceNotFound
	}

	db, ok := dbInterface.(DBExecer)
	if !ok {
		return nil, ErrInvalidDBInstance
	}

	return db, nil
}
