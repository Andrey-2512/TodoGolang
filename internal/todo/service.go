package todo

import (
	"context"
	"fmt"
	"todo/domain/entity"
)

type TaskService struct {
	repo taskRepository
}

type taskRepository interface {
	CreateAndCheckLimit(ctx context.Context, t *entity.Task) (*entity.Task, error)
	GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error)
	Update(ctx context.Context, t *entity.Task) (*entity.Task, error)
	Delete(ctx context.Context, id, userId int) error
	CountTasksUser(ctx context.Context, userId int) (int, error)
}

func NewTaskService(repo taskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (t *TaskService) CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	createdTask, err := t.repo.CreateAndCheckLimit(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed create task: %w", err)
	}
	return createdTask, nil
}

func (t *TaskService) GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error) {
	task, err := t.repo.GetUserTaskById(ctx, taskId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get user task by id: %w", err)
	}
	return task, nil
}

func (t *TaskService) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	tasks, err := t.repo.GetAllUserTasks(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("failed get all user tasks by user id: %w", err)
	}
	return tasks, nil
}

func (t *TaskService) UpdateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	updatedTask, err := t.repo.Update(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed update task: %w", err)
	}
	return updatedTask, nil
}

func (t *TaskService) DeleteTask(ctx context.Context, id, userId int) error {
	err := t.repo.Delete(ctx, id, userId)
	if err != nil {
		return fmt.Errorf("failed delete task: %w", err)
	}
	return nil
}
