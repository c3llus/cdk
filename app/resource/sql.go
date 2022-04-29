package resource

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

// SQLDB interface for infra
type SQLDB interface {
	GetMaster() *sqlx.DB
	GetSlave() *sqlx.DB
	PingContext(ctx context.Context) error
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	SetConnMaxLifetime(time.Duration)

	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
}
