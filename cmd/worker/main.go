package main

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/pkg/db"
)

func main() {
	cfg := config.NewConfig()

	pgClient, err := db.NewPostgresClient(cfg)
	if err != nil {
		log.Fatalf("Postgres: %v", err)
	}
	defer pgClient.Close()

	srv := asynq.NewServer(
		asynq.RedisFailoverClientOpt{
			MasterName:    cfg.Redis.MasterName,
			SentinelAddrs: cfg.Redis.Sentinels,
			Password:      cfg.Redis.Password,
			DB:            cfg.Redis.Database,
		}, asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":  1,
				"critical": 5,
			},
		})

	mux := asynq.NewServeMux()
	mux.HandleFunc("message:send", nil)

	if err := srv.Run(mux); err != nil {
		log.Fatalf("Asynq server stopped: %v", err)
	}
}
