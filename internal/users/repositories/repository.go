package repositories

import (
	"context"
	"errors"
	"fmt"
	"todo/domain/apperrors"
	"todo/domain/entity"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UsersRepository struct {
	db *pgxpool.Pool
}

func NewUsersRepository(db *pgxpool.Pool) *UsersRepository {
	return &UsersRepository{db: db}
}

func (u *UsersRepository) Create(ctx context.Context, user *entity.User) (*entity.User, error) {
	query := "INSERT INTO users (username, hash_password) VALUES ($1, $2) RETURNING id"

	var id int

	err := u.db.QueryRow(ctx, query, user.Username, user.Password).Scan(&id)

	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, apperrors.ErrUserAlreadyExists
			}
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &entity.User{
		Id:       id,
		Username: user.Username,
		Password: user.Password,
	}, nil

}

func (u *UsersRepository) GetById(ctx context.Context, id int) (*entity.User, error) {
	var user entity.User
	query := "SELECT id, username, hash_password FROM users WHERE id = $1"

	err := u.db.QueryRow(ctx, query, id).Scan(&user.Id, &user.Username, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil

}

func (u *UsersRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	query := "SELECT id, username, hash_password FROM users WHERE username = $1"

	err := u.db.QueryRow(ctx, query, username).Scan(&user.Id, &user.Username, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil

}

func (u *UsersRepository) Exists(ctx context.Context, username string) (bool, error) {
	query := "SELECT EXISTS(SELECT FROM users WHERE username = $1)"
	var exists bool
	err := u.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if user exists: %w", err)
	}
	return exists, nil
}
