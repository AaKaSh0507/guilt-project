package ml

import (
	"context"

	gen "guiltmachine/internal/proto/gen/ml"
)

type Service interface {
	Roast(ctx context.Context, req *gen.RoastRequest) (*gen.RoastResponse, error)
}

type MLService struct {
	infer Service
}

func NewMLService(infer Service) *MLService {
	return &MLService{infer: infer}
}

func (m *MLService) Roast(ctx context.Context, req *gen.RoastRequest) (*gen.RoastResponse, error) {
	return m.infer.Roast(ctx, req)
}
