package resource

type Resource interface {
	GetSQLDB(dbname string) (SQLDB, error)
	// TODO:
	// GetRedis(redisname string) (Redis, error)
}
