package db

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/tuanta7/qworker/config"
)

func NewRedisSentinelClient(cfg *config.Config) (*redis.Client, error) {
	c := redis.NewFailoverClient(&redis.FailoverOptions{
		SentinelAddrs: cfg.Redis.Sentinels,
		MasterName:    cfg.Redis.MasterName,
		Password:      cfg.Redis.Password,
		DB:            cfg.Redis.Database,
		PoolSize:      10,
		MinIdleConns:  3,
	})

	ctx := context.Background()

	err := c.Set(ctx, "key", "value", time.Minute).Err()
	if err != nil {
		return nil, err
	}

	_, err = c.Get(ctx, "key").Result()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func MustNewRedisSentinelClient(cfg *config.Config) *redis.Client {
	client, err := NewRedisSentinelClient(cfg)
	if err != nil {
		log.Fatalf("db.NewRedisSentinelClient(): %v", err)
	}
	return client
}
