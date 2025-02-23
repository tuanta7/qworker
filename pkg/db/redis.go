package db

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/tuanta7/qworker/config"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisSentinelClient(cfg *config.Config) (*RedisClient, error) {
	c := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs: cfg.Redis.Sentinels,
		MasterName:    cfg.Redis.MasterName,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.Database,
	})

	ctx := context.Background()

	err := c.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		return nil, err
	}

	_, err = c.Get(ctx, "key").Result()
	if err != nil {
		return nil, err
	}

	return &RedisClient{c}, nil
}

func (c *RedisClient) Close() {
	c.Client.Close()
}
