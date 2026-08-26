package services

import (
	"context"
	"errors"
	"fmt"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/security"
)

type AuthService struct {
	usersRepo usersRepository
	hasher    hasher
	jwt       jwtManager
	whitelist whitelistRepository
}

type jwtManager interface {
	CreateAccessToken(user *entity.UserPayload) (string, error)
	ParseAccessToken(jwtToken string) (*security.UserClaims, error)
	CreateRefreshToken(user *entity.UserPayload) (string, error)
	ParseRefreshToken(jwtToken string) (*security.UserClaims, error)
}
type usersRepository interface {
	GetById(ctx context.Context, id int) (*entity.User, error)
	GetByUsername(ctx context.Context, username string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) (*entity.User, error)
}

type whitelistRepository interface {
	ConsumeAndAddToken(ctx context.Context, jti, newJti string, exp time.Duration) error
	Del(ctx context.Context, jti string) error
	Add(ctx context.Context, jti string, exp time.Duration) error
}

type hasher interface {
	Hash(password string) (string, error)
	Verify(hashPassword string, password string) (bool, error)
}

func NewAuthService(usersRepo usersRepository, hasher hasher, jwt jwtManager, whitelist whitelistRepository) *AuthService {
	return &AuthService{usersRepo: usersRepo, hasher: hasher, jwt: jwt, whitelist: whitelist}
}

func (u *AuthService) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "services.AuthService.Register"
	hashPassword, err := u.hasher.Hash(user.Password)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	created, err := u.usersRepo.Create(ctx, &entity.User{Username: user.Username, Password: hashPassword})

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return created, nil
}

func (u *AuthService) Login(ctx context.Context, user *entity.User) (string, string, error) {
	const op = "services.AuthService.Login"
	userDB, err := u.usersRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	accessPassword, err := u.hasher.Verify(userDB.Password, user.Password)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	if !accessPassword {
		return "", "", fmt.Errorf("%s: %w", op, apperrors.ErrInvalidAuthCredentials)
	}

	accessToken, err := u.jwt.CreateAccessToken(&entity.UserPayload{UserID: userDB.Id, Username: userDB.Username})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	refreshToken, err := u.jwt.CreateRefreshToken(&entity.UserPayload{UserID: userDB.Id, Username: userDB.Username})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	err = u.whitelist.Add(ctx, claims.JTI, time.Until(claims.ExpiresAt.Time))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (u *AuthService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	const op = "services.AuthService.Refresh"
	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	userId := claims.UserId
	user, err := u.usersRepo.GetById(ctx, userId)

	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return "", "", fmt.Errorf("%s: %w", op, apperrors.ErrInvalidToken)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	payload := entity.UserPayload{UserID: userId, Username: user.Username}

	access, err := u.jwt.CreateAccessToken(&payload)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refresh, err := u.jwt.CreateRefreshToken(&payload)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	claimsNewRefresh, err := u.jwt.ParseRefreshToken(refresh)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	ttl := time.Until(claimsNewRefresh.ExpiresAt.Time)

	err = u.whitelist.ConsumeAndAddToken(ctx, claims.JTI, claimsNewRefresh.JTI, ttl)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return access, refresh, nil

}

func (u *AuthService) RevokeToken(ctx context.Context, token string) error {
	const op = "services.AuthService.RevokeToken"
	claims, err := u.jwt.ParseRefreshToken(token)
	if err != nil {
		if errors.Is(err, apperrors.ErrSessionExpired) {
			return nil
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	err = u.whitelist.Del(ctx, claims.JTI)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
