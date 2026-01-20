package domain

import "time"

type User struct {
	ID        string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CreateUserDTO struct {
	Email        string `validate:"required,email"`
	PasswordHash string `validate:"required"`
}

type GetUserByIDDTO struct {
	ID string `validate:"required"`
}
