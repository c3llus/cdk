package sqlcdk

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	// TODO:
	// 1. Separate Master and Slave
	// 2. support sql; not only sqlx

	Master
	Slave

	master *sqlx.DB
	slave  *sqlx.DB

	// driver string // multi driver support
}

func (db *DB) GetMaster() *sqlx.DB {
	return db.master
}

func (db *DB) GetSlave() *sqlx.DB {
	return db.slave
}

func (db *DB) PingContext(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- db.master.PingContext(ctx)
	}()

	go func() {
		errCh <- db.slave.PingContext(ctx)
	}()

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) SetMaxIdleConns(n int) {
	db.master.SetMaxIdleConns(n)
	db.slave.SetMaxIdleConns(n)
}

func (db *DB) SetMaxOpenConns(n int) {
	db.master.SetMaxOpenConns(n)
	db.slave.SetMaxOpenConns(n)
}

func (db *DB) SetConnMaxLifetime(t time.Duration) {
	db.master.SetConnMaxLifetime(t)
	db.slave.SetConnMaxLifetime(t)
}

type Master interface {
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

type Slave interface {
}
