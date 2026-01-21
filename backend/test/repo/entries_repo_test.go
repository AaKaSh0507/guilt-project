package repo_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
)

func TestEntriesRepo(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := sqlcrepo.New(db)

	t.Run("create and list entries", func(t *testing.T) {
		u, err := repo.Users.CreateUser(ctx, "entry@test.com", "hashedpassword")
		if err != nil {
			t.Fatalf("create user failed: %v", err)
		}

		s, err := repo.Sessions.CreateSession(ctx, u.ID, nil)
		if err != nil {
			t.Fatalf("create session failed: %v", err)
		}

		e, err := repo.Entries.CreateEntry(ctx, s.ID, "test entry text", 5)
		if err != nil {
			t.Fatalf("create entry failed: %v", err)
		}

		e2, err := repo.Entries.ListEntriesBySession(ctx, s.ID)
		if err != nil {
			t.Fatalf("list entries failed: %v", err)
		}
		if len(e2) == 0 {
			t.Fatalf("expected at least one entry")
		}
		if e2[0].EntryText != e.EntryText {
			t.Fatalf("entry text mismatch")
		}
	})

	t.Run("ordering list entries", func(t *testing.T) {
		u, _ := repo.Users.CreateUser(ctx, "entryorder@test.com", "hashedpassword")
		s, _ := repo.Sessions.CreateSession(ctx, u.ID, nil)

		// Create multiple entries
		e1, _ := repo.Entries.CreateEntry(ctx, s.ID, "first entry", 3)
		e2, _ := repo.Entries.CreateEntry(ctx, s.ID, "second entry", 4)
		e3, _ := repo.Entries.CreateEntry(ctx, s.ID, "third entry", 5)

		list, err := repo.Entries.ListEntriesBySession(ctx, s.ID)
		if err != nil || len(list) != 3 {
			t.Fatalf("list entries failed or wrong count: %v", err)
		}

		// Verify entries exist in result
		if list[0].EntryText != e1.EntryText {
			t.Fatalf("first entry mismatch")
		}
		if list[1].EntryText != e2.EntryText {
			t.Fatalf("second entry mismatch")
		}
		if list[2].EntryText != e3.EntryText {
			t.Fatalf("third entry mismatch")
		}
	})

	t.Run("fk session constraint", func(t *testing.T) {
		// Try to create entry with non-existent session
		fakeSessionID := [16]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00}
		uuid := [16]byte(fakeSessionID)
		_, err := repo.Entries.CreateEntry(ctx, uuid, "bad session entry", 5)
		if err == nil {
			t.Fatalf("expected FK violation for non-existent session")
		}
	})
}
