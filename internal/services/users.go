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
	blacklist entity.BlacklistRepository
}

type UsersService interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, userEntity *entity.User) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, string, error)
	RevokeToken(ctx context.Context, token string) error
}

func NewUsersService(usersRepo entity.UsersRepository, hasher auth.Hasher, jwt auth.JWTManager, blacklist entity.BlacklistRepository) UsersService {
	return &usersService{usersRepo: usersRepo, hasher: hasher, jwt: jwt, blacklist: blacklist}
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

func (u *usersService) Login(ctx context.Context, userEntity *entity.User) (string, string, error) {
	user, err := u.usersRepo.GetByUsername(ctx, userEntity.Username)
	if err != nil {
		return "", "", err
	}
	accessPassword, err := u.hasher.Verify(user.Password, userEntity.Password)
	if err != nil {
		return "", "", err
	}
	if !accessPassword {
		return "", "", apperrors.ErrInvalidCredentials
	}

	accessToken, err := u.jwt.CreateAccessToken(&entity.UserPayload{UserID: user.Id, Username: user.Username})
	if err != nil {
		return "", "", err
	}
	refreshToken, err := u.jwt.CreateRefreshToken(&entity.UserPayload{UserID: user.Id, Username: user.Username})
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

	if ttl := time.Until(claims.ExpiresAt.Time); ttl > 0 {
		err = u.blacklist.AddNX(ctx, claims.JTI, ttl)
		if err != nil {
			return "", "", err
		}
	}

	userId := claims.UserId
	user, err := u.usersRepo.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return "", "", apperrors.ErrInvalidToken
		}
		return "", "", err
	}

	access, err := u.jwt.CreateAccessToken(&entity.UserPayload{UserID: userId, Username: user.Username})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refresh, err := u.jwt.CreateRefreshToken(&entity.UserPayload{UserID: userId, Username: user.Username})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return access, refresh, nil

}

func (u *usersService) RevokeToken(ctx context.Context, token string) error {
	claims, err := u.jwt.ParseRefreshToken(token)
	if err != nil {
		return err
	}
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}
	err = u.blacklist.AddNX(ctx, claims.JTI, ttl)
	if err != nil {
		return err
	}
	return nil
}
