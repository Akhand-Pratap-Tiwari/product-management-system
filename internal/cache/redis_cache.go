package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(host string, port int) *RedisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})

	return &RedisCache{client: rdb}
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, jsonData, expiration).Err()
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal([]byte(result), &data)
	return data, err
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
