package domain

type CreateUserDTO struct {
	Email        string `validate:"required,email"`
	PasswordHash string `validate:"required"`
}

type GetUserByEmailDTO struct {
	Email string `validate:"required,email"`
}
