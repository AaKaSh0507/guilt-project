package grpc

import (
	"context"
	"encoding/json"

	v1 "guiltmachine/internal/proto/gen"
	"guiltmachine/internal/services"
	"github.com/sqlc-dev/pqtype"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ScoreHandler struct {
	v1.UnimplementedScoreServiceServer
	svc *services.ScoreService
}

func NewScoreHandler(svc *services.ScoreService) *ScoreHandler {
	return &ScoreHandler{svc: svc}
}

func (h *ScoreHandler) CreateScore(ctx context.Context, req *v1.CreateScoreRequest) (*v1.CreateScoreResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}

	sc, err := h.svc.CreateScore(ctx, req.SessionId, req.Score, req.MetaJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &v1.CreateScoreResponse{
		ScoreId:   sc.ID.String(),
		SessionId: sc.SessionID.String(),
		Score:     sc.AggregateScore,
		MetaJson:  rawMetaScore(sc.Meta),
		CreatedAt: timestamppb.New(sc.CreatedAt),
	}, nil
}

func (h *ScoreHandler) GetScore(ctx context.Context, req *v1.GetScoreRequest) (*v1.GetScoreResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}

	sc, err := h.svc.GetScore(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.GetScoreResponse{
		ScoreId:   sc.ID.String(),
		SessionId: sc.SessionID.String(),
		Score:     sc.AggregateScore,
		MetaJson:  rawMetaScore(sc.Meta),
		CreatedAt: timestamppb.New(sc.CreatedAt),
	}, nil
}

func rawMetaScore(m pqtype.NullRawMessage) string {
	if !m.Valid {
		return ""
	}
	var v any
	if err := json.Unmarshal(m.RawMessage, &v); err != nil {
		return ""
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
