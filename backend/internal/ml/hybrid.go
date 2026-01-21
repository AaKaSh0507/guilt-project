package ml

import (
	"context"
	"strings"
)

type HybridOrchestrator struct {
	llm LLM
}

func NewHybridOrchestrator(llm LLM) *HybridOrchestrator {
	return &HybridOrchestrator{llm: llm}
}

func (h *HybridOrchestrator) Run(ctx context.Context, in HybridInput) (*HybridOutput, error) {
	raw, err := h.llm.Generate(ctx, in)
	if err != nil {
		return nil, err
	}

	roast := adjustPersona(raw, in.Persona, in.Intensity)
	safeRoast, safetyFlags := safetyFilter(roast)

	score := guiltScore(in.Text)

	return &HybridOutput{
		GuiltScore:  score,
		RoastText:   safeRoast,
		Tags:        []string{"hybrid"},
		SafetyFlags: safetyFlags,
	}, nil
}

func guiltScore(text string) float64 {
	w := float64(len(strings.Fields(text)))
	s := w / 12.0
	if s > 1 {
		s = 1
	}
	return s
}

func adjustPersona(raw string, persona Persona, intensity int) string {
	switch persona {
	case PersonaRoast:
		if intensity > 6 {
			return "🔥 " + raw
		}
		return "🙂 " + raw
	case PersonaCoach:
		return "Coach: " + raw
	case PersonaChill:
		return "Chill: " + raw
	default:
		return raw
	}
}

func safetyFilter(raw string) (string, []string) {
	flags := []string{}
	lower := strings.ToLower(raw)

	// trivial safety filter for demo
	if strings.Contains(lower, "kill") {
		flags = append(flags, "violent_content")
		raw = strings.ReplaceAll(raw, "kill", "avoid")
	}
	return raw, flags
}
