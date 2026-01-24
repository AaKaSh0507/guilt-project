package services

import (
	"context"
	"encoding/json"
	"errors"

	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"

	"github.com/google/uuid"
)

type ScoreService struct {
	repo repository.ScoresRepository
}

func NewScoreService(r repository.ScoresRepository) *ScoreService {
	return &ScoreService{repo: r}
}

func (s *ScoreService) CreateScore(ctx context.Context, sessionID string, entryID *string, score int32, meta string) (sqlc.GuiltScore, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return sqlc.GuiltScore{}, errors.New("invalid session_id")
	}

	var eid *uuid.UUID
	if entryID != nil && *entryID != "" {
		parsed, err := uuid.Parse(*entryID)
		if err != nil {
			return sqlc.GuiltScore{}, errors.New("invalid entry_id")
		}
		eid = &parsed
	}

	var decoded any
	if meta != "" {
		_ = json.Unmarshal([]byte(meta), &decoded)
	}

	sc, err := s.repo.CreateScore(ctx, sid, eid, score, decoded)
	if err != nil {
		return sqlc.GuiltScore{}, err
	}

	return sc, nil
}

func (s *ScoreService) GetScore(ctx context.Context, sessionID string) (sqlc.GuiltScore, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return sqlc.GuiltScore{}, errors.New("invalid session_id")
	}

	sc, err := s.repo.GetScoreBySession(ctx, sid)
	if err != nil {
		return sqlc.GuiltScore{}, err
	}

	return sc, nil
}
