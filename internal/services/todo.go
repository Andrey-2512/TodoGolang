package services

import (
	"context"
	"todo/domain/entity"
)

type taskService struct {
	repo entity.TaskRepository
}

func NewTaskService(repo entity.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

type TaskService interface {
	CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error)
	GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error)
	UpdateTask(ctx context.Context, task *entity.Task) (*entity.Task, error)
	DeleteTask(ctx context.Context, id, userId int) error
}

func (t *taskService) CreateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	return t.repo.Create(ctx, task)
}

func (t *taskService) GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error) {
	return t.repo.GetUserTaskById(ctx, taskId, userId)
}

func (t *taskService) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	return t.repo.GetAllUserTasks(ctx, userId)
}

func (t *taskService) UpdateTask(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	return t.repo.Update(ctx, task)
}

func (t *taskService) DeleteTask(ctx context.Context, id, userId int) error {
	return t.repo.Delete(ctx, id, userId)
}
