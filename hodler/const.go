package hodler

import "fmt"

var (
	dsn = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
)

// TODO: passed via config
const (
	host     = "localhost"
	port     = 6969
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
	driver   = "postgres"
)
