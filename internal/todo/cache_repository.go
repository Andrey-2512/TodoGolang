package todo

import (
	"context"
	"fmt"
	"time"
	"todo/domain/entity"

	"encoding/json"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type CacheTaskRepository struct {
	taskRepo       taskRepository
	redisClient    *redis.Client
	userTaskPrefix string
	taskPrefix     string
	cacheTTL       time.Duration
	sft            singleflight.Group
}

func (c *CacheTaskRepository) taskKey(id, userId int) string {
	return fmt.Sprintf("%s%d:%s%d", c.taskPrefix, id, c.userTaskPrefix, userId)
}

func (c *CacheTaskRepository) userTasksKey(userId int) string {
	return fmt.Sprintf("%s%s%d", c.taskPrefix, c.userTaskPrefix, userId)
}

func NewCacheTaskRepository(taskRepo taskRepository, redisClient *redis.Client, cacheTTL time.Duration, taskPrefix, userTaskPrefix string) *CacheTaskRepository {
	return &CacheTaskRepository{taskRepo: taskRepo, redisClient: redisClient, taskPrefix: taskPrefix, userTaskPrefix: userTaskPrefix, cacheTTL: cacheTTL}
}

func (c *CacheTaskRepository) CreateAndCheckLimit(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	createdTask, err := c.taskRepo.CreateAndCheckLimit(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("failed create task in pg: %w", err)
	}

	allTasksKey := c.userTasksKey(createdTask.UserId)
	c.redisClient.Del(ctx, allTasksKey)

	return createdTask, nil

}

func (c *CacheTaskRepository) GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error) {
	key := c.taskKey(id, userId)
	res, err := c.redisClient.Get(ctx, key).Result()
	if err == nil {
		var task entity.Task
		err = json.Unmarshal([]byte(res), &task)
		if err == nil {
			return &task, nil
		}
		c.redisClient.Del(ctx, key)

	}

	v, err, _ := c.sft.Do(key, func() (any, error) {
		task, err := c.taskRepo.GetUserTaskById(context.WithoutCancel(ctx), id, userId)
		if err != nil {
			return nil, fmt.Errorf("failed get task in pg: %w", err)
		}
		data, err := json.Marshal(task)
		if err == nil {
			c.redisClient.Set(ctx, key, data, c.cacheTTL)
		}
		return task, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed get task: %w", err)
	}
	return v.(*entity.Task), nil
}

func (c *CacheTaskRepository) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	key := c.userTasksKey(userId)
	res, err := c.redisClient.Get(ctx, key).Result()

	if err == nil {
		tasks := make([]entity.Task, 0)
		err = json.Unmarshal([]byte(res), &tasks)
		if err == nil {
			return tasks, nil
		}
		c.redisClient.Del(ctx, key)
	}
	v, err, _ := c.sft.Do(key, func() (any, error) {
		tasks, err := c.taskRepo.GetAllUserTasks(context.WithoutCancel(ctx), userId)
		if err != nil {
			return nil, fmt.Errorf("failed get all tasks in pg: %w", err)
		}

		data, err := json.Marshal(tasks)
		if err == nil {
			c.redisClient.Set(ctx, key, data, c.cacheTTL)
		}

		return tasks, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed get all tasks: %w", err)
	}
	return v.([]entity.Task), nil

}

func (c *CacheTaskRepository) Update(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	updatedTask, err := c.taskRepo.Update(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("failed update task in pg: %w", err)
	}
	taskKey := c.taskKey(updatedTask.Id, updatedTask.UserId)
	allTasksKey := c.userTasksKey(updatedTask.UserId)
	c.redisClient.Del(ctx, taskKey, allTasksKey)

	return updatedTask, nil

}
func (c *CacheTaskRepository) Delete(ctx context.Context, id, userId int) error {
	err := c.taskRepo.Delete(ctx, id, userId)
	if err != nil {
		return fmt.Errorf("failed delete task in pg: %w", err)
	}
	taskKey := c.taskKey(id, userId)
	allTasksKey := c.userTasksKey(userId)

	c.redisClient.Del(ctx, taskKey, allTasksKey)
	return nil
}

func (c *CacheTaskRepository) CountTasksUser(ctx context.Context, userId int) (int, error) {
	count, err := c.taskRepo.CountTasksUser(ctx, userId)
	if err != nil {
		return 0, fmt.Errorf("failed get count tasks in pg: %w", err)
	}
	return count, nil
}
