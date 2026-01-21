package repo_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
)

func TestPreferencesRepo(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()
	repo := sqlcrepo.New(db)

	t.Run("upsert and fetch preferences", func(t *testing.T) {
		u, err := repo.Users.CreateUser(ctx, "prefs@test.com", "hashedpassword")
		if err != nil {
			t.Fatalf("create user failed: %v", err)
		}

		// Upsert prefs
		theme := "dark"
		pref, err := repo.Preferences.UpsertPreferences(ctx, u.ID, &theme, true, nil)
		if err != nil {
			t.Fatalf("upsert preferences failed: %v", err)
		}

		// Fetch
		pref2, err := repo.Preferences.GetPreferencesByUserID(ctx, u.ID)
		if err != nil {
			t.Fatalf("get preferences failed: %v", err)
		}
		if pref2.Theme != pref.Theme {
			t.Fatalf("mismatch in theme")
		}
	})

	t.Run("upsert consistency - update overwrites", func(t *testing.T) {
		u, _ := repo.Users.CreateUser(ctx, "prefsupsert@test.com", "hashedpassword")

		// First upsert
		theme1 := "light"
		pref1, err := repo.Preferences.UpsertPreferences(ctx, u.ID, &theme1, false, nil)
		if err != nil {
			t.Fatalf("first upsert failed: %v", err)
		}
		if pref1.Theme.String != "light" {
			t.Fatalf("expected theme light, got %s", pref1.Theme.String)
		}

		// Second upsert with different theme
		theme2 := "dark"
		pref2, err := repo.Preferences.UpsertPreferences(ctx, u.ID, &theme2, true, nil)
		if err != nil {
			t.Fatalf("second upsert failed: %v", err)
		}
		if pref2.Theme.String != "dark" {
			t.Fatalf("expected updated theme dark, got %s", pref2.Theme.String)
		}

		// Verify persistence
		pref3, _ := repo.Preferences.GetPreferencesByUserID(ctx, u.ID)
		if pref3.Theme.String != "dark" {
			t.Fatalf("expected theme to persist as dark, got %s", pref3.Theme.String)
		}
	})

	t.Run("fk user constraint", func(t *testing.T) {
		// Try to upsert preferences for non-existent user
		fakeUserID := [16]byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99}
		uuid := [16]byte(fakeUserID)
		_, err := repo.Preferences.UpsertPreferences(ctx, uuid, nil, false, nil)
		if err == nil {
			t.Fatalf("expected FK violation for non-existent user")
		}
	})
}
