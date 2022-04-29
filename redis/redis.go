package redis

import (
	"github.com/go-redis/redis"
)

type Redis struct {
	Redis redis.Conn
}
