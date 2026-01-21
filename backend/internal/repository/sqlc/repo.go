package sqlc

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"

	dbpkg "guiltmachine/internal/db"
	sqlc "guiltmachine/internal/db/sqlc"
	"guiltmachine/internal/repository"
)

type Repos struct {
	Users       repository.UsersRepository
	Sessions    repository.SessionsRepository
	Entries     repository.EntriesRepository
	Scores      repository.ScoresRepository
	Preferences repository.PreferencesRepository
}

func New(db dbpkg.DB) *Repos {
	q := sqlc.New(db)
	return &Repos{
		Users:       &usersRepo{q},
		Sessions:    &sessionsRepo{q},
		Entries:     &entriesRepo{q},
		Scores:      &scoresRepo{q},
		Preferences: &preferencesRepo{q},
	}
}

// USERS

type usersRepo struct{ q *sqlc.Queries }

func (r *usersRepo) CreateUser(ctx context.Context, email string, passwordHash string) (sqlc.User, error) {
	params := sqlc.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	}
	return r.q.CreateUser(ctx, params)
}

func (r *usersRepo) GetUserByID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.q.GetUserByID(ctx, id)
}

func (r *usersRepo) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.q.GetUserByEmail(ctx, email)
}

// SESSIONS

type sessionsRepo struct{ q *sqlc.Queries }

func (r *sessionsRepo) CreateSession(ctx context.Context, userID uuid.UUID, notes *string) (sqlc.GuiltSession, error) {
	var ns sql.NullString
	if notes != nil {
		ns = sql.NullString{String: *notes, Valid: true}
	}
	params := sqlc.CreateSessionParams{
		UserID: userID,
		Notes:  ns,
	}
	return r.q.CreateSession(ctx, params)
}

func (r *sessionsRepo) EndSession(ctx context.Context, sessionID uuid.UUID) (sqlc.GuiltSession, error) {
	return r.q.EndSession(ctx, sessionID)
}

func (r *sessionsRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (sqlc.GuiltSession, error) {
	return r.q.GetSessionByID(ctx, id)
}

func (r *sessionsRepo) ListSessionsByUser(ctx context.Context, userID uuid.UUID, limit int32, offset int32) ([]sqlc.GuiltSession, error) {
	params := sqlc.ListSessionsByUserParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}
	return r.q.ListSessionsByUser(ctx, params)
}

// ENTRIES

type entriesRepo struct{ q *sqlc.Queries }

func (r *entriesRepo) CreateEntry(ctx context.Context, sessionID uuid.UUID, text string, level int32) (sqlc.GuiltEntry, error) {
	params := sqlc.CreateEntryParams{
		SessionID:  sessionID,
		EntryText:  text,
		GuiltLevel: sql.NullInt32{Int32: level, Valid: true},
	}
	row, err := r.q.CreateEntry(ctx, params)
	if err != nil {
		return sqlc.GuiltEntry{}, err
	}
	return sqlc.GuiltEntry{
		ID:         row.ID,
		SessionID:  row.SessionID,
		EntryText:  row.EntryText,
		GuiltLevel: row.GuiltLevel,
		RoastText:  row.RoastText,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}, nil
}

func (r *entriesRepo) ListEntriesBySession(ctx context.Context, sessionID uuid.UUID) ([]sqlc.GuiltEntry, error) {
	rows, err := r.q.ListEntriesBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	entries := make([]sqlc.GuiltEntry, len(rows))
	for i, row := range rows {
		entries[i] = sqlc.GuiltEntry{
			ID:         row.ID,
			SessionID:  row.SessionID,
			EntryText:  row.EntryText,
			GuiltLevel: row.GuiltLevel,
			RoastText:  row.RoastText,
			CreatedAt:  row.CreatedAt,
			UpdatedAt:  row.UpdatedAt,
		}
	}
	return entries, nil
}

func (r *entriesRepo) UpdateRoast(ctx context.Context, entryID uuid.UUID, roastText sql.NullString) error {
	params := sqlc.UpdateRoastParams{
		ID:        entryID,
		RoastText: roastText,
	}
	return r.q.UpdateRoast(ctx, params)
}

// SCORES

type scoresRepo struct{ q *sqlc.Queries }

func (r *scoresRepo) CreateScore(ctx context.Context, sessionID uuid.UUID, score int32, meta any) (sqlc.GuiltScore, error) {
	var rm pqtype.NullRawMessage
	if meta != nil {
		b, _ := json.Marshal(meta)
		rm = pqtype.NullRawMessage{RawMessage: b, Valid: true}
	}
	params := sqlc.CreateScoreParams{
		SessionID:      sessionID,
		AggregateScore: score,
		Meta:           rm,
	}
	return r.q.CreateScore(ctx, params)
}

func (r *scoresRepo) GetScoreBySession(ctx context.Context, sessionID uuid.UUID) (sqlc.GuiltScore, error) {
	return r.q.GetScoreBySession(ctx, sessionID)
}

// PREFERENCES

type preferencesRepo struct{ q *sqlc.Queries }

func (r *preferencesRepo) UpsertPreferences(ctx context.Context, userID uuid.UUID, theme *string, notifications bool, metadata any) (sqlc.UserPreference, error) {
	var ts sql.NullString
	if theme != nil {
		ts = sql.NullString{String: *theme, Valid: true}
	}
	var rm pqtype.NullRawMessage
	if metadata != nil {
		b, _ := json.Marshal(metadata)
		rm = pqtype.NullRawMessage{RawMessage: b, Valid: true}
	}
	params := sqlc.UpsertUserPreferencesParams{
		UserID:               userID,
		Theme:                ts,
		NotificationsEnabled: notifications,
		Metadata:             rm,
	}
	return r.q.UpsertUserPreferences(ctx, params)
}

func (r *preferencesRepo) GetPreferencesByUserID(ctx context.Context, userID uuid.UUID) (sqlc.UserPreference, error) {
	return r.q.GetPreferencesByUserID(ctx, userID)
}
