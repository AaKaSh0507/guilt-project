package services

import (
	"context"
	"encoding/json"
	"errors"

	cacheDomain "guiltmachine/internal/cache/domain"
	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"

	"github.com/google/uuid"
)

type PreferencesService struct {
	repo  repository.PreferencesRepository
	cache *cacheDomain.PreferencesCache
}

func NewPreferencesService(r repository.PreferencesRepository, c *cacheDomain.PreferencesCache) *PreferencesService {
	return &PreferencesService{repo: r, cache: c}
}

func (s *PreferencesService) UpsertPreferences(ctx context.Context, userID string, theme *string, notifications bool, metadata string) (sqlc.UserPreference, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return sqlc.UserPreference{}, errors.New("invalid user_id")
	}

	var decoded any
	if metadata != "" {
		_ = json.Unmarshal([]byte(metadata), &decoded)
	}

	pref, err := s.repo.UpsertPreferences(ctx, uid, theme, notifications, decoded)
	if err != nil {
		return sqlc.UserPreference{}, err
	}

	return pref, nil
}

func (s *PreferencesService) GetPreferences(ctx context.Context, userID string) (sqlc.UserPreference, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return sqlc.UserPreference{}, errors.New("invalid user_id")
	}

	pref, err := s.repo.GetPreferencesByUserID(ctx, uid)
	if err != nil {
		return sqlc.UserPreference{}, err
	}

	return pref, nil
}
