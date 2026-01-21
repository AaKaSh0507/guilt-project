package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestPreferencesService_UpsertPreferences(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	u, _ := users.CreateUser(ctx, "prefs-upsert@test.com", "password")

	theme := "dark"
	p, err := prefs.UpsertPreferences(ctx, u.ID.String(), &theme, true, "")
	if err != nil {
		t.Fatalf("upsert preferences failed: %v", err)
	}

	if !p.Theme.Valid || p.Theme.String != theme {
		t.Fatalf("theme mismatch: expected %s, got %v", theme, p.Theme)
	}
	if !p.NotificationsEnabled {
		t.Fatal("notifications should be enabled")
	}
}

func TestPreferencesService_UpsertPreferencesNoTheme(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	u, _ := users.CreateUser(ctx, "prefs-no-theme@test.com", "password")

	p, err := prefs.UpsertPreferences(ctx, u.ID.String(), nil, false, "")
	if err != nil {
		t.Fatalf("upsert preferences failed: %v", err)
	}

	if p.Theme.Valid {
		t.Fatalf("theme should be null, got %s", p.Theme.String)
	}
	if p.NotificationsEnabled {
		t.Fatal("notifications should be disabled")
	}
}

func TestPreferencesService_GetPreferences(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	u, _ := users.CreateUser(ctx, "prefs-get@test.com", "password")

	theme := "light"
	_, err := prefs.UpsertPreferences(ctx, u.ID.String(), &theme, true, "")
	if err != nil {
		t.Fatalf("upsert preferences failed: %v", err)
	}

	p2, err := prefs.GetPreferences(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("get preferences failed: %v", err)
	}

	if !p2.Theme.Valid || p2.Theme.String != theme {
		t.Fatalf("theme mismatch: expected %s, got %v", theme, p2.Theme)
	}
	if p2.UserID != u.ID {
		t.Fatalf("user ID mismatch: expected %s, got %s", u.ID, p2.UserID)
	}
}

func TestPreferencesService_UpsertPreferencesInvalidUUID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	_, err := prefs.UpsertPreferences(ctx, "not-a-uuid", nil, false, "")
	if err == nil {
		t.Fatal("expected error for invalid UUID, but got none")
	}
}

func TestPreferencesService_GetPreferencesInvalidUUID(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	prefs := svcs.NewPreferencesService(repo.Preferences, nil)

	_, err := prefs.GetPreferences(ctx, "not-a-uuid")
	if err == nil {
		t.Fatal("expected error for invalid UUID, but got none")
	}
}
