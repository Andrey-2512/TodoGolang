package entity

import (
	"context"
)

type User struct {
	Id       int
	Username string
	Password string
}

type UsersRepository interface {
	GetById(ctx context.Context, id int) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Exists(ctx context.Context, username string) (bool, error)
}
