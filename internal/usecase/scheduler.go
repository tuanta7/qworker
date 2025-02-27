package usecase

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/tuanta7/qworker/internal/domain"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
)

type SchedulerUsecase struct {
	schedulerRepo *pgrepo.SchedulerRepository
	asynqClient   *asynq.Client
}

func NewSchedulerUsecase(schedulerRepo *pgrepo.SchedulerRepository, asynqClient *asynq.Client) *SchedulerUsecase {
	return &SchedulerUsecase{
		schedulerRepo: schedulerRepo,
		asynqClient:   asynqClient,
	}
}

func (u *SchedulerUsecase) SendSyncMessage(connectorID uint64) func() {
	return func() {
		message := domain.Message{
			ConnectorID: connectorID,
			JobType:     domain.JobTypeIncrementalSync,
		}

		fmt.Println("Send sync message:", message)

		payload, err := json.Marshal(message)
		if err != nil {
			log.Printf("Marshal JSON: %v", err)
		}

		task := asynq.NewTask("user:sync", payload)
		if _, err := u.asynqClient.Enqueue(task); err != nil {
			log.Printf("Enqueue task: %v", err)
		}
	}
}
