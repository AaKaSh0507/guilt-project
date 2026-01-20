package services

import (
	"context"
	"errors"
	"regexp"

	"github.com/google/uuid"
	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"
)

type UserService struct {
	repo repository.UsersRepository
}

func NewUserService(r repository.UsersRepository) *UserService {
	return &UserService{repo: r}
}

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func (s *UserService) CreateUser(ctx context.Context, email string, passwordHash string) (sqlc.User, error) {
	if !emailRegex.MatchString(email) {
		return sqlc.User{}, errors.New("invalid email format")
	}

	u, err := s.repo.CreateUser(ctx, email, passwordHash)
	if err != nil {
		return sqlc.User{}, err
	}

	return u, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (sqlc.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return sqlc.User{}, errors.New("invalid UUID")
	}

	u, err := s.repo.GetUserByID(ctx, userUUID)
	if err != nil {
		return sqlc.User{}, err
	}

	return u, nil
}
