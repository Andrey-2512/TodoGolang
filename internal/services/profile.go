package services

import (
	"context"
	"fmt"
	"todo/domain/entity"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userId int, username string) (*entity.UserProfile, error)
}

type profileService struct {
	taskRepo   entity.TaskRepository
	tasksLimit int
}

func NewProfileService(taskRepo entity.TaskRepository, tasksLimit int) ProfileService {
	return &profileService{taskRepo: taskRepo, tasksLimit: tasksLimit}
}

func (p *profileService) GetProfile(ctx context.Context, userId int, username string) (*entity.UserProfile, error) {
	count, err := p.taskRepo.CountTasksUser(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("failed get profile: %w", err)
	}

	return &entity.UserProfile{Username: username, CurrentTasks: count, TasksLimit: p.tasksLimit}, nil
}
