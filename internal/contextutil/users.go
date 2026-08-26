package contextutil

import "context"

type contextKey string

const userIdKey contextKey = "user_id"
const usernameKey contextKey = "username"

func SetUserIdInContext(ctx context.Context, userId int) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func SetUsernameInContext(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

func GetUserIdFromContext(ctx context.Context) (int, bool) {
	userId, ok := ctx.Value(userIdKey).(int)
	return userId, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}
