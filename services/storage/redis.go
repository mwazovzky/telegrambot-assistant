package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisService struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisService(client *redis.Client, ttl time.Duration) *RedisService {
	return &RedisService{client, ttl}
}

func (rs *RedisService) Exists(key string) (bool, error) {
	ctx := context.Background()

	exists, err := rs.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if exists == 0 {
		return false, nil
	}
	return true, nil
}

func (rs *RedisService) Get(key string) (string, error) {
	ctx := context.Background()
	return rs.client.Get(ctx, key).Result()
}

func (rs *RedisService) Set(key string, value string) error {
	ctx := context.Background()
	return rs.client.Set(ctx, key, value, rs.ttl).Err()
}
