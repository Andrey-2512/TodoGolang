package apperrors

import "errors"

var (
	ErrTokenAlreadyWhitelisted = errors.New("token already whitelisted")
	ErrTokenNotInWhitelist     = errors.New("token not in whitelisted")
)
