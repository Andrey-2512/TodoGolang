package delivery

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/jsonutil"
	"todo/internal/services"
)

type UsersHandler struct {
	userService services.UsersService
}

func NewUsersHandler(service services.UsersService) *UsersHandler {
	return &UsersHandler{userService: service}
}

type userClaims struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var userData userClaims

	err := jsonutil.DecodeBody(w, r, &userData)
	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to login incorrect data"}, w, http.StatusBadRequest)
		return
	}

	user, err := u.userService.Register(ctx, &entity.User{Username: userData.Username, Password: userData.Password})
	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserAlreadyExists) {
			jsonutil.JSONResponse(map[string]any{"detail": "User already exists"}, w, http.StatusConflict)
			return
		}
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to register"}, w, http.StatusInternalServerError)
		return
	}
	jsonutil.JSONResponse(map[string]any{"detail": fmt.Sprintf("User created %s", user.Username)}, w, http.StatusCreated)
}

func (u *UsersHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var userData userClaims

	err := jsonutil.DecodeBody(w, r, &userData)
	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "Failed to login incorrect data"}, w, http.StatusBadRequest)
		return
	}

	accessToken, refreshToken, err := u.userService.Login(ctx, &entity.User{Username: userData.Username, Password: userData.Password})

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrUserNotFound) || errors.Is(err, apperrors.ErrInvalidCredentials) {
			jsonutil.JSONResponse(map[string]any{"detail": "Failed to login incorrect username or password"}, w, http.StatusUnauthorized)
			return
		}

		jsonutil.JSONResponse(map[string]any{"detail": "Failed to login"}, w, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "refresh_token", Value: refreshToken, Secure: false, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 60 * 60 * 24 * 7})

	jsonutil.JSONResponse(map[string]any{"access_token": accessToken}, w, http.StatusOK)
}

func (u *UsersHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	refreshTokenCookie, err := r.Cookie("refresh_token")

	if err != nil {
		jsonutil.JSONResponse(map[string]any{"detail": "You are not login"}, w, http.StatusUnauthorized)
		return
	}
	accessToken, err := u.userService.Refresh(ctx, refreshTokenCookie.Value)

	if err != nil {

		if errors.Is(err, context.DeadlineExceeded) {
			jsonutil.JSONResponse(map[string]any{"detail": "Request too long"}, w, http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, context.Canceled) {
			return
		}

		if errors.Is(err, apperrors.ErrSessionExpired) {
			jsonutil.JSONResponse(map[string]any{"detail": "Your session has been expired, please login again"}, w, http.StatusUnauthorized)
			return
		}
		if errors.Is(err, apperrors.ErrInvalidToken) {
			jsonutil.JSONResponse(map[string]any{"detail": "Invalid token"}, w, http.StatusUnauthorized)
			return
		}

		jsonutil.JSONResponse(map[string]any{"detail": "Failed to refresh"}, w, http.StatusInternalServerError)
		return
	}
	jsonutil.JSONResponse(map[string]any{"access_token": accessToken}, w, http.StatusOK)

}
