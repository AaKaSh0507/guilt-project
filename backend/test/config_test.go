package test

import (
	"os"
	"testing"
)

func TestConfigPresent(t *testing.T) {
	if os.Getenv("TEST_DB_URL") == "" {
		t.Fatal("TEST_DB_URL not set")
	}
	if os.Getenv("TEST_REDIS_URL") == "" {
		t.Fatal("TEST_REDIS_URL not set")
	}
}
