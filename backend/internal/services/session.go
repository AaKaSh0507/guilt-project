package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"
)

type SessionService struct {
	repo repository.SessionsRepository
}

func NewSessionService(r repository.SessionsRepository) *SessionService {
	return &SessionService{repo: r}
}

func (s *SessionService) CreateSession(ctx context.Context, userID string, notes *string) (sqlc.GuiltSession, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return sqlc.GuiltSession{}, errors.New("invalid user_id")
	}

	sess, err := s.repo.CreateSession(ctx, uid, notes)
	if err != nil {
		return sqlc.GuiltSession{}, err
	}

	return sess, nil
}

func (s *SessionService) EndSession(ctx context.Context, id string) (sqlc.GuiltSession, error) {
	sid, err := uuid.Parse(id)
	if err != nil {
		return sqlc.GuiltSession{}, errors.New("invalid id")
	}

	sess, err := s.repo.EndSession(ctx, sid)
	if err != nil {
		return sqlc.GuiltSession{}, err
	}

	return sess, nil
}

func (s *SessionService) GetSession(ctx context.Context, id string) (sqlc.GuiltSession, error) {
	sid, err := uuid.Parse(id)
	if err != nil {
		return sqlc.GuiltSession{}, errors.New("invalid id")
	}

	sess, err := s.repo.GetSessionByID(ctx, sid)
	if err != nil {
		return sqlc.GuiltSession{}, err
	}

	return sess, nil
}

func (s *SessionService) ListSessionsByUser(ctx context.Context, userID string, limit, offset int32) ([]sqlc.GuiltSession, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user_id")
	}

	sessions, err := s.repo.ListSessionsByUser(ctx, uid, limit, offset)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
