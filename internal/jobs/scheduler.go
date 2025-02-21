package jobs

import (
	"time"

	"github.com/robfig/cron/v3"
)

func NewCronJob(period time.Duration) *cron.Cron {
	return cron.New(cron.WithSeconds())
}
