package repository

import (
	"context"
	"database/sql"

	sqlc "guiltmachine/internal/db/sqlc"

	"github.com/google/uuid"
)

type UsersRepository interface {
	CreateUser(ctx context.Context, email string, passwordHash string) (sqlc.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (sqlc.User, error)
}

type SessionsRepository interface {
	CreateSession(ctx context.Context, userID uuid.UUID, notes *string) (sqlc.GuiltSession, error)
	EndSession(ctx context.Context, sessionID uuid.UUID) (sqlc.GuiltSession, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (sqlc.GuiltSession, error)
	ListSessionsByUser(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]sqlc.GuiltSession, error)
}

type EntriesRepository interface {
	CreateEntry(ctx context.Context, sessionID uuid.UUID, text string, level int32) (sqlc.GuiltEntry, error)
	ListEntriesBySession(ctx context.Context, sessionID uuid.UUID) ([]sqlc.GuiltEntry, error)
	UpdateRoast(ctx context.Context, entryID uuid.UUID, roastText sql.NullString) error
	UpdateEntryStatus(ctx context.Context, entryID uuid.UUID, status string) error
	GetEntry(ctx context.Context, entryID uuid.UUID) (sqlc.GuiltEntry, error)
}

type ScoresRepository interface {
	CreateScore(ctx context.Context, sessionID uuid.UUID, score int32, meta any) (sqlc.GuiltScore, error)
	GetScoreBySession(ctx context.Context, sessionID uuid.UUID) (sqlc.GuiltScore, error)
}

type PreferencesRepository interface {
	UpsertPreferences(ctx context.Context, userID uuid.UUID, theme *string, notifications bool, metadata any) (sqlc.UserPreference, error)
	GetPreferencesByUserID(ctx context.Context, userID uuid.UUID) (sqlc.UserPreference, error)
}
