package services

import (
	"context"
	"encoding/json"
	"errors"

	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/ml"
	gen "guiltmachine/internal/proto/gen/ml"
	"guiltmachine/internal/repository"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type EntryService struct {
	repo       repository.EntriesRepository
	scoresRepo repository.ScoresRepository
	mlService  *ml.MLService
}

func NewEntryService(r repository.EntriesRepository) *EntryService {
	return &EntryService{repo: r}
}

func NewEntryServiceWithML(r repository.EntriesRepository, scoresRepo repository.ScoresRepository, mlService *ml.MLService) *EntryService {
	return &EntryService{
		repo:       r,
		scoresRepo: scoresRepo,
		mlService:  mlService,
	}
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

	// Process through ML layer if available
	if s.mlService != nil {
		resp, err := s.mlService.Roast(ctx, &gen.RoastRequest{
			EntryText:      text,
			UserId:         sid.String(),
			HumorIntensity: int32(5), // default intensity
			History:        []string{},
		})
		if err != nil {
			// Log error but don't fail entry creation
			// In production, this should be logged properly
			_ = err
		} else if s.scoresRepo != nil && resp != nil {
			// Store the guilt score from ML response
			score := int32(resp.GuiltScore * 100) // Convert to 0-100 scale
			meta := map[string]interface{}{
				"roast_text":   resp.RoastText,
				"tags":         resp.Tags,
				"safety_flags": resp.SafetyFlags,
			}
			metaJSON, _ := json.Marshal(meta)
			metaBytes := pqtype.NullRawMessage{
				RawMessage: metaJSON,
				Valid:      true,
			}
			_, _ = s.scoresRepo.CreateScore(ctx, sid, score, metaBytes)
		}
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
