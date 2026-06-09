package users

import (
	"context"
	"net/http"
	"time"
	"todo/domain/entity"
	"todo/pkg/contextutil"
	"todo/pkg/jsonrender"
)

type profileService interface {
	GetProfile(ctx context.Context, userId int, username string) (*entity.UserProfile, error)
}

type ProfileHandler struct {
	profileService profileService
	handlerTimeout time.Duration
}

type Profile struct {
	Username     string `json:"username"`
	CurrentTasks int    `json:"current_tasks"`
	TasksLimit   int    `json:"tasks_limit"`
}

func NewProfileHandler(profileService profileService, handlerTimeout time.Duration) *ProfileHandler {
	return &ProfileHandler{profileService: profileService, handlerTimeout: handlerTimeout}
}

func (p *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), p.handlerTimeout)
	defer cancel()
	userId, ok := contextutil.GetUserIdFromContext(ctx)

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	username, ok := contextutil.GetUsernameFromContext(ctx)

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	profile, err := p.profileService.GetProfile(ctx, userId, username)

	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to get profile"}, w, http.StatusInternalServerError)
		return
	}

	jsonrender.JSONResponse(Profile{Username: profile.Username, TasksLimit: profile.TasksLimit, CurrentTasks: profile.CurrentTasks}, w, http.StatusOK)
}
