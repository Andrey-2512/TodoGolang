package services

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
	UpdatePatch(ctx context.Context, t *entity.PatchTask) (*entity.Task, error)
	Delete(ctx context.Context, id, userId int) error
	CountTasksUser(ctx context.Context, userId int) (int, error)
	UpdatePut(ctx context.Context, t *entity.Task) (*entity.Task, error)
}

func NewTaskService(repo taskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (t *TaskService) CreateTaskAndCheckLimit(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	const op = "services.TaskService.CreateTaskAndCheckLimit"
	createdTask, err := t.repo.CreateAndCheckLimit(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return createdTask, nil
}

func (t *TaskService) GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error) {
	const op = "services.TaskService.GetUserTaskById"
	task, err := t.repo.GetUserTaskById(ctx, taskId, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return task, nil
}

func (t *TaskService) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	const op = "services.TaskService.GetAllUserTasks"
	tasks, err := t.repo.GetAllUserTasks(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return tasks, nil
}

func (t *TaskService) UpdateTaskPatch(ctx context.Context, task *entity.PatchTask) (*entity.Task, error) {
	const op = "services.TaskService.UpdateTaskPatch"
	updatedTask, err := t.repo.UpdatePatch(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return updatedTask, nil
}

func (t *TaskService) UpdateTaskPut(ctx context.Context, task *entity.Task) (*entity.Task, error) {
	const op = "services.TaskService.UpdateTaskPut"
	updatedTask, err := t.repo.UpdatePut(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return updatedTask, nil
}

func (t *TaskService) DeleteTask(ctx context.Context, id, userId int) error {
	const op = "services.TaskService.DeleteTask"
	err := t.repo.Delete(ctx, id, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
