# QWorker

## Scheduler

- Load configuration from PostgreSQL and update dynamically on database notifications.
- Schedule cron jobs using robfig/cron.
- Each job dispatches a message to Redis using hibiken/asynq.
- The message-sending interval is defined by the loaded configuration.

## Worker

- The worker receives a message, retrieves connector information from the database, and executes the assigned job.

## Notes

### Redis Sentinel

Redis Sentinel is a distributed system, it provides high availability for Redis when not using Redis Cluster.

- Sentinel nodes continuously monitor the Redis master. If the master node becomes unavailable, the Sentinel nodes initiate an election process.
- During this process, they select one of the slave nodes to become the new master.

### Asynq Package

- Asynq uses `robfig/cron` for scheduling and `go-redis/redis` for interacting with Redis.
