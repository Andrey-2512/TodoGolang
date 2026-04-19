package services

import (
	"context"
	"errors"
	"fmt"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/auth"
)

type usersService struct {
	usersRepo entity.UsersRepository
	hasher    auth.Hasher
	jwt       auth.JWTManager
	whitelist entity.WhitelistRepository
}

type UsersService interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, user *entity.User) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	RevokeToken(ctx context.Context, token string) error
}

func NewUsersService(usersRepo entity.UsersRepository, hasher auth.Hasher, jwt auth.JWTManager, whitelist entity.WhitelistRepository) UsersService {
	return &usersService{usersRepo: usersRepo, hasher: hasher, jwt: jwt, whitelist: whitelist}
}

func (u *usersService) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
	exists, err := u.usersRepo.Exists(ctx, user.Username)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperrors.ErrUserAlreadyExists
	}

	hashPassword, err := u.hasher.Hash(user.Password)

	if err != nil {
		return nil, err
	}

	return u.usersRepo.Create(ctx, &entity.User{Username: user.Username, Password: hashPassword})
}

func (u *usersService) Login(ctx context.Context, user *entity.User) (string, string, error) {
	userDB, err := u.usersRepo.GetByUsername(ctx, user.Username)
	if err != nil {
		return "", "", err
	}
	accessPassword, err := u.hasher.Verify(userDB.Password, user.Password)
	if err != nil {
		return "", "", err
	}
	if !accessPassword {
		return "", "", apperrors.ErrInvalidCredentials
	}

	accessToken, err := u.jwt.CreateAccessToken(&entity.UserPayload{UserID: userDB.Id, Username: userDB.Username})
	if err != nil {
		return "", "", err
	}
	refreshToken, err := u.jwt.CreateRefreshToken(&entity.UserPayload{UserID: userDB.Id, Username: userDB.Username})
	if err != nil {
		return "", "", err
	}

	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}
	err = u.whitelist.Add(ctx, claims.JTI, time.Until(claims.ExpiresAt.Time))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *usersService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	userId := claims.UserId
	user, err := u.usersRepo.GetById(ctx, userId)

	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return "", "", apperrors.ErrInvalidToken
		}
		return "", "", err
	}

	payload := entity.UserPayload{UserID: userId, Username: user.Username}

	access, err := u.jwt.CreateAccessToken(&payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, err := u.jwt.CreateRefreshToken(&payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	claimsNewRefresh, err := u.jwt.ParseRefreshToken(refresh)
	if err != nil {
		return "", "", err
	}

	ttl := time.Until(claimsNewRefresh.ExpiresAt.Time)

	err = u.whitelist.ConsumeAndAddToken(ctx, claims.JTI, claimsNewRefresh.JTI, ttl)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil

}

func (u *usersService) RevokeToken(ctx context.Context, token string) error {
	claims, err := u.jwt.ParseRefreshToken(token)
	if err != nil {
		return err
	}
	if ttl := time.Until(claims.ExpiresAt.Time); ttl <= 0 {
		return nil
	}
	err = u.whitelist.Del(ctx, claims.JTI)
	if err != nil {
		return err
	}
	return nil
}
