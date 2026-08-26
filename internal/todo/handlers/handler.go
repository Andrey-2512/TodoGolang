package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/contextutil"
	"todo/internal/jsonrender"
	"todo/internal/optional"
	"todo/internal/optionalconv"
	"todo/internal/validation"
)

type taskService interface {
	CreateTaskAndCheckLimit(ctx context.Context, task *entity.Task) (*entity.Task, error)
	GetUserTaskById(ctx context.Context, taskId, userId int) (*entity.Task, error)
	GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error)
	UpdateTaskPatch(ctx context.Context, task *entity.PatchTask) (*entity.Task, error)
	DeleteTask(ctx context.Context, id, userId int) error
	UpdateTaskPut(ctx context.Context, task *entity.Task) (*entity.Task, error)
}

type TaskHandler struct {
	taskService    taskService
	handlerTimeout time.Duration
}
type taskCreateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type taskUpdatePatchRequest struct {
	Title       optional.Optional[string] `json:"title"`
	Description optional.Optional[string] `json:"description"`
}

type taskResponse struct {
	Id          int     `json:"id"`
	Title       *string `json:"title"`
	Description *string `json:"description,omitzero"`
}

type taskUpdatePutRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

func (r *taskCreateRequest) Validate() error {
	errs := make(validation.Errors)

	validation.Check[*string](errs, "title", r.Title,
		validation.RequiredString("Field title required"),
		validation.RuneLength(0, 250, "Field title must not be longer than 250 characters"),
	)

	validation.Check[*string](errs, "description", r.Description,
		validation.RuneLength(0, 2500, "Field description must not be longer than 2500 characters"),
	)

	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (r *taskUpdatePatchRequest) Validate() error {
	errs := make(validation.Errors)
	if r.Title.Set {
		if r.Title.Null {
			errs["title"] = "Field cannot be null"
		} else {
			validation.Check[*string](errs, "title", &r.Title.Val,
				validation.RequiredString("Field title required"),
				validation.RuneLength(0, 250, "Field title must not be longer than 250 characters"),
			)
		}

	}

	if r.Description.Set && !r.Description.Null {
		validation.Check[*string](errs, "description", &r.Description.Val,
			validation.RuneLength(0, 2500, "Field description must not be longer than 2500 characters"),
		)
	}

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func (r *taskUpdatePutRequest) Validate() error {
	errs := make(validation.Errors)

	validation.Check[*string](errs, "title", r.Title,
		validation.RequiredString("Field title required"),
		validation.RuneLength(0, 250, "Field title must not be longer than 250 characters"),
	)

	validation.Check[*string](errs, "description", r.Description,
		validation.RuneLength(0, 2500, "Field description must not be longer than 2500 characters"),
	)

	if len(errs) != 0 {
		return errs
	}

	return nil
}

func NewTaskHandler(taskService taskService, handlerTimeout time.Duration) *TaskHandler {
	return &TaskHandler{taskService: taskService, handlerTimeout: handlerTimeout}
}

func (h *TaskHandler) GetAllTasksHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
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

	resp := make([]taskResponse, len(tasks))

	for i, task := range tasks {
		resp[i] = taskResponse{
			Id:          task.Id,
			Title:       task.Title,
			Description: task.Description,
		}
	}

	jsonrender.JSONResponse(resp, w, http.StatusOK)

}

func (h *TaskHandler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
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

	jsonrender.JSONResponse(taskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)

}

func (h *TaskHandler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
	defer cancel()
	var t taskCreateRequest
	err := jsonrender.DecodeBody(w, r, &t)

	if err != nil {
		if bytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Body too large (max. %d KB)", bytesError.Limit/1024)}, w, http.StatusRequestEntityTooLarge)
			return
		}
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

	task, err := h.taskService.CreateTaskAndCheckLimit(ctx, &entity.Task{Title: t.Title, Description: t.Description, UserId: userId})
	if err != nil {
		if limitErr, ok := errors.AsType[*apperrors.ErrLimitTasksReached](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Sorry, you have reached your task limit (%d), please delete unnecessary tasks and try again", limitErr.TasksLimit)}, w, http.StatusUnprocessableEntity)
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

	jsonrender.JSONResponse(taskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusCreated)
}

func (h *TaskHandler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
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

func (h *TaskHandler) UpdateTaskPatch(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
	defer cancel()
	var t taskUpdatePatchRequest

	err := jsonrender.DecodeBody(w, r, &t)
	if err != nil {
		if bytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Body too large (max. %d KB)", bytesError.Limit/1024)}, w, http.StatusRequestEntityTooLarge)
			return
		}
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

	task, err := h.taskService.UpdateTaskPatch(ctx, &entity.PatchTask{
		Id:          id,
		Title:       optionalconv.FromJSONToEntity(t.Title),
		Description: optionalconv.FromJSONToEntity(t.Description),
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
	jsonrender.JSONResponse(taskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)
}

func (h *TaskHandler) UpdateTaskPut(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), h.handlerTimeout)
	defer cancel()
	var t taskUpdatePutRequest

	err := jsonrender.DecodeBody(w, r, &t)
	if err != nil {
		if bytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Body too large (max. %d KB)", bytesError.Limit/1024)}, w, http.StatusRequestEntityTooLarge)
			return
		}
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

	task, err := h.taskService.UpdateTaskPut(ctx, &entity.Task{
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

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to update"}, w, http.StatusInternalServerError)
		return

	}
	jsonrender.JSONResponse(taskResponse{Id: task.Id, Title: task.Title, Description: task.Description}, w, http.StatusOK)
}
