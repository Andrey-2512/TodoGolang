package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"todo/domain/apperrors"
	"todo/domain/entity"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type taskRepository struct {
	db *pgxpool.Pool
}

func NewTaskRepository(db *pgxpool.Pool) entity.TaskRepository {

	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, t *entity.Task) (*entity.Task, error) {

	task := &entity.Task{}
	query := "INSERT INTO tasks (title, description, user_id) VALUES ($1, $2, $3) RETURNING id, title, description, user_id"

	err := r.db.QueryRow(ctx, query, t.Title, t.Description, t.UserId).Scan(&task.Id, &task.Title, &task.Description, &task.UserId)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return nil, apperrors.ErrUserNotFound
			}
		}
		return nil, fmt.Errorf("failed to create task: %w", err)
	}
	return task, nil
}

func (r *taskRepository) GetUserTaskById(ctx context.Context, id, userId int) (*entity.Task, error) {
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

func (r *taskRepository) GetAllUserTasks(ctx context.Context, userId int) ([]entity.Task, error) {
	var listTask = make([]entity.Task, 0)

	query := "SELECT id, title, description FROM tasks WHERE user_id = $1"
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

func (r *taskRepository) Update(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	var queryParts []string
	var args []any
	ArgID := 1

	if t.Description != nil {
		queryParts = append(queryParts, fmt.Sprintf("description = $%d", ArgID))
		args = append(args, *t.Description)
		ArgID++
	}

	if t.Title != nil {
		queryParts = append(queryParts, fmt.Sprintf("title = $%d", ArgID))
		args = append(args, *t.Title)
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
func (r *taskRepository) Delete(ctx context.Context, id, userId int) error {
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
