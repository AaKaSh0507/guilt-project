package domain

import (
	"context"
	"encoding/json"
	"time"

	basecache "guiltmachine/internal/cache"
	"guiltmachine/internal/cache/redis"
)

// SessionRecord is what we store in Redis
type SessionRecord struct {
	UserID string `json:"user_id"`
}

// TTL recommendation: session tokens often need explicit expiry
var sessionTTL = 24 * time.Hour

type SessionCache struct {
	cache basecache.Cache
}

func NewSessionCache(c basecache.Cache) *SessionCache {
	return &SessionCache{cache: c}
}

func (s *SessionCache) SetSession(ctx context.Context, token string, userID string) error {
	record := SessionRecord{UserID: userID}
	b, err := json.Marshal(record)
	if err != nil {
		return err
	}
	key := redis.KeySession(token)
	return s.cache.Set(ctx, key, b, sessionTTL)
}

func (s *SessionCache) GetSession(ctx context.Context, token string) (userID string, ok bool, err error) {
	key := redis.KeySession(token)
	b, err := s.cache.Get(ctx, key)
	if err != nil {
		if err == basecache.ErrNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	var rec SessionRecord
	if err := json.Unmarshal(b, &rec); err != nil {
		return "", false, err
	}
	return rec.UserID, true, nil
}

func (s *SessionCache) DeleteSession(ctx context.Context, token string) error {
	key := redis.KeySession(token)
	return s.cache.Del(ctx, key)
}
