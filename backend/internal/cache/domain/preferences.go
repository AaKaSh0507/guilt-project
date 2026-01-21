package domain

import (
	"context"
	"encoding/json"
	"time"

	basecache "guiltmachine/internal/cache"
	"guiltmachine/internal/cache/redis"
)

// Example preferences structure; adapt to your actual model
type PreferencesRecord struct {
	HumorIntensity int    `json:"humor_intensity"`
	Persona        string `json:"persona"`
	Timezone       string `json:"timezone"`
}

var preferencesTTL = 12 * time.Hour // Optional cache TTL

type PreferencesCache struct {
	cache basecache.Cache
}

func NewPreferencesCache(c basecache.Cache) *PreferencesCache {
	return &PreferencesCache{cache: c}
}

func (p *PreferencesCache) SetPreferences(ctx context.Context, userID string, prefs PreferencesRecord) error {
	b, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	key := redis.KeyPreferences(userID)
	return p.cache.Set(ctx, key, b, preferencesTTL)
}

func (p *PreferencesCache) GetPreferences(ctx context.Context, userID string) (PreferencesRecord, bool, error) {
	key := redis.KeyPreferences(userID)
	b, err := p.cache.Get(ctx, key)
	if err != nil {
		if err == basecache.ErrNotFound {
			return PreferencesRecord{}, false, nil
		}
		return PreferencesRecord{}, false, err
	}
	var prefs PreferencesRecord
	if err := json.Unmarshal(b, &prefs); err != nil {
		return PreferencesRecord{}, false, err
	}
	return prefs, true, nil
}

func (p *PreferencesCache) InvalidatePreferences(ctx context.Context, userID string) error {
	key := redis.KeyPreferences(userID)
	return p.cache.Del(ctx, key)
}
