package apperrors

import "errors"

var (
	ErrTokenNotInWhitelist = errors.New("token not in whitelisted")
)
