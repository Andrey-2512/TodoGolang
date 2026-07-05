package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/config"
	"todo/pkg/jsonrender"
	"todo/pkg/validation"
)

type AuthHandler struct {
	userService    authService
	handlerTimeout time.Duration
	cookieSecure   bool
	refreshTTL     time.Duration
}

type authService interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, user *entity.User) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	RevokeToken(ctx context.Context, token string) error
}

func NewAuthHandler(service authService, jwt config.JWTConfig, server config.HTTPServer) *AuthHandler {
	return &AuthHandler{userService: service, handlerTimeout: server.HandlerTimeout, cookieSecure: server.CookieSecure, refreshTTL: jwt.RefreshTTL}
}

type userRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var regexpValidate = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func (r *userRequest) Validate() error {
	errs := make(validation.Errors)

	validation.Check[*string](errs, "username", &r.Username,
		validation.RequiredString("Check username is required"),
		validation.RuneLength(8, 64, "Username must be between 8 and 64 characters long"),
		validation.Match(regexpValidate, "In field username only Latin letters, numbers, underscores are allowed"),
	)

	validation.Check[*string](errs, "password", &r.Password,
		validation.RequiredString("Check password is required"),
		validation.RuneLength(8, 64, "Password must be between 8 and 64 characters long"),
	)
	if len(errs) != 0 {
		return errs
	}
	return nil
}

func (u *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), u.handlerTimeout)
	defer cancel()
	var userData userRequest

	err := jsonrender.DecodeBody(w, r, &userData)
	if err != nil {
		if bytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Body too large (max. %d KB)", bytesError.Limit/1024)}, w, http.StatusRequestEntityTooLarge)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect json data"}, w, http.StatusBadRequest)
		return
	}

	err = userData.Validate()

	if err != nil {
		if errorsValidation, ok := errors.AsType[validation.Errors](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": errorsValidation}, w, http.StatusUnprocessableEntity)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to process request"}, w, http.StatusInternalServerError)
		return

	}

	user, err := u.userService.Register(ctx, &entity.User{Username: userData.Username, Password: userData.Password})
	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			jsonrender.JSONResponse(map[string]any{"detail": "User already exists"}, w, http.StatusConflict)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to register"}, w, http.StatusInternalServerError)
		return
	}
	jsonrender.JSONResponse(map[string]any{"detail": user.Username}, w, http.StatusCreated)
}

func (u *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), u.handlerTimeout)
	defer cancel()

	var userData userRequest

	err := jsonrender.DecodeBody(w, r, &userData)
	if err != nil {
		if bytesError, ok := errors.AsType[*http.MaxBytesError](err); ok {
			jsonrender.JSONResponse(map[string]any{"detail": fmt.Sprintf("Body too large (max. %d KB)", bytesError.Limit/1024)}, w, http.StatusRequestEntityTooLarge)
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Incorrect json data"}, w, http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := u.userService.Login(ctx, &entity.User{Username: userData.Username, Password: userData.Password})

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrInvalidAuthCredentials) {
			jsonrender.JSONResponse(map[string]any{"detail": "Failed to login incorrect username or password"}, w, http.StatusUnauthorized)
			return
		}

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to login"}, w, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: refreshToken, Secure: u.cookieSecure, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: int(u.refreshTTL.Seconds())})

	jsonrender.JSONResponse(map[string]any{"access_token": accessToken}, w, http.StatusOK)
}

func (u *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), u.handlerTimeout)
	defer cancel()

	refreshTokenCookie, err := r.Cookie("refresh_token")

	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "You are not login"}, w, http.StatusUnauthorized)
		return
	}
	accessToken, refreshToken, err := u.userService.Refresh(ctx, refreshTokenCookie.Value)

	if err != nil {
		if errors.Is(err, apperrors.ErrSessionExpired) {
			jsonrender.JSONResponse(map[string]any{"detail": "Your session has been expired, please login again"}, w, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, apperrors.ErrInvalidToken) || errors.Is(err, apperrors.ErrTokenAlreadyWhitelisted) || errors.Is(err, apperrors.ErrTokenNotInWhitelist) || errors.Is(err, apperrors.ErrInvalidTokenType) {
			jsonrender.JSONResponse(map[string]any{"detail": "Invalid token"}, w, http.StatusUnauthorized)
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		jsonrender.JSONResponse(map[string]any{"detail": "Failed to refresh"}, w, http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name: "refresh_token", Value: refreshToken,
		Secure:   u.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   int(u.refreshTTL.Seconds()),
		SameSite: http.SameSiteLaxMode,
	},
	)

	jsonrender.JSONResponse(map[string]any{"access_token": accessToken}, w, http.StatusOK)

}

func (u *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), u.handlerTimeout)
	defer cancel()
	refreshTokenCookie, err := r.Cookie("refresh_token")
	if err != nil {
		jsonrender.JSONResponse(map[string]any{"detail": "You are not login"}, w, http.StatusUnauthorized)
		return
	}
	err = u.userService.RevokeToken(ctx, refreshTokenCookie.Value)
	if err != nil {
		if errors.Is(err, apperrors.ErrTokenAlreadyWhitelisted) || errors.Is(err, apperrors.ErrInvalidToken) || errors.Is(err, apperrors.ErrInvalidTokenType) {
			http.SetCookie(w, &http.Cookie{
				Name:     "refresh_token",
				Value:    "",
				Secure:   u.cookieSecure,
				HttpOnly: true,
				Path:     "/",
				MaxAge:   -1,
			})
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if errors.Is(err, context.DeadlineExceeded) {
			jsonrender.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}
		jsonrender.JSONResponse(map[string]any{"detail": "Failed to logout"}, w, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Secure:   u.cookieSecure,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusNoContent)

}
