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
	const op = "repositories.UsersRepository.Create"
	query := "INSERT INTO users (username, hash_password) VALUES ($1, $2) RETURNING id"

	var id int

	err := u.db.QueryRow(ctx, query, user.Username, user.Password).Scan(&id)

	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, fmt.Errorf("%s: %w", op, apperrors.ErrUserAlreadyExists)
			}
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &entity.User{
		Id:       id,
		Username: user.Username,
		Password: user.Password,
	}, nil

}

func (u *UsersRepository) GetById(ctx context.Context, id int) (*entity.User, error) {
	const op = "repositories.UsersRepository.GetById"
	var user entity.User
	query := "SELECT id, username, hash_password FROM users WHERE id = $1"

	err := u.db.QueryRow(ctx, query, id).Scan(&user.Id, &user.Username, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil

}

func (u *UsersRepository) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	const op = "repositories.UsersRepository.GetByUsername"
	var user entity.User
	query := "SELECT id, username, hash_password FROM users WHERE username = $1"

	err := u.db.QueryRow(ctx, query, username).Scan(&user.Id, &user.Username, &user.Password)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, apperrors.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil

}
