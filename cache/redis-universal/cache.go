package redis_universal

import (
	"encoding"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/AndreeJait/GO-ANDREE-UTILITIES/cache"
)

type (
	Option struct {
		Address      []string
		Password     string
		DB           int
		PoolSize     int
		MinIdleConns int
		ReadOnly     bool
		DialTimeout  time.Duration
		PoolTimeout  time.Duration
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		MaxConnAge   time.Duration
	}

	redisUniversalClient struct {
		r redis.UniversalClient
	}
)

func New(option *Option) (cache.Cache, error) {
	var client redis.UniversalClient

	client = redis.NewUniversalClient(&redis.UniversalOptions{
		DB:           option.DB,
		Addrs:        option.Address,
		Password:     option.Password,
		PoolSize:     option.PoolSize,
		PoolTimeout:  option.PoolTimeout,
		ReadTimeout:  option.ReadTimeout,
		WriteTimeout: option.WriteTimeout,
		DialTimeout:  option.DialTimeout,
		MinIdleConns: option.MinIdleConns,
		MaxConnAge:   option.MaxConnAge,
		ReadOnly:     option.ReadOnly,
	})

	if _, err := client.Ping().Result(); err != nil {
		return nil, errors.Wrap(err, "Failed to connect to redis!")
	}

	return &redisUniversalClient{r: client}, nil
}

func (c *redisUniversalClient) Ping() error {
	if _, err := c.r.Ping().Result(); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (c *redisUniversalClient) SetWithExpiration(key string, value interface{}, duration time.Duration) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.Set(key, value, duration).Result(); err != nil {
		return errors.Wrapf(err, "failed to set cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) Set(key string, value interface{}) error {
	if err := check(c); err != nil {
		return err
	}

	return c.SetWithExpiration(key, value, 0)
}

func (c *redisUniversalClient) Get(key string, data interface{}) error {
	if _, ok := data.(encoding.BinaryUnmarshaler); !ok {
		return errors.New(fmt.Sprintf("failed to get cache with key %s!: redis: can't unmarshal (implement encoding.BinaryUnmarshaler)", key))
	}

	if err := check(c); err != nil {
		return err
	}

	val, err := c.r.Get(key).Result()

	if err == redis.Nil {
		return errors.Wrapf(err, "key %s does not exits", key)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to get key %s!", key)
	}

	if err := data.(encoding.BinaryUnmarshaler).UnmarshalBinary([]byte(val)); err != nil {
		return err
	}

	return nil
}

func (c *redisUniversalClient) Keys(pattern string) ([]string, error) {
	if err := check(c); err != nil {
		return []string{}, err
	}

	return c.r.Keys(pattern).Result()
}

func (c *redisUniversalClient) Remove(key string) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.Del(key).Result(); err != nil {
		return errors.Wrapf(err, "failed to remove key %s!", key)
	}

	return nil
}

func (c *redisUniversalClient) RemoveByPattern(pattern string, countPerLoop int64) error {
	if err := check(c); err != nil {
		return err
	}

	iteration := 1
	for {
		keys, _, err := c.r.Scan(0, pattern, countPerLoop).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to scan redis pattern %s!", pattern)
		}

		if len(keys) == 0 {
			break
		}

		if _, err := c.r.Del(keys...).Result(); err != nil {
			return errors.Wrapf(err, "failed iteration-%d to remove key with pattern %s", iteration, pattern)
		}

		iteration++
	}

	return nil
}

func (c *redisUniversalClient) FlushDatabase() error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.FlushDB().Result(); err != nil {
		return errors.Wrap(err, "failed to flush db!")
	}

	return nil
}

func (c *redisUniversalClient) FlushAll() error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.FlushAll().Result(); err != nil {
		return errors.Wrap(err, "failed to flush db!")
	}

	return nil
}

func (c *redisUniversalClient) Close() error {
	if err := c.r.Close(); err != nil {
		return errors.Wrap(err, "failed to close redis client")
	}

	return nil
}

func check(c *redisUniversalClient) error {
	if c.r == nil {
		return errors.New("redis client is not connected")
	}

	return nil
}

func (c *redisUniversalClient) SetZSetWithExpiration(key string, duration time.Duration, data ...redis.Z) error {
	if err := c.SetZSet(key, data...); err != nil {
		return err
	}

	if _, err := c.r.Expire(key, duration).Result(); err != nil {
		c.r.Del(key)
		return errors.Wrapf(err, "failed to zadd cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) SetZSet(key string, data ...redis.Z) error {
	if err := check(c); err != nil {
		return err
	}

	c.r.Del(key)
	if _, err := c.r.ZAdd(key, data...).Result(); err != nil {
		return errors.Wrapf(err, "failed to zadd cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) GetZSet(key string) ([]redis.Z, error) {
	if err := check(c); err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := c.r.ZRangeWithScores(key, 0, -1).Result()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run zrange command")
	}

	if len(data) <= 0 {
		return nil, errors.New(fmt.Sprintf("key %s does not exits", key))
	}

	return data, nil
}

func (c *redisUniversalClient) HMSetWithExpiration(key string, value map[string]interface{}, ttl time.Duration) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.HMSet(key, value).Result(); err != nil {
		return errors.Wrapf(err, "failed to HMSet cache with key %s!", key)
	}

	if _, err := c.r.Expire(key, ttl).Result(); err != nil {
		c.r.Del(key)
		return errors.Wrapf(err, "failed to HMSet cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) HMSet(key string, value map[string]interface{}) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.HMSet(key, value).Result(); err != nil {
		return errors.Wrapf(err, "failed to HMSet cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) HSetWithExpiration(key, field string, value interface{}, ttl time.Duration) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.HSet(key, field, value).Result(); err != nil {
		return errors.Wrapf(err, "failed to HSet cache with key %s!", key)
	}
	if _, err := c.r.Expire(key, ttl).Result(); err != nil {
		c.r.Del(key)
		return errors.Wrapf(err, "failed to HMSet cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) HSet(key, field string, value interface{}) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.HSet(key, field, value).Result(); err != nil {
		return errors.Wrapf(err, "failed to HSet cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	if err := check(c); err != nil {
		return nil, err
	}

	val, err := c.r.HMGet(key, fields...).Result()
	if err == redis.Nil {
		return nil, errors.Wrapf(err, "key %s does not exits", key)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key %s!", key)
	}

	return val, nil
}

func (c *redisUniversalClient) HGetAll(key string) (map[string]string, error) {
	if err := check(c); err != nil {
		return nil, err
	}

	val, err := c.r.HGetAll(key).Result()
	if err == redis.Nil {
		return nil, errors.Wrapf(err, "key %s does not exits", key)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key %s!", key)
	}

	return val, nil
}

func (c *redisUniversalClient) HGet(key, field string, response interface{}) error {
	if _, ok := response.(encoding.BinaryUnmarshaler); !ok {
		return errors.New(fmt.Sprintf("failed to get cache with key %s!: redis: can't unmarshal (implement encoding.BinaryUnmarshaler)", key))
	}

	if err := check(c); err != nil {
		return err
	}

	val, err := c.r.HGet(key, field).Result()
	if err == redis.Nil {
		return errors.Wrapf(err, "key %s does not exits", key)
	}

	if err != nil {
		return errors.Wrapf(err, "failed to get key %s!", key)
	}

	if err := response.(encoding.BinaryUnmarshaler).UnmarshalBinary([]byte(val)); err != nil {
		return err
	}

	return nil
}

func (c *redisUniversalClient) MGet(key []string) ([]interface{}, error) {
	if err := check(c); err != nil {
		return nil, err
	}

	val, err := c.r.MGet(key...).Result()
	if err == redis.Nil {
		return nil, errors.Wrapf(err, "key %s does not exits", key)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key %s!", key)
	}

	return val, nil
}

func (c *redisUniversalClient) MSetWithExpiration(keys []string, values []interface{}, ttls []time.Duration) error {
	if len(ttls) < len(keys) {
		return errors.New("values or ttls must have same count with keys")
	}

	var failedKeys []string

	if err := c.MSet(keys, values); err != nil {
		return errors.WithStack(err)
	}

	for i := range keys {
		if _, err := c.r.Expire(keys[i], ttls[i]).Result(); err != nil {
			failedKeys = append(failedKeys, keys[i])
			c.r.Del(keys[i])
		}
	}

	if len(failedKeys) > 0 {
		return errors.New("failed to set some keys " + strings.Join(failedKeys, ","))
	}
	return nil
}

func (c *redisUniversalClient) MSet(keys []string, values []interface{}) error {
	if err := check(c); err != nil {
		return errors.WithStack(err)
	}

	if len(values) < len(keys) {
		return errors.New("values or ttls must have same count with keys")
	}

	var pairs []interface{}

	for i := range keys {
		pairs = append(pairs, keys[i], values[i])
	}
	_, err := c.r.MSet(pairs...).Result()

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (c *redisUniversalClient) SetNx(key string, value interface{}, ttl time.Duration) (bool, error) {
	if err := check(c); err != nil {
		return false, errors.WithStack(err)
	}

	nx := c.r.SetNX(key, value, ttl)
	return nx.Result()
}

func (c *redisUniversalClient) Client() cache.Cache {
	return c
}

func (c *redisUniversalClient) Pipeline() cache.Pipe {
	return &pipe{instance: c.r.Pipeline()}
}

func (c *redisUniversalClient) Subscribe(channel string) (cache.PubSub, error) {
	return nil, errors.New("pubsub not implemented in universal client")
}

func (c *redisUniversalClient) HDel(key string, fields ...string) error {
	if err := check(c); err != nil {
		return err
	}

	if _, err := c.r.HDel(key, fields...).Result(); err != nil {
		return errors.Wrapf(err, "failed to HDel cache with key %s!", key)
	}
	return nil
}

func (c *redisUniversalClient) ZIncrBy(key string, increment float64, member string) (float64, error) {
	if err := check(c); err != nil {
		return 0, errors.WithStack(err)
	}

	res := c.r.ZIncrBy(key, increment, member)
	return res.Result()
}

func (c *redisUniversalClient) TTL(key string) (duration time.Duration, err error) {
	if err = check(c); err != nil {
		return duration, err
	}

	duration, err = c.r.TTL(key).Result()

	if err != nil {
		return duration, errors.Wrapf(err, "failed to get TTL with key %s!", key)
	}

	return duration, nil
}

func (c *redisUniversalClient) Incr(key string) error {
	if err := check(c); err != nil {
		return errors.WithStack(err)
	}

	return c.r.Incr(key).Err()
}

func (c *redisUniversalClient) IncrBy(key string, value int64) error {
	if err := check(c); err != nil {
		return errors.WithStack(err)
	}

	return c.r.IncrBy(key, value).Err()
}
