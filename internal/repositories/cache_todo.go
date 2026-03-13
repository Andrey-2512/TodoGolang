package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"todo/domain/entity"

	"github.com/redis/go-redis/v9"
)

type cacheTaskRepository struct {
	taskRepo       entity.TaskRepository
	redisClient    *redis.Client
	userTaskPrefix string
	taskPrefix     string
	cacheTTL       time.Duration
}

func (c *cacheTaskRepository) taskKey(id, userId int) string {
	return fmt.Sprintf("%s%d:%s%d", c.taskPrefix, id, c.userTaskPrefix, userId)
}

func (c *cacheTaskRepository) userTasksKey(userId int) string {
	return fmt.Sprintf("%s%s%d", c.taskPrefix, c.userTaskPrefix, userId)
}

func NewCacheTaskRepository(taskRepo entity.TaskRepository, redisClient *redis.Client, cacheTTL time.Duration, taskPrefix, userTaskPrefix string) entity.TaskRepository {
	return &cacheTaskRepository{taskRepo: taskRepo, redisClient: redisClient, taskPrefix: taskPrefix, userTaskPrefix: userTaskPrefix, cacheTTL: cacheTTL}
}

func (c *cacheTaskRepository) Create(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	createdTask, err := c.taskRepo.Create(ctx, t)
	if err != nil {
		return nil, err
	}

	allTasksKey := c.userTasksKey(createdTask.UserId)
	c.redisClient.Del(ctx, allTasksKey)

	return createdTask, nil

}

func (c *cacheTaskRepository) GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error) {
	key := c.taskKey(id, userId)
	res, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			task, err := c.taskRepo.GetUserTaskById(ctx, id, userId)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(task)
			if err == nil {
				c.redisClient.Set(ctx, key, data, c.cacheTTL)
				return task, nil
			}
		}
	}
	var task entity.Task
	err = json.Unmarshal([]byte(res), &task)
	if err != nil {
		c.redisClient.Del(ctx, key)
		return c.taskRepo.GetUserTaskById(ctx, id, userId)
	}
	return &task, nil

}

func (c *cacheTaskRepository) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	key := c.userTasksKey(userId)
	res, err := c.redisClient.Get(ctx, key).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			tasks, err := c.taskRepo.GetAllUserTasks(ctx, userId)
			if err != nil {
				return nil, err
			}
			data, err := json.Marshal(tasks)
			if err == nil {
				c.redisClient.Set(ctx, key, data, c.cacheTTL)
			}

			return tasks, nil
		}
	}

	tasks := make([]entity.Task, 0)
	err = json.Unmarshal([]byte(res), &tasks)
	if err != nil {
		c.redisClient.Del(ctx, key)
		return c.taskRepo.GetAllUserTasks(ctx, userId)
	}
	return tasks, nil

}

func (c *cacheTaskRepository) Update(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	updatedTask, err := c.taskRepo.Update(ctx, t)
	if err != nil {
		return nil, err
	}
	taskKey := c.taskKey(updatedTask.Id, updatedTask.UserId)
	allTasksKey := c.userTasksKey(updatedTask.UserId)
	c.redisClient.Del(ctx, taskKey, allTasksKey)

	return updatedTask, nil

}
func (c *cacheTaskRepository) Delete(ctx context.Context, id, userId int) error {
	err := c.taskRepo.Delete(ctx, id, userId)
	if err != nil {
		return err
	}
	taskKey := c.taskKey(id, userId)
	allTasksKey := c.userTasksKey(userId)

	c.redisClient.Del(ctx, taskKey, allTasksKey)
	return nil
}
