package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

// TestServicesIntegration tests complete user workflow across all services
func TestServicesIntegration_CompleteWorkflow(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)
	scores := svcs.NewScoreService(repo.Scores)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	// 1. Create user
	u, err := users.CreateUser(ctx, "integration@test.com", "password-hash")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// 2. Set user preferences
	theme := "dark"
	p, err := prefs.UpsertPreferences(ctx, u.ID.String(), &theme, true, "")
	if err != nil {
		t.Fatalf("upsert preferences failed: %v", err)
	}
	if !p.Theme.Valid || p.Theme.String != theme {
		t.Fatal("preferences not set correctly")
	}

	// 3. Create session
	notes := "guilt tracking session"
	sess, err := sessions.CreateSession(ctx, u.ID.String(), &notes)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	// 4. Create multiple entries in session
	entryIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		e, err := entries.CreateEntry(ctx, sess.ID.String(), "test entry", int32(i+1))
		if err != nil {
			t.Fatalf("create entry %d failed: %v", i, err)
		}
		entryIDs[i] = e.ID.String()
	}

	// 5. List entries in session
	entryList, err := entries.ListEntries(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("list entries failed: %v", err)
	}
	if len(entryList) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entryList))
	}

	// 6. Create score for session
	s, err := scores.CreateScore(ctx, sess.ID.String(), 72, "")
	if err != nil {
		t.Fatalf("create score failed: %v", err)
	}
	if s.AggregateScore != 72 {
		t.Fatalf("score mismatch: expected 72, got %d", s.AggregateScore)
	}

	// 7. Retrieve score
	s2, err := scores.GetScore(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("get score failed: %v", err)
	}
	if s2.ID != s.ID {
		t.Fatal("score ID mismatch")
	}

	// 8. End session
	sessEnded, err := sessions.EndSession(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("end session failed: %v", err)
	}
	if !sessEnded.EndTime.Valid {
		t.Fatal("session end time not set")
	}

	// 9. List all user sessions
	allSessions, err := sessions.ListSessionsByUser(ctx, u.ID.String(), 10, 0)
	if err != nil {
		t.Fatalf("list sessions failed: %v", err)
	}
	if len(allSessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(allSessions))
	}

	// 10. Verify user data persists
	u2, err := users.GetUser(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("get user failed: %v", err)
	}
	if u2.Email != u.Email {
		t.Fatal("user email mismatch")
	}
}

// TestServicesIntegration_MultipleUsers tests isolation between users
func TestServicesIntegration_MultipleUsers(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)

	// Create two users
	u1, _ := users.CreateUser(ctx, "user1@test.com", "pass1")
	u2, _ := users.CreateUser(ctx, "user2@test.com", "pass2")

	// Each user creates a session
	sess1, _ := sessions.CreateSession(ctx, u1.ID.String(), nil)
	sess2, _ := sessions.CreateSession(ctx, u2.ID.String(), nil)

	// Each session gets entries
	_, _ = entries.CreateEntry(ctx, sess1.ID.String(), "u1 entry", 1)
	_, _ = entries.CreateEntry(ctx, sess2.ID.String(), "u2 entry", 2)

	// Verify data isolation
	sess1List, _ := sessions.ListSessionsByUser(ctx, u1.ID.String(), 10, 0)
	sess2List, _ := sessions.ListSessionsByUser(ctx, u2.ID.String(), 10, 0)

	if len(sess1List) != 1 || len(sess2List) != 1 {
		t.Fatal("session isolation failed")
	}

	// Entries should only appear in correct session
	entries1, _ := entries.ListEntries(ctx, sess1.ID.String())
	entries2, _ := entries.ListEntries(ctx, sess2.ID.String())

	if len(entries1) != 1 || len(entries2) != 1 {
		t.Fatal("entry isolation failed")
	}
	if entries1[0].EntryText != "u1 entry" {
		t.Fatal("user1 entry text mismatch")
	}
	if entries2[0].EntryText != "u2 entry" {
		t.Fatal("user2 entry text mismatch")
	}
}
