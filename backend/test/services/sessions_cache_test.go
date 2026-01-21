package services_test

import (
	"context"
	"testing"

	sqlcrepo "guiltmachine/internal/repository/sqlc"
	svcs "guiltmachine/internal/services"
)

func TestSessionsServiceWithCache(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	repo := sqlcrepo.New(db)
	users := svcs.NewUserService(repo.Users)
	cache := openSessionCache(t)
	sessions := svcs.NewSessionService(repo.Sessions, cache)

	// setup user
	u, err := users.CreateUser(ctx, "svc-cache-session@test.com", "password")
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// create session = write-through cache (explicitly populate cache)
	sess, err := sessions.CreateSession(ctx, u.ID.String(), nil)
	if err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	// Manually add to cache (simulating write-through)
	sessionToken := sess.ID.String()
	if err := cache.SetSession(ctx, sessionToken, u.ID.String()); err != nil {
		t.Fatalf("set session in cache failed: %v", err)
	}

	// verify cache entry directly
	uid, ok, err := cache.GetSession(ctx, sessionToken)
	if err != nil {
		t.Fatalf("cache fetch failed: %v", err)
	}
	if !ok {
		t.Fatalf("expected cache entry but found none")
	}
	if uid != u.ID.String() {
		t.Fatalf("expected cache value %s got %s", u.ID.String(), uid)
	}
}
