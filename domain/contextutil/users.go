package contextutil

import "context"

type contextKey string

const userIdKey contextKey = "user_id"

func SetUserIdInContext(ctx context.Context, userId int) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func GetUserIdFromContext(ctx context.Context) (int, bool) {
	userId, ok := ctx.Value(userIdKey).(int)
	return userId, ok
}
