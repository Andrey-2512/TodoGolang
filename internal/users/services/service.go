package services

import (
	"context"
	"fmt"
	"todo/domain/entity"
)

type taskRepository interface {
	CountTasksUser(ctx context.Context, userId int) (int, error)
}

type ProfileService struct {
	taskRepo   taskRepository
	tasksLimit int
}

func NewProfileService(taskRepo taskRepository, maxTasksPerUser int) *ProfileService {
	return &ProfileService{taskRepo: taskRepo, tasksLimit: maxTasksPerUser}
}

func (p *ProfileService) GetProfile(ctx context.Context, userId int, username string) (*entity.UserProfile, error) {
	const op = "services.ProfileService.GetProfile"
	count, err := p.taskRepo.CountTasksUser(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.UserProfile{Username: username, CurrentTasks: count, TasksLimit: p.tasksLimit}, nil
}
