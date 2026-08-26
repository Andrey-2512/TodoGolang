package repositories

import (
	"context"
	"encoding/json/v2"
	"fmt"
	"time"
	"todo/domain/entity"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

type taskRepository interface {
	CreateAndCheckLimit(ctx context.Context, t *entity.Task) (*entity.Task, error)
	GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error)
	UpdatePatch(ctx context.Context, t *entity.PatchTask) (*entity.Task, error)
	Delete(ctx context.Context, id, userId int) error
	CountTasksUser(ctx context.Context, userId int) (int, error)
	UpdatePut(ctx context.Context, t *entity.Task) (*entity.Task, error)
}
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

func NewCacheTaskRepository(taskRepo taskRepository, redisClient *redis.Client, tasksPrefix, userTasksPrefix string, cacheTaskTTL time.Duration) *CacheTaskRepository {
	return &CacheTaskRepository{taskRepo: taskRepo, redisClient: redisClient, taskPrefix: tasksPrefix, userTaskPrefix: userTasksPrefix, cacheTTL: cacheTaskTTL}
}

func (c *CacheTaskRepository) CreateAndCheckLimit(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	const op = "repositories.CacheTaskRepository.CreateAndCheckLimit"
	createdTask, err := c.taskRepo.CreateAndCheckLimit(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	allTasksKey := c.userTasksKey(createdTask.UserId)
	c.redisClient.Del(ctx, allTasksKey)

	return createdTask, nil

}

func (c *CacheTaskRepository) GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error) {
	const op = "repositories.CacheTaskRepository.GetUserTaskById"
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
		ctx := context.WithoutCancel(ctx)
		task, err := c.taskRepo.GetUserTaskById(ctx, id, userId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		data, err := json.Marshal(task)
		if err == nil {
			c.redisClient.Set(ctx, key, data, c.cacheTTL)
		}
		return task, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return v.(*entity.Task), nil
}

func (c *CacheTaskRepository) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	const op = "repositories.CacheTaskRepository.GetAllUserTasks"
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
		ctx := context.WithoutCancel(ctx)
		tasks, err := c.taskRepo.GetAllUserTasks(ctx, userId)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(tasks)
		if err == nil {
			c.redisClient.Set(ctx, key, data, c.cacheTTL)
		}

		return tasks, nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return v.([]entity.Task), nil

}

func (c *CacheTaskRepository) UpdatePatch(ctx context.Context, t *entity.PatchTask) (*entity.Task, error) {
	const op = "repositories.CacheTaskRepository.UpdatePatch"
	updatedTask, err := c.taskRepo.UpdatePatch(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	taskKey := c.taskKey(updatedTask.Id, updatedTask.UserId)
	allTasksKey := c.userTasksKey(updatedTask.UserId)
	c.redisClient.Del(ctx, taskKey, allTasksKey)

	return updatedTask, nil

}

func (c *CacheTaskRepository) UpdatePut(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	const op = "repositories.CacheTaskRepository.UpdatePut"
	updatedTask, err := c.taskRepo.UpdatePut(ctx, t)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	taskKey := c.taskKey(updatedTask.Id, updatedTask.UserId)
	allTasksKey := c.userTasksKey(updatedTask.UserId)
	c.redisClient.Del(ctx, taskKey, allTasksKey)

	return updatedTask, nil

}
func (c *CacheTaskRepository) Delete(ctx context.Context, id, userId int) error {
	const op = "repositories.CacheTaskRepository.Delete"
	err := c.taskRepo.Delete(ctx, id, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	taskKey := c.taskKey(id, userId)
	allTasksKey := c.userTasksKey(userId)

	c.redisClient.Del(ctx, taskKey, allTasksKey)
	return nil
}

func (c *CacheTaskRepository) CountTasksUser(ctx context.Context, userId int) (int, error) {
	const op = "repositories.CacheTaskRepository.CountTasksUser"
	count, err := c.taskRepo.CountTasksUser(ctx, userId)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return count, nil
}
