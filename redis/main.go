package redis

import (
	"github.com/go-redis/redis"
	redistore "gopkg.in/boj/redistore.v1"
)

var (
	// Store Redis Session
	Store *redistore.RediStore
)

// Client Redis Client
var Client *redis.Client

// ConnectRedisStore session store
func ConnectRedisStore(host string, password string, port string, key []byte) (err error) {
	Store, err = redistore.NewRediStore(10, "tcp", host+":"+port, password, key)

	return err
}

// ConnectRedis Redis Client
func ConnectRedis(host, password, port string) (err error) {
	// 建立連線
	Client = redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password,
		DB:       0, // use default DB
	})

	_, err = Client.Ping().Result()

	return err
}
