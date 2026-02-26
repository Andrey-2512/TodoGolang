package services

import (
	"context"
	"errors"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"todo/internal/auth"
)

type UserRegisterRequest struct {
	Id           int    `json:"id"`
	Username     string `json:"username"`
	HashPassword string `json:"hash_password"`
}

type usersService struct {
	usersRepo entity.UsersRepository
	hasher    auth.Hasher
	jwt       auth.JWTManager
}

type UsersService interface {
	Register(ctx context.Context, user *entity.User) (*entity.User, error)
	Login(ctx context.Context, userEntity *entity.User) (string, string, error)
	Refresh(ctx context.Context, refreshToken string) (string, error)
}

func NewUsersService(usersRepo entity.UsersRepository, hasher auth.Hasher, jwt auth.JWTManager) UsersService {
	return &usersService{usersRepo: usersRepo, hasher: hasher, jwt: jwt}
}

func (u *usersService) Register(ctx context.Context, user *entity.User) (*entity.User, error) {
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

	accessToken, err := u.jwt.CreateAccessToken(entity.UserPayload{UserID: user.Id, Username: user.Username})
	if err != nil {
		return "", "", err
	}
	refreshToken, err := u.jwt.CreateRefreshToken(entity.UserPayload{UserID: user.Id, Username: user.Username})
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (u *usersService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	claims, err := u.jwt.ParseRefreshToken(refreshToken)
	if err != nil {

		return "", err
	}
	userId := claims.UserId
	user, err := u.usersRepo.GetById(ctx, userId)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return "", apperrors.ErrInvalidToken
		}
		return "", err
	}
	if user == nil {
		return "", apperrors.ErrInvalidToken
	}
	return u.jwt.CreateAccessToken(entity.UserPayload{UserID: userId, Username: user.Username})

}
