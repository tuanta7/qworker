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

- Sentinel nodes continuously monitor the Redis master. If the master node becomes unavailable, the Sentinel nodes
  initiate an election process.
- During this process, they select one of the slave nodes to become the new master.

### Asynq Package

- Asynq uses `robfig/cron` for scheduling and `go-redis/redis` for interacting with Redis.

### PostgreSQL

#### Change Data Capture

- Listen/Notify (Pub-Sub): PostgreSQL’s built-in LISTEN/NOTIFY mechanism enables real-time event-driven notifications, ideal for lightweight CDC scenarios.
- Streaming Replication: PostgreSQL supports streaming WAL replication to other PostgreSQL databases, which can be leveraged to capture and process data changes
  in real time.

#### Procedures vs Functions

| Procedure                                                                 | Function                                                   |
|---------------------------------------------------------------------------|------------------------------------------------------------|
| Can perform transactions                                                  | Cannot perform transactions(COMMIT, ROLLBACK not allowed). |
| Cannot be used in SQL queries, executed using CALL keyword                | Can be used in SQL queries (e.g., SELECT my_function()).   |
| Does not return a value (but can use OUT parameters).                     | Returns a value, must use RETURN to provide a result.      |
| Used for complex operations like batch updates, logging, and admin tasks. | Used for computations, data retrieval, or transformations. |

#### Trigger Function

- A function with RETURNS TRIGGER must be called by a CREATE TRIGGER statement.
- `pg_notify()` can not be used in a trigger because it does not return a trigger 

#### Row Variables inside Trigger function

- `NEW`: The new row after an INSERT or UPDATE.
- `OLD`: The old row before an UPDATE or DELETE.
- `TG_OP` (Trigger Operation): Trigger type (INSERT, UPDATE, DELETE, or TRUNCATE).
- `TG_TABLE_NAME`: The name of the table the trigger is on.
- `TG_RELID`: The table’s OID (Object Identifier).
- `TG_NARGS`: Number of arguments passed to the trigger
- `TG_ARGV[]`: Array of trigger arguments

#### Delete vs Truncate