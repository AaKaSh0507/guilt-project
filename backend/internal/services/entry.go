package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/ml"
	"guiltmachine/internal/queue"
	"guiltmachine/internal/repository"

	"github.com/google/uuid"
)

type EntryService struct {
	repo         repository.EntriesRepository
	scoresRepo   repository.ScoresRepository
	orchestrator *ml.HybridOrchestrator
	prefsService *PreferencesService
	queue        *queue.Producer
}

func NewEntryService(r repository.EntriesRepository) *EntryService {
	return &EntryService{repo: r}
}

func NewEntryServiceWithHybrid(r repository.EntriesRepository, scoresRepo repository.ScoresRepository, orchestrator *ml.HybridOrchestrator, prefsService *PreferencesService) *EntryService {
	return &EntryService{
		repo:         r,
		scoresRepo:   scoresRepo,
		orchestrator: orchestrator,
		prefsService: prefsService,
	}
}

func NewEntryServiceWithQueue(r repository.EntriesRepository, scoresRepo repository.ScoresRepository, mlService *ml.MLService, prefsService *PreferencesService, producer *queue.Producer) *EntryService {
	return &EntryService{
		repo:         r,
		scoresRepo:   scoresRepo,
		prefsService: prefsService,
		queue:        producer,
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

	// If queue available, enqueue ML job asynchronously
	if s.queue != nil {
		job := queue.EntryMLJob{
			EntryID:   e.ID.String(),
			UserID:    sid.String(),
			Text:      text,
			Persona:   "roast",
			Intensity: 3,
		}
		_ = s.queue.Enqueue(ctx, job)
		_ = s.repo.UpdateEntryStatus(ctx, e.ID, "pending")
	} else if s.orchestrator != nil {
		// Fallback to synchronous processing
		// Default intensity and persona
		intensity := 5
		persona := ml.PersonaRoast

		// Try to get user preferences if available
		if s.prefsService != nil {
			// Get the user_id from the session
			// Note: This is a simplified approach; you may need to fetch user_id from session first
			// For now, using default values
		}

		// Run hybrid orchestrator
		output, err := s.orchestrator.Run(ctx, ml.HybridInput{
			Text:      text,
			UserID:    sid.String(),
			Intensity: intensity,
			Persona:   persona,
			History:   []string{},
		})
		if err != nil {
			// Log error but don't fail entry creation
			_ = err
		} else if output != nil {
			// Store the roast text in the entry
			roastText := sql.NullString{String: output.RoastText, Valid: true}
			err = s.repo.UpdateRoast(ctx, e.ID, roastText)
			if err != nil {
				// Log error but don't fail entry creation
				_ = err
			}

			// Store the guilt score if scores repository available
			if s.scoresRepo != nil {
				score := int32(output.GuiltScore * 100) // Convert to 0-100 scale
				_, _ = s.scoresRepo.CreateScore(ctx, sid, &e.ID, score, nil)
			}
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

// GetEntry retrieves a single entry by ID
func (s *EntryService) GetEntry(ctx context.Context, entryID string) (sqlc.GuiltEntry, error) {
	eid, err := uuid.Parse(entryID)
	if err != nil {
		return sqlc.GuiltEntry{}, errors.New("invalid entry_id")
	}

	return s.repo.GetEntry(ctx, eid)
}

// GetEntryScore retrieves the score for a specific entry
func (s *EntryService) GetEntryScore(ctx context.Context, entryID string) (int32, error) {
	eid, err := uuid.Parse(entryID)
	if err != nil {
		return 0, errors.New("invalid entry_id")
	}

	if s.scoresRepo == nil {
		return 0, nil
	}

	score, err := s.scoresRepo.GetScoreByEntry(ctx, eid)
	if err != nil {
		return 0, nil // Return 0 if no score found
	}

	return score.AggregateScore, nil
}

func (s *EntryService) ProcessMLJob(ctx context.Context, entryID string) error {
	eid, err := uuid.Parse(entryID)
	if err != nil {
		return errors.New("invalid entry_id")
	}

	e, err := s.repo.GetEntry(ctx, eid)
	if err != nil {
		return err
	}

	var intensity = 3
	var persona = ml.PersonaRoast

	if s.prefsService != nil {
		prefs, _ := s.prefsService.GetPreferences(ctx, e.SessionID.String())
		if prefs.Metadata.Valid {
			// Parse metadata for intensity and persona if available
			var meta map[string]interface{}
			_ = json.Unmarshal(prefs.Metadata.RawMessage, &meta)
			if val, ok := meta["humor_intensity"]; ok {
				if fval, ok := val.(float64); ok {
					intensity = int(fval)
				}
			}
			if val, ok := meta["persona"]; ok {
				if sval, ok := val.(string); ok {
					switch sval {
					case "roast":
						persona = ml.PersonaRoast
					case "coach":
						persona = ml.PersonaCoach
					case "chill":
						persona = ml.PersonaChill
					default:
						persona = ml.PersonaNeutral
					}
				}
			}
		}
	}

	// For now, use orchestrator if available, otherwise skip
	if s.orchestrator != nil {
		out, err := s.orchestrator.Run(ctx, ml.HybridInput{
			Text:      e.EntryText,
			UserID:    e.SessionID.String(),
			Intensity: intensity,
			Persona:   persona,
		})
		if err != nil {
			_ = s.repo.UpdateEntryStatus(ctx, e.ID, "failed")
			return err
		}

		// Update roast text
		roastText := sql.NullString{String: out.RoastText, Valid: true}
		_ = s.repo.UpdateRoast(ctx, e.ID, roastText)

		// Create score with entry_id
		if s.scoresRepo != nil {
			score := int32(out.GuiltScore * 100) // Convert to 0-100 scale
			_, _ = s.scoresRepo.CreateScore(ctx, e.SessionID, &e.ID, score, nil)
		}

		// Mark as completed
		_ = s.repo.UpdateEntryStatus(ctx, e.ID, "completed")
	}

	return nil
}
