package delivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todo/domain/apperrors"
	"todo/domain/contextutil"
	"todo/domain/entity"
	"todo/internal/delivery/http/json/render"
	"todo/internal/services"

	"github.com/go-ozzo/ozzo-validation/v4"
)

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {

	return &TaskHandler{taskService: taskService}
}

type TaskCreateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type TaskUpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (r *TaskCreateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Title,
			validation.Required.Error("Field title required"),
			validation.RuneLength(0, 250).Error("Field title must not be longer than 250 characters")),
		validation.Field(&r.Description,
			validation.RuneLength(0, 2500).Error("Field description must not be longer than 2500 characters")),
	)
}

func (r *TaskUpdateRequest) Validate() error {
	return validation.ValidateStruct(r,
		validation.Field(&r.Title, validation.When(r.Title != nil,
			validation.Required.Error("Field title cannot be empty string")),
			validation.RuneLength(0, 250).Error("Field title must not be longer than 250 characters")),
		validation.Field(&r.Description, validation.RuneLength(0, 2500).Error("Field description must not be longer than 2500 characters")),
	)
}

type TaskResponse struct {
	Id          int     `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description,omitzero"`
}

func (h *TaskHandler) GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	tasks, err := h.taskService.GetAllUserTasks(ctx, userId)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to get tasks."}, w, http.StatusInternalServerError)
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

	jsonrender.JSONResponse(resp, w, http.StatusOK)

}

func (h *TaskHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	taskId, err := strconv.Atoi(r.PathValue("task_id"))

	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			jsonrender.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return
		}

		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return

	}

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	task, err := h.taskService.GetUserTaskById(ctx, taskId, userId)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonrender.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return
		}

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to get tasks"}, w, http.StatusInternalServerError)
		return

	}

	jsonrender.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)

}

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var t TaskCreateRequest
	err := jsonrender.DecodeBody(w, r, &t)

	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect json data"}, w, http.StatusBadRequest)
		return
	}

	err = t.Validate()

	if err != nil {
		if errorsValidation, ok := errors.AsType[validation.Errors](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": errorsValidation}, w, http.StatusUnprocessableEntity)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to process request"}, w, http.StatusInternalServerError)
		return
	}

	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	task, err := h.taskService.CreateTask(ctx, &entity.Task{Title: t.Title, Description: t.Description, UserId: userId})
	if err != nil {
		var limitErr *apperrors.ErrLimitTasksReached
		if ok := errors.As(err, &limitErr); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Sorry, you reach limit tasks (%d), please delete dont need tasks and try again", limitErr.TasksLimit)}, w, http.StatusBadRequest)
			return

		}

		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserNotFound) {
			jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
			return
		}

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to create task"}, w, http.StatusInternalServerError)
		return
	}

	jsonrender.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusCreated)
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	id, err := strconv.Atoi(r.PathValue("task_id"))

	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return
	}
	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
		return
	}

	err = h.taskService.DeleteTask(ctx, id, userId)

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonrender.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return

		}

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to delete task"}, w, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var t TaskUpdateRequest

	err := jsonrender.DecodeBody(w, r, &t)
	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect json data"}, w, http.StatusBadRequest)
		return
	}

	err = t.Validate()

	if err != nil {
		if errorsValidation, ok := errors.AsType[validation.Errors](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": errorsValidation}, w, http.StatusUnprocessableEntity)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to process request"}, w, http.StatusInternalServerError)
		return
	}

	id, err := strconv.Atoi(r.PathValue("task_id"))
	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect task id"}, w, http.StatusBadRequest)
		return
	}
	userId, ok := contextutil.GetUserIdFromContext(r.Context())

	if !ok {
		jsonrender.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
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
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrTaskNotFound) {
			jsonrender.JSONResponse(map[string]any{"detail": "Task not found"}, w, http.StatusNotFound)
			return

		}
		if errors.Is(err, apperrors.ErrNoFieldsToUpdate) {
			jsonrender.JSONResponse(map[string]any{"detail": "No fields to update"}, w, http.StatusBadRequest)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to update"}, w, http.StatusInternalServerError)
		return

	}
	jsonrender.JSONResponse(TaskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)
}
