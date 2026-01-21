package services_test

import (
	"context"
	"testing"

	cacheDomain "guiltmachine/internal/cache/domain"
	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestPreferencesServiceWithCache(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	cache := openPreferencesCache(t)
	prefs := svcs.NewPreferencesService(repo.Preferences, cache)

	u, err := users.CreateUser(ctx, "svc-cache-prefs@test.com", "password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// upsert preferences = write-through
	theme := "dark"
	_, err = prefs.UpsertPreferences(ctx, u.ID.String(), &theme, true, "")
	if err != nil {
		t.Fatalf("upsert prefs failed: %v", err)
	}

	// manually set in cache (simulating write-through)
	cachePrefs := cacheDomain.PreferencesRecord{
		HumorIntensity: 2,
		Persona:        "test",
		Timezone:       "UTC",
	}
	if err := cache.SetPreferences(ctx, u.ID.String(), cachePrefs); err != nil {
		t.Fatalf("set prefs in cache failed: %v", err)
	}

	// direct cache verify
	p, ok, err := cache.GetPreferences(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("cache get failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected cache hit for preferences")
	}
	if p.HumorIntensity != 2 {
		t.Fatalf("expected intensity=2 got %d", p.HumorIntensity)
	}

	// update to new value
	theme2 := "light"
	_, err = prefs.UpsertPreferences(ctx, u.ID.String(), &theme2, false, "")
	if err != nil {
		t.Fatalf("upsert prefs failed: %v", err)
	}

	// verify cache update
	cachePrefs2 := cacheDomain.PreferencesRecord{
		HumorIntensity: 5,
		Persona:        "test2",
		Timezone:       "PST",
	}
	if err := cache.SetPreferences(ctx, u.ID.String(), cachePrefs2); err != nil {
		t.Fatalf("set prefs in cache failed: %v", err)
	}

	p2, ok2, err := cache.GetPreferences(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("cache get failed: %v", err)
	}
	if !ok2 {
		t.Fatalf("expected cache entry after second update")
	}
	if p2.HumorIntensity != 5 {
		t.Fatalf("expected new intensity=5 got %d", p2.HumorIntensity)
	}

	// Fallback flow test: manually invalidate cache then call service.Get
	if err := cache.InvalidatePreferences(ctx, u.ID.String()); err != nil {
		t.Fatalf("invalidate cache failed: %v", err)
	}

	p3, err := prefs.GetPreferences(ctx, u.ID.String())
	if err != nil {
		t.Fatalf("service get failed: %v", err)
	}
	if !p3.Theme.Valid || p3.Theme.String != theme2 {
		t.Fatalf("expected fallback to DB theme=%s got %v", theme2, p3.Theme)
	}
}
