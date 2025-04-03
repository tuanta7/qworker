package main

import (
	connectoruc "github.com/tuanta7/qworker/internal/connector"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
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
	cfg := config.NewConfig()

	aead, err := cipherx.New([]byte(cfg.AesGsmSecret), cipherx.AEAD)
	if err != nil {
		panic(err)
	}

	zl := logger.MustNewLogger(cfg.Logger.Level)
	ldapClient := ldapclient.NewLDAPClient(cfg.StartTLSConfig.SkipVerify)

	pgClient := db.MustNewPostgresClient(cfg, db.WithMaxConns(10))
	defer pgClient.Close()

	redisClient := db.MustNewRedisSentinelClient(cfg)
	defer redisClient.Close()

	srv := asynq.NewServerFromRedisClient(redisClient,
		asynq.Config{
			Concurrency:    6,
			Queues:         config.QueuePriority,
			StrictPriority: true,
		})

	asynqInspector := asynq.NewInspectorFromRedisClient(redisClient)
	defer asynqInspector.Close()

	userRepository := pgrepo.NewUserRepository(pgClient)
	connectorRepository := pgrepo.NewConnectorRepository(pgClient)
	connectorUsecase := connectoruc.NewUseCase(connectorRepository, zl)
	workerUsecase := workeruc.NewUseCase(asynqInspector, ldapClient, aead, connectorRepository, userRepository, zl)

	mux := NewRouter(cfg, zl, workerUsecase, connectorUsecase)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("asynq server stopped: %v", err)
	}
}
