package apperrors

import "errors"

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)
