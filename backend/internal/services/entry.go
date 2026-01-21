package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"
)

type EntryService struct {
	repo repository.EntriesRepository
}

func NewEntryService(r repository.EntriesRepository) *EntryService {
	return &EntryService{repo: r}
}

func (s *EntryService) CreateEntry(ctx context.Context, sessionID string, text string, level int32) (sqlc.GuiltEntry, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return sqlc.GuiltEntry{}, errors.New("invalid session_id")
	}

	e, err := s.repo.CreateEntry(ctx, sid, text, level)
	if err != nil {
		return sqlc.GuiltEntry{}, err
	}

	return e, nil
}

func (s *EntryService) ListEntries(ctx context.Context, sessionID string) ([]sqlc.GuiltEntry, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, errors.New("invalid session_id")
	}

	entries, err := s.repo.ListEntriesBySession(ctx, sid)
	if err != nil {
		return nil, err
	}

	return entries, nil
}
