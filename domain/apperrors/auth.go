package apperrors

import "errors"

var (
	ErrSessionExpired   = errors.New("session expired")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenType = errors.New("invalid token type")
)
