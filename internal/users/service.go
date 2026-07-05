package users

import (
	"context"
	"fmt"
	"todo/domain/entity"
	"todo/internal/config"
)

type taskRepository interface {
	CountTasksUser(ctx context.Context, userId int) (int, error)
}

type ProfileService struct {
	taskRepo   taskRepository
	tasksLimit int
}

func NewProfileService(taskRepo taskRepository, app config.AppConfig) *ProfileService {
	return &ProfileService{taskRepo: taskRepo, tasksLimit: app.MaxTasksPerUser}
}

func (p *ProfileService) GetProfile(ctx context.Context, userId int, username string) (*entity.UserProfile, error) {
	count, err := p.taskRepo.CountTasksUser(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("failed get profile: %w", err)
	}

	return &entity.UserProfile{Username: username, CurrentTasks: count, TasksLimit: p.tasksLimit}, nil
}
