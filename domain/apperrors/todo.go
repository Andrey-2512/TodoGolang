package apperrors

import "errors"

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)

type ErrLimitTasksReached struct {
	TasksLimit int
}

func (l *ErrLimitTasksReached) Error() string {
	return "limit tasks reached"
}
