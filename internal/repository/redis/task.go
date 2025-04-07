package redisrepo

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

type TaskRepository struct {
	*redis.Client
}

func NewTaskRepository(client *redis.Client) *TaskRepository {
	return &TaskRepository{client}
}

func (r *TaskRepository) Exist(ctx context.Context, queue, id string) (bool, error) {
	key := fmt.Sprintf("async{%s}:t:%s", queue, id)
	val, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return val > 0, nil
}
