# QWorker

## Scheduler

- Load configuration from PostgreSQL and update dynamically on database notifications.
- Schedule cron jobs using robfig/cron.
- Each job dispatches a message to Redis using hibiken/asynq.
- The message-sending interval is defined by the loaded configuration.

## Worker

- The worker receives a message, retrieves connector information from the database, and executes the assigned job.
