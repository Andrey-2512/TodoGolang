package todo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/config"
	"todo/pkg/optional"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	db         *pgxpool.Pool
	limitTasks int
}

func NewTaskRepository(db *pgxpool.Pool, app config.AppConfig) *TaskRepository {
	return &TaskRepository{db: db, limitTasks: app.MaxTasksPerUser}
}

type UpdatePatchTaskInput struct {
	Id          int
	Title       optional.Optional[string]
	Description optional.Optional[string]
	UserId      int
}

func (r *TaskRepository) CreateAndCheckLimit(ctx context.Context, t *entity.Task) (*entity.Task, error) {

	tx, err := r.db.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed start transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var lockedUserId int
	err = tx.QueryRow(ctx, "SELECT id FROM users WHERE id = $1 FOR UPDATE", t.UserId).Scan(&lockedUserId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed block user: %w", err)
	}

	var count int
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM tasks WHERE user_id = $1", t.UserId).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("failed get count tasks: %w", err)
	}

	if count >= r.limitTasks {
		return nil, &apperrors.ErrLimitTasksReached{TasksLimit: r.limitTasks}
	}

	task := &entity.Task{}
	query := "INSERT INTO tasks (title, description, user_id) VALUES ($1, $2, $3) RETURNING id, title, description, user_id"

	err = tx.QueryRow(ctx, query, t.Title, t.Description, t.UserId).Scan(&task.Id, &task.Title, &task.Description, &task.UserId)

	if err != nil {
		return nil, fmt.Errorf("failed create task: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed commit transaction: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error) {
	var t entity.Task
	query := "SELECT id, title, description FROM tasks WHERE id = $1 AND user_id = $2"
	err := r.db.QueryRow(ctx, query, id, userId).Scan(&t.Id, &t.Title, &t.Description)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return &t, nil

}

func (r *TaskRepository) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	var listTask = make([]entity.Task, 0)

	query := "SELECT id, title, description FROM tasks WHERE user_id = $1 ORDER BY id DESC"
	rows, err := r.db.Query(ctx, query, userId)

	if err != nil {
		return nil, fmt.Errorf("failed to get all user tasks: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var t entity.Task
		err := rows.Scan(&t.Id, &t.Title, &t.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to get all user tasks: %w", err)
		}

		listTask = append(listTask, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get all user tasks: %w", err)
	}

	return listTask, nil
}

func (r *TaskRepository) UpdatePatch(ctx context.Context, t *UpdatePatchTaskInput) (*entity.Task, error) {
	var queryParts []string
	var args []any
	ArgID := 1

	if t.Description.Set {
		queryParts = append(queryParts, fmt.Sprintf("description = $%d", ArgID))
		ArgID++
		if t.Description.Null {
			args = append(args, nil)
		} else {
			args = append(args, t.Description.Val)
		}
	}

	if t.Title.Set {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", ArgID))
		args = append(args, t.Title.Val)
		ArgID++
	}

	if len(queryParts) <= 0 {
		return nil, apperrors.ErrNoFieldsToUpdate
	}

	args = append(args, t.Id)

	args = append(args, t.UserId)

	query := "UPDATE tasks SET " + strings.Join(queryParts, ", ") + fmt.Sprintf(" WHERE id = $%d AND user_id = $%d RETURNING id, title, description, user_id", ArgID, ArgID+1)

	var task entity.Task
	err := r.db.QueryRow(ctx, query, args...).Scan(&task.Id, &task.Title, &task.Description, &task.UserId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to update tasks: %w", err)
	}

	return &task, nil

}

func (r *TaskRepository) UpdatePut(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	query := "UPDATE tasks SET title = $1, description = $2 WHERE id = $3 AND user_id = $4 RETURNING id, title, description, user_id"

	var updatedTask entity.Task

	err := r.db.QueryRow(ctx, query, t.Title, t.Description, t.Id, t.UserId).Scan(&updatedTask.Id, &updatedTask.Title, &updatedTask.Description, &updatedTask.UserId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to update tasks: %w", err)
	}
	return &updatedTask, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id, userId int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userId)
	if err != nil {

		return fmt.Errorf("failed to delete user task: %w", err)
	}
	rows := result.RowsAffected()

	if rows == 0 {
		return apperrors.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) CountTasksUser(ctx context.Context, userId int) (int, error) {
	var count int
	query := "SELECT COUNT(id) FROM tasks WHERE user_id = $1"
	err := r.db.QueryRow(ctx, query, userId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get count tasks: %w", err)
	}
	return count, nil
}
