package services

import (
	"context"
	"errors"

	"guiltmachine/internal/auth"
	cacheDomain "guiltmachine/internal/cache/domain"
	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"

	"github.com/google/uuid"
)

type SessionService struct {
	repo  repository.SessionsRepository
	cache *cacheDomain.SessionCache
	jwt   *auth.JWTManager
}

func NewSessionService(r repository.SessionsRepository, c *cacheDomain.SessionCache) *SessionService {
	return &SessionService{repo: r, cache: c}
}

// NewSessionServiceWithJWT creates a SessionService with JWT support
func NewSessionServiceWithJWT(r repository.SessionsRepository, c *cacheDomain.SessionCache, jwt *auth.JWTManager) *SessionService {
	return &SessionService{repo: r, cache: c, jwt: jwt}
}

// SetJWTManager sets the JWT manager (useful for testing or lazy initialization)
func (s *SessionService) SetJWTManager(jwt *auth.JWTManager) {
	s.jwt = jwt
}

// CreateSessionResult holds the result of creating a session
type CreateSessionResult struct {
	Session sqlc.GuiltSession
	JWT     string
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

// CreateSessionWithJWT creates a session and issues a JWT token
func (s *SessionService) CreateSessionWithJWT(ctx context.Context, userID string, notes *string) (CreateSessionResult, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return CreateSessionResult{}, errors.New("invalid user_id")
	}

	sess, err := s.repo.CreateSession(ctx, uid, notes)
	if err != nil {
		return CreateSessionResult{}, err
	}

	result := CreateSessionResult{Session: sess}

	// Issue JWT if manager is configured
	if s.jwt != nil {
		token, err := s.jwt.Issue(userID, sess.ID.String())
		if err != nil {
			return CreateSessionResult{}, err
		}
		result.JWT = token
	}

	return result, nil
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
