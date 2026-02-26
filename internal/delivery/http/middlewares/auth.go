package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"todo/domain/apperrors"
	contextutil "todo/domain/contextutil"
	"todo/internal/auth"
	"todo/internal/jsonutil"
)

type Auth struct {
	jwtManager auth.JWTManager
}

func NewAuthMiddleware(jwtManager auth.JWTManager) *Auth {
	return &Auth{jwtManager: jwtManager}
}

func (m *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := strings.Fields(r.Header.Get("Authorization"))

		if len(fields) != 2 || !strings.EqualFold(fields[0], "bearer") {
			jsonutil.JSONResponse(map[string]any{"detail": "Unauthorized"}, w, http.StatusUnauthorized)
			return
		}

		claims, err := m.jwtManager.ParseAccessToken(fields[1])
		if err != nil {
			if errors.Is(err, apperrors.ErrSessionExpired) {
				jsonutil.JSONResponse(map[string]any{"detail": "Your session has been expired, please login again"}, w, http.StatusUnauthorized)
				return
			}

			if errors.Is(err, apperrors.ErrInvalidToken) || errors.Is(err, apperrors.ErrInvalidTokenType) {
				jsonutil.JSONResponse(map[string]any{"detail": "Invalid token"}, w, http.StatusUnauthorized)
				return
			}

			jsonutil.JSONResponse(map[string]any{"detail": "Failed to fetch"}, w, http.StatusInternalServerError)
			return
		}

		ctx := contextutil.SetUserIdInContext(r.Context(), claims.UserId)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
