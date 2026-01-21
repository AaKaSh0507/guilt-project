package ml

import (
	"context"
	"strings"
	"testing"

	ml "guiltmachine/internal/ml"
)

// MockLLM is a mock implementation of the LLM interface for testing
type MockLLM struct {
	responseText string
	err          error
}

func (m *MockLLM) Generate(ctx context.Context, in ml.HybridInput) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.responseText, nil
}

func TestHybridOrchestratorBasic(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "you really procrastinated hard"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated the entire day",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaRoast,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	if output == nil {
		t.Fatalf("expected output, got nil")
	}

	if output.GuiltScore <= 0 || output.GuiltScore > 1 {
		t.Fatalf("expected guilt score between 0-1, got %f", output.GuiltScore)
	}

	if output.RoastText == "" {
		t.Fatalf("expected roast text, got empty")
	}

	if len(output.Tags) == 0 {
		t.Fatalf("expected tags, got none")
	}
}

func TestHybridOrchestratorPersonaRoastHighIntensity(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "lazy developer"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 8,
		Persona:   ml.PersonaRoast,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// High intensity roast should have fire emoji
	if !strings.HasPrefix(output.RoastText, "🔥") {
		t.Fatalf("expected fire emoji for high intensity roast, got: %s", output.RoastText)
	}
}

func TestHybridOrchestratorPersonaRoastLowIntensity(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "you could do better"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 3,
		Persona:   ml.PersonaRoast,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Low intensity roast should have smile emoji
	if !strings.HasPrefix(output.RoastText, "🙂") {
		t.Fatalf("expected smile emoji for low intensity roast, got: %s", output.RoastText)
	}
}

func TestHybridOrchestratorPersonaCoach(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "let's work on your focus"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaCoach,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	if !strings.HasPrefix(output.RoastText, "Coach:") {
		t.Fatalf("expected Coach prefix, got: %s", output.RoastText)
	}
}

func TestHybridOrchestratorPersonaChill(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "no worries, next time"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaChill,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	if !strings.HasPrefix(output.RoastText, "Chill:") {
		t.Fatalf("expected Chill prefix, got: %s", output.RoastText)
	}
}

func TestHybridOrchestratorSafetyFilter(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "you should kill your procrastination"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaNeutral,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Check that "kill" was replaced
	if len(output.SafetyFlags) == 0 || output.SafetyFlags[0] != "violent_content" {
		t.Fatalf("expected violent_content flag, got: %v", output.SafetyFlags)
	}

	if len(output.RoastText) > 0 && output.RoastText[0:4] == "kill" {
		t.Fatalf("expected 'kill' to be replaced, got: %s", output.RoastText)
	}
}

func TestHybridOrchestratorGuiltScore(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "procrastination"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	// Test short text
	inputShort := ml.HybridInput{
		Text:      "lazy",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaNeutral,
		History:   []string{},
	}

	outputShort, err := orchestrator.Run(ctx, inputShort)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Test longer text
	inputLong := ml.HybridInput{
		Text:      "I procrastinated the entire day and did nothing productive at all today",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaNeutral,
		History:   []string{},
	}

	outputLong, err := orchestrator.Run(ctx, inputLong)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Longer text should have higher score
	if outputLong.GuiltScore <= outputShort.GuiltScore {
		t.Fatalf("longer text should have higher score: short=%f, long=%f", outputShort.GuiltScore, outputLong.GuiltScore)
	}

	// Score should be capped at 1
	if outputLong.GuiltScore > 1 {
		t.Fatalf("score should not exceed 1, got %f", outputLong.GuiltScore)
	}
}

func TestHybridOrchestratorGuiltScoreCap(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "procrastination"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	// Very long text that would exceed score of 1
	inputVeryLong := ml.HybridInput{
		Text:      "I procrastinated the entire day and did nothing productive at all today and I feel really bad about it and I don't know what to do anymore because I keep making the same mistakes over and over again",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaNeutral,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, inputVeryLong)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Score should be capped at 1
	if output.GuiltScore > 1.0 {
		t.Fatalf("score should be capped at 1.0, got %f", output.GuiltScore)
	}
}

func TestHybridOrchestratorHybridTag(t *testing.T) {
	ctx := context.Background()
	mockLLM := &MockLLM{responseText: "you procrastinated"}

	orchestrator := ml.NewHybridOrchestrator(mockLLM)

	input := ml.HybridInput{
		Text:      "I procrastinated",
		UserID:    "user1",
		Intensity: 5,
		Persona:   ml.PersonaNeutral,
		History:   []string{},
	}

	output, err := orchestrator.Run(ctx, input)
	if err != nil {
		t.Fatalf("hybrid run failed: %v", err)
	}

	// Should always have "hybrid" tag
	found := false
	for _, tag := range output.Tags {
		if tag == "hybrid" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'hybrid' tag, got: %v", output.Tags)
	}
}
