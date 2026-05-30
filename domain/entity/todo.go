package entity

import "context"

type Task struct {
	Id          int
	Title       *string
	Description *string
	UserId      int
}

type TaskRepository interface {
	Create(ctx context.Context, t *Task) (*Task, error)
	GetUserTaskById(ctx context.Context, id, userId int) (*Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]Task, error)
	Update(ctx context.Context, t *Task) (*Task, error)
	Delete(ctx context.Context, id, userId int) error
	CountTasksUser(ctx context.Context, userId int) (int, error)
}
