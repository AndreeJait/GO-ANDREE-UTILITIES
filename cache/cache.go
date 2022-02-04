package cache

import (
	"time"

	"github.com/AndreeJait/GO-ANDREE-UTILITIES/util"

	"github.com/go-redis/redis"
)

type (
	Pipe interface {
		Set(key string, value interface{}) error
		SetWithExpiration(key string, value interface{}, expired time.Duration) error
		Get(key string, object interface{}) error
		Exec() error
	}

	PubSub interface {
		Receive() error
		Publish(message string) error
		Channel() <-chan *redis.Message
		Close() error
	}

	Cache interface {
		util.Ping
		SetWithExpiration(string, interface{}, time.Duration) error
		Set(string, interface{}) error
		Get(string, interface{}) error

		SetZSetWithExpiration(string, time.Duration, ...redis.Z) error
		SetZSet(string, ...redis.Z) error
		GetZSet(string) ([]redis.Z, error)

		HMSetWithExpiration(key string, value map[string]interface{}, ttl time.Duration) error
		HMSet(key string, value map[string]interface{}) error
		HSetWithExpiration(key, field string, value interface{}, ttl time.Duration) error
		HSet(key, field string, value interface{}) error
		HMGet(key string, fields ...string) ([]interface{}, error)
		HGetAll(key string) (map[string]string, error)
		HGet(key, field string, response interface{}) error
		HDel(key string, fields ...string) error
		MSetWithExpiration(keys []string, values []interface{}, ttls []time.Duration) error
		MSet(keys []string, values []interface{}) error
		MGet(key []string) ([]interface{}, error)
		SetNx(key string, value interface{}, ttl time.Duration) (bool, error)

		Keys(string) ([]string, error)
		TTL(key string) (time.Duration, error)

		Remove(string) error
		RemoveByPattern(string, int64) error
		FlushDatabase() error
		FlushAll() error
		Close() error

		Pipeline() Pipe
		Client() Cache
		Subscribe(channel string) (PubSub, error)
		ZIncrBy(key string, increment float64, member string) (float64, error)
		Incr(key string) error
		IncrBy(key string, value int64) error
	}

	PoolCallback func(client Cache)

	Pool interface {
		Use(callback PoolCallback)
		Client() Cache
		Close() error
	}
)
