package hodler

import (
	"context"
	"errors"
	"sync"

	"github.com/c3llus/cdk/app/resource"
	sqlcdk "github.com/c3llus/cdk/sql"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
)

type Hodler struct {
	dbs map[string]*sqlcdk.DB
	rds map[string]*redis.Client
}

func New(ctx context.Context) (*Hodler, error) {
	var (
		hodler = Hodler{
			dbs: make(map[string]*sqlcdk.DB),
			rds: make(map[string]*redis.Client),
		}

		group sync.WaitGroup
		mux   sync.Mutex
		errs  []error
		err   error
	)

	appendErr := func(err error) {
		// errs is being used by many goroutines,
		// so we need to protect it using mutex
		mux.Lock()
		errs = append(errs, err)
		mux.Unlock()
	}

	// connect to databases
	go hodler.connectSQLDB(ctx, &group, appendErr)

	// connect to redis
	go hodler.connectRedis(ctx, &group, appendErr)

	// wait for all connections attempt
	group.Wait()

	// check for error, if error length is greater than 1
	// set err to errs[0]
	if len(errs) > 0 {
		err = errs[0]
	}

	return &hodler, err
}

func (hdlr *Hodler) connectSQLDB(ctx context.Context, group *sync.WaitGroup, appendErr func(err error)) {

	// TODO:
	// 1. Support multiple driver; not only Postgres
	// 2. Support Master-Slave

	group.Add(1)
	defer group.Done()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	hdlr.dbs[dbname] = &sqlcdk.DB{
		Master: db,
		Slave:  db,
	}
}

func (hdlr *Hodler) connectRedis(ctx context.Context, group *sync.WaitGroup, appendErr func(err error)) {

}

func (hdlr *Hodler) GetSQLDB(dbname string) (resource.SQLDB, error) {
	res, ok := hdlr.dbs[dbname]
	if !ok {
		err := errors.New("GetSQLDB")
		return nil, err
	}
	return res, nil
}

func (hdlr *Hodler) GetRedis(redisname string) (*redis.Client, error) {
	res, ok := hdlr.rds[redisname]
	if !ok {
		err := errors.New("GetRedis")
		return nil, err
	}
	return res, nil
}
