package repositories

import (
	"context"
	"database/sql"
)

// DatabaseTransactioner defines the common query methods shared by *sql.DB and *sql.Tx.
type DatabaseTransactioner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Databaser extends DatabaseTransactioner with connection-level operations.
type Databaser interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	DatabaseTransactioner
	PingContext(ctx context.Context) error
	Close() error
}

// Transactioner extends DatabaseTransactioner with transaction commit/rollback.
type Transactioner interface {
	Commit() error
	Rollback() error
	DatabaseTransactioner
}
