package delivery

import (
	"context"
	"errors"
	"todo/domain/apperrors"
	"todo/domain/contextutil"
	"todo/domain/entity"
	"todo/internal/services"

	"net/http"
	"strconv"
	"time"
	"todo/internal/jsonutil"
)

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {

	return &TaskHandler{taskService: taskService}
}

type TaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type TaskResponse struct {
	Id          int     `json:"id"`
	Title       *string `json:"title,omitzero"`
	Description *string `json:"description,omitzero"`
}

func (h *TaskHandler) GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	tasks, err := h.taskService.GetAllUserTasks(ctx, userId)

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to get tasks."}, w, http.StatusInternalServerError)
		return
	}

	resp := make([]TaskResponse, len(tasks))

	for i, task := range tasks {
		resp[i] = TaskResponse{
			Id:          task.Id,
			Title:       task.Title,
			Description: task.Description,
		}
	}

	jsonutil.JSONResponse(resp, w, http.StatusOK)

}

func (h *TaskHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	taskId, err := strconv.Atoi(r.PathValue("task_id"))

	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			jsonutil.JSONResponse(map[string]any{"detail": "Id too big"}, w, http.StatusBadRequest)
			return
		}

		jsonutil.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return

	}

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	task, err := h.taskService.GetUserTaskById(ctx, taskId, userId)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonutil.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return
		}

		jsonutil.JSONResponse(map[string]any{"detail": "Failed to get tasks"}, w, http.StatusInternalServerError)
		return

	}

	jsonutil.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)

}

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var t TaskRequest
	err := jsonutil.DecodeBody(w, r, &t)

	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Incorrect data"}, w, http.StatusBadRequest)
		return
	}

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	task, err := h.taskService.CreateTask(ctx, &entity.Task{Title: t.Title, Description: t.Description, UserId: userId})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserNotFound) {
			jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
			return
		}
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to create task"}, w, http.StatusInternalServerError)
		return
	}

	jsonutil.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusCreated)
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	id, err := strconv.Atoi(r.PathValue("task_id"))

	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return
	}
	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	err = h.taskService.DeleteTask(ctx, id, userId)

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonutil.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return

		}

		jsonutil.JSONResponse(map[string]any{"detail": "Failed to delete task"}, w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var t TaskRequest
	err := jsonutil.DecodeBody(w, r, &t)
	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Incorrect data"}, w, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(r.PathValue("task_id"))
	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return
	}
	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	task, err := h.taskService.UpdateTask(ctx, &entity.Task{
		Id:          id,
		Title:       t.Title,
		Description: t.Description,
		UserId:      userId,
	})

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonutil.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return

		}
		if errors.Is(err, apperrors.ErrNoFieldsToUpdate) {
			jsonutil.JSONResponse(map[string]any{"detail": "No fields to update"}, w, http.StatusBadRequest)
			return
		}
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to update"}, w, http.StatusInternalServerError)
		return

	}
	jsonutil.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)
}
