package delivery

import (
	"context"
	"net/http"
	"time"
	"todo/domain/contextutil"
	jsonrender "todo/internal/delivery/http/json/render"
	"todo/internal/services"
)

type ProfileHandler struct {
	profileService services.ProfileService
}

type Profile struct {
	Username     string `json:"username"`
	CurrentTasks int    `json:"current_tasks"`
	TasksLimit   int    `json:"tasks_limit"`
}

func NewProfileHandler(profileService services.ProfileService) *ProfileHandler {
	return &ProfileHandler{profileService: profileService}
}

func (p *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
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
