package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestSessionsService_CreateSession(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil) // nil cache for now

	u, err := users.CreateUser(ctx, "session-create@test.com", "password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	sess, err := sessions.CreateSession(ctx, u.ID.String(), nil)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	if sess.UserID != u.ID {
		t.Fatalf("user ID mismatch: expected %s, got %s", u.ID, sess.UserID)
	}
	if sess.EndTime.Valid {
		t.Fatal("new session should not have end time")
	}
}

func TestSessionsService_CreateSessionWithNotes(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)

	u, _ := users.CreateUser(ctx, "session-notes@test.com", "password")

	notes := "test notes"
	sess, err := sessions.CreateSession(ctx, u.ID.String(), &notes)
	if err != nil {
		t.Fatalf("create session with notes failed: %v", err)
	}

	if !sess.Notes.Valid || sess.Notes.String != notes {
		t.Fatalf("notes mismatch: expected %s, got %v", notes, sess.Notes)
	}
}

func TestSessionsService_GetSession(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)

	u, _ := users.CreateUser(ctx, "session-get@test.com", "password")
	sess, err := sessions.CreateSession(ctx, u.ID.String(), nil)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	sess2, err := sessions.GetSession(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("get session failed: %v", err)
	}

	if sess2.ID != sess.ID {
		t.Fatalf("session ID mismatch: expected %s, got %s", sess.ID, sess2.ID)
	}
}

func TestSessionsService_EndSession(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)

	u, _ := users.CreateUser(ctx, "session-end@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	ended, err := sessions.EndSession(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("end session failed: %v", err)
	}

	if !ended.EndTime.Valid {
		t.Fatal("ended session should have end time")
	}
}

func TestSessionsService_ListSessionsByUser(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)

	u, _ := users.CreateUser(ctx, "session-list@test.com", "password")

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		_, err := sessions.CreateSession(ctx, u.ID.String(), nil)
		if err != nil {
			t.Fatalf("create session failed: %v", err)
		}
	}

	list, err := sessions.ListSessionsByUser(ctx, u.ID.String(), 10, 0)
	if err != nil {
		t.Fatalf("list sessions failed: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(list))
	}

	for _, sess := range list {
		if sess.UserID != u.ID {
			t.Fatalf("user ID mismatch in listed session: expected %s, got %s", u.ID, sess.UserID)
		}
	}
}

func TestSessionsService_GetSessionInvalidUUID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	sessions := svcs.NewSessionService(repo.Sessions, nil)

	_, err := sessions.GetSession(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, but got none")
	}
}
