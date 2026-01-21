package ml

import (
	"context"
	"strings"

	gen "guiltmachine/internal/proto/gen/ml"
)

type InferenceStub struct{}

func NewInferenceStub() *InferenceStub {
	return &InferenceStub{}
}

func (s *InferenceStub) Roast(ctx context.Context, req *gen.RoastRequest) (*gen.RoastResponse, error) {
	text := req.EntryText

	score := float64(len(strings.Fields(text))) / 10.0
	if score > 1 {
		score = 1
	}

	roast := "Mild roast: you really wrote that?"
	if req.HumorIntensity > 5 {
		roast = "Aggressive roast: bro what is this?"
	}

	return &gen.RoastResponse{
		GuiltScore:  score,
		RoastText:   roast,
		Tags:        []string{"stub"},
		SafetyFlags: []string{},
	}, nil
}
