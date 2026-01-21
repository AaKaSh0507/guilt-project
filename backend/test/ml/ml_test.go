package ml

import (
	"context"
	"testing"

	ml "guiltmachine/internal/ml"
	gen "guiltmachine/internal/proto/gen/ml"
)

func TestMLStubRoast(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	resp, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "I procrastinated again",
		UserId:         "u1",
		HumorIntensity: 3,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if resp.GuiltScore <= 0 {
		t.Fatalf("expected non-zero score, got %f", resp.GuiltScore)
	}

	if resp.RoastText == "" {
		t.Fatalf("expected roast text, got empty string")
	}
}

func TestMLStubRoastMildIntensity(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	resp, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "I procrastinated",
		UserId:         "u1",
		HumorIntensity: 2,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if resp.RoastText != "Mild roast: you really wrote that?" {
		t.Fatalf("expected mild roast, got: %s", resp.RoastText)
	}
}

func TestMLStubRoastAggressiveIntensity(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	resp, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "I procrastinated again and again",
		UserId:         "u1",
		HumorIntensity: 8,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if resp.RoastText != "Aggressive roast: bro what is this?" {
		t.Fatalf("expected aggressive roast, got: %s", resp.RoastText)
	}
}

func TestMLStubGuiltScoreCalculation(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	// Test with short text
	respShort, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "lazy",
		UserId:         "u1",
		HumorIntensity: 5,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	// Test with longer text
	respLong, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "I procrastinated the entire day and did nothing productive at all",
		UserId:         "u1",
		HumorIntensity: 5,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if respLong.GuiltScore <= respShort.GuiltScore {
		t.Fatalf("longer text should have higher guilt score: short=%f, long=%f", respShort.GuiltScore, respLong.GuiltScore)
	}
}

func TestMLStubScoresCapped(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	// Test that scores never exceed 1.0
	resp, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "word word word word word word word word word word word word word word word word word word word word",
		UserId:         "u1",
		HumorIntensity: 5,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if resp.GuiltScore > 1.0 {
		t.Fatalf("guilt score should not exceed 1.0, got %f", resp.GuiltScore)
	}
}

func TestMLStubResponseStructure(t *testing.T) {
	ctx := context.Background()
	s := ml.NewInferenceStub()

	resp, err := s.Roast(ctx, &gen.RoastRequest{
		EntryText:      "test entry",
		UserId:         "u1",
		HumorIntensity: 5,
		History:        []string{},
	})
	if err != nil {
		t.Fatalf("roast failed: %v", err)
	}

	if resp.GuiltScore < 0 || resp.GuiltScore > 1 {
		t.Fatalf("invalid guilt score: %f", resp.GuiltScore)
	}

	if len(resp.Tags) == 0 {
		t.Fatalf("expected tags, got empty")
	}

	if len(resp.SafetyFlags) != 0 {
		t.Fatalf("expected no safety flags for stub, got %v", resp.SafetyFlags)
	}
}
