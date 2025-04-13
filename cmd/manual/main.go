package main

import (
	"encoding/json"
	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/config"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"strconv"
	"time"
)

func main() {
	cfg := config.InitConfig()
	zl := logger.MustNewLogger(cfg.Logger.Level)

	redisClient := db.MustNewRedisSentinelClient(cfg)
	defer redisClient.Close()

	asynqClient := asynq.NewClientFromRedisClient(redisClient)
	defer asynqClient.Close()

	message := &domain.QueueMessage{
		ConnectorID: 2,
		TaskType:    config.QueueTask[config.QueueFullSync],
	}
	taskID := strconv.FormatUint(message.ConnectorID, 10)
	payload, _ := json.Marshal(message)

	for {
		info, err := asynqClient.Enqueue(
			asynq.NewTask(message.TaskType, payload),
			asynq.TaskID(taskID),
			asynq.Queue(config.QueueFullSync),
			asynq.MaxRetry(0),
			asynq.Retention(0))
		if err != nil {
			zl.Error("Enqueue failed", zap.Error(err))
		}
		zl.Info("Enqueued", zap.Any("info", info))
		time.Sleep(10 * time.Second)
	}

}
