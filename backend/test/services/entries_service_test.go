package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestEntriesService_CreateEntry(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)

	u, _ := users.CreateUser(ctx, "entry-create@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	e, err := entries.CreateEntry(ctx, sess.ID.String(), "test entry text", 5)
	if err != nil {
		t.Fatalf("create entry failed: %v", err)
	}

	if e.EntryText != "test entry text" {
		t.Fatalf("entry text mismatch: expected 'test entry text', got %s", e.EntryText)
	}
	if !e.GuiltLevel.Valid || e.GuiltLevel.Int32 != 5 {
		t.Fatalf("guilt level mismatch: expected 5, got %v", e.GuiltLevel)
	}
	if e.SessionID != sess.ID {
		t.Fatalf("session ID mismatch: expected %s, got %s", sess.ID, e.SessionID)
	}
}

func TestEntriesService_CreateEntryNoGuiltLevel(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)

	u, _ := users.CreateUser(ctx, "entry-no-level@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	e, err := entries.CreateEntry(ctx, sess.ID.String(), "entry without level", 0)
	if err != nil {
		t.Fatalf("create entry failed: %v", err)
	}

	if e.EntryText != "entry without level" {
		t.Fatalf("entry text mismatch: expected 'entry without level', got %s", e.EntryText)
	}
}

func TestEntriesService_ListEntriesBySession(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)

	u, _ := users.CreateUser(ctx, "entry-list@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	// Create multiple entries
	for i := 0; i < 3; i++ {
		_, err := entries.CreateEntry(ctx, sess.ID.String(), "entry text", int32(i))
		if err != nil {
			t.Fatalf("create entry failed: %v", err)
		}
	}

	list, err := entries.ListEntries(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("list entries failed: %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(list))
	}

	for _, e := range list {
		if e.SessionID != sess.ID {
			t.Fatalf("session ID mismatch in listed entry: expected %s, got %s", sess.ID, e.SessionID)
		}
	}
}

func TestEntriesService_ListEntriesEmptySession(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	sessions := svcs.NewSessionService(repo.Sessions, nil)
	entries := svcs.NewEntryService(repo.Entries)

	u, _ := users.CreateUser(ctx, "entry-empty@test.com", "password")
	sess, _ := sessions.CreateSession(ctx, u.ID.String(), nil)

	list, err := entries.ListEntries(ctx, sess.ID.String())
	if err != nil {
		t.Fatalf("list entries failed: %v", err)
	}

	if len(list) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(list))
	}
}

func TestEntriesService_CreateEntryInvalidSessionID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	entries := svcs.NewEntryService(repo.Entries)

	_, err := entries.CreateEntry(ctx, "not-a-uuid", "text", 5)
	if err == nil {
		t.Fatal("expected error for invalid session UUID, but got none")
	}
}
