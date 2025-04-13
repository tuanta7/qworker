package main

import (
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
	"github.com/tuanta7/qworker/internal/usecase/connector"
	"github.com/tuanta7/qworker/internal/usecase/worker"
	"github.com/tuanta7/qworker/pkg/cipherx"
	"github.com/tuanta7/qworker/pkg/ldapclient"
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
)

func main() {
	cfg := config.InitConfig()

	aead, err := cipherx.New(cipherx.AEAD, []byte(cfg.AESSecret))
	if err != nil {
		panic(err)
	}

	zl := logger.MustNewLogger(cfg.Logger.Level)
	ldapClient := ldapclient.NewLDAPClient(cfg.StartTLS.SkipVerify)

	pgClient := db.MustNewPostgresClient(cfg, db.WithMaxConns(10), db.WithMinConns(3))
	defer pgClient.Close()

	redisClient := db.MustNewRedisSentinelClient(cfg)
	defer redisClient.Close()

	asynqInspector := asynq.NewInspectorFromRedisClient(redisClient)
	defer asynqInspector.Close()

	srv := asynq.NewServerFromRedisClient(redisClient, asynq.Config{
		Concurrency:    6,
		Queues:         config.QueuePriority,
		StrictPriority: true,
	})

	userRepository := pgrepo.NewUserRepository(pgClient)
	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	connectorUsecase := connectoruc.NewUseCase(connectorRepository, zl)
	workerUsecase := workeruc.NewUseCase(asynqInspector, ldapClient, aead, connectorRepository, userRepository, zl)

	mux := NewRouter(cfg, zl, workerUsecase, connectorUsecase)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server stopped: %v", err)
	}
}
