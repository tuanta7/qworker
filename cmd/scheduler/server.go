package main

import (
	"context"
	"encoding/json"
	"github.com/tuanta7/qworker/internal/domain"
	"github.com/tuanta7/qworker/pkg/db"
	"github.com/tuanta7/qworker/pkg/logger"
	"go.uber.org/zap"
	"strings"
	"time"
)

type Scheduler struct {
	pgClient db.PostgresClient
	zl       *logger.ZapLogger
	handlers map[string]SchedulerHandlerFunc
}

type SchedulerHandlerFunc func(c context.Context, msg *domain.NotifyMessage) error

func NewScheduler(pgClient db.PostgresClient, zl *logger.ZapLogger) *Scheduler {
	return &Scheduler{
		pgClient: pgClient,
		zl:       zl,
		handlers: make(map[string]SchedulerHandlerFunc),
	}
}

func (s *Scheduler) Listen(ctx context.Context, channelName string, buffer int) {
	conn, err := s.pgClient.Pool().Acquire(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, "LISTEN "+channelName)
	if err != nil {
		panic(err)
	}
	defer conn.Exec(ctx, "UNLISTEN "+channelName)

	notifyChan := make(chan string, buffer)
	go s.ProcessNotifications(notifyChan)

	for {
		notification, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			s.zl.Error("listen - conn.Conn().WaitForNotification", zap.Error(err))
			time.Sleep(10 * time.Second)
			continue
		}

		if notification.Channel == channelName {
			notifyChan <- notification.Payload
		}
	}
}

func (s *Scheduler) ProcessNotifications(notifyChan <-chan string) {
	for n := range notifyChan {
		s.zl.Info("notification received", zap.Any("notification", n))

		message := &domain.NotifyMessage{}
		err := json.Unmarshal([]byte(n), message)
		if err != nil {
			s.zl.Error("failed to unmarshal notification message", zap.Error(err))
			continue
		}

		requestHandler, exists := s.handlers[strings.ToLower(message.Action)]
		if exists {
			err = requestHandler(context.TODO(), message)
			if err != nil {
				s.zl.Warn("error while handling trigger action", zap.Error(err))
				continue
			}
		} else {
			s.zl.Error("unknown action", zap.String("action", message.Action))
		}
	}
}

func (s *Scheduler) RegisterHandler(action string, handler func(c context.Context, msg *domain.NotifyMessage) error) {
	if s.handlers == nil {
		s.handlers = make(map[string]SchedulerHandlerFunc)
	}
	s.handlers[action] = handler
}
