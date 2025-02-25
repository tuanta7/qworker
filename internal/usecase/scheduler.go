package usecase

import (
	"time"

	"github.com/robfig/cron/v3"
	pgrepo "github.com/tuanta7/qworker/internal/repository/postgres"
)

type SchedulerUsecase struct {
	schedulerRepo *pgrepo.SchedulerRepository
}

func NewSchedulerUsecase(schedulerRepo *pgrepo.SchedulerRepository) *SchedulerUsecase {
	return &SchedulerUsecase{
		schedulerRepo: schedulerRepo,
	}
}

func (u *SchedulerUsecase) NewJob(period time.Duration) *cron.Cron {
	return cron.New(cron.WithSeconds())
}

func (u *SchedulerUsecase) TerminateJob(job *cron.Cron) {
	job.Stop()
}

func (u *SchedulerUsecase) NewSyncMessage() []byte {
	return nil
}
