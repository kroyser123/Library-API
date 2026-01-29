package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	ctx    context.Context
	client *redis.Client
}

func NewRedisCache(host, port, password string, db int) Cache {
	return &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password,
			DB:       db,
		}),
		ctx: context.Background(),
	}
}
func (c *RedisCache) Set(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(c.ctx, key, data, ttl).Err()
}
func (c *RedisCache) Get(key string, value interface{}) error {
	data, err := c.client.Get(c.ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}
func (c *RedisCache) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}
func (c *RedisCache) Clear() error {
	return c.client.FlushDB(c.ctx).Err()
}
func (c *RedisCache) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}
