package services

import (
	"context"
	"fmt"
	"todo/domain/apperrors"
	"todo/domain/entity"
)

type taskService struct {
	repo     entity.TaskRepository
	maxTasks int
}

func NewTaskService(repo entity.TaskRepository, maxTasks int) TaskService {
	return &taskService{repo: repo, maxTasks: maxTasks}
}

type TaskService interface {
	CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error)
	GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error)
	UpdateTask(ctx context.Context, task *entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, id, userId int) error
}

func (t *taskService) CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	countTasks, err := t.repo.CountTasksUser(ctx, task.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to check count tasks: %w", err)
	}
	if countTasks >= t.maxTasks {
		return nil, apperrors.ErrLimitTasksReached
	}
	createdTask, err := t.repo.Create(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed create task: %w", err)
	}
	return createdTask, nil
}

func (t *taskService) GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error) {
	task, err := t.repo.GetUserTaskById(ctx, taskId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get user task by id: %w", err)
	}
	return task, nil
}

func (t *taskService) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	tasks, err := t.repo.GetAllUserTasks(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get all user tasks by user id: %w", err)
	}
	return tasks, nil
}

func (t *taskService) UpdateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	updatedTask, err := t.repo.Update(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed update task: %w", err)
	}
	return updatedTask, nil
}

func (t *taskService) DeleteTask(ctx context.Context, id, userId int) error {
	err := t.repo.Delete(ctx, id, userId)
	if err != nil {
		return fmt.Errorf("failed delete task: %w", err)
	}
	return nil
}
