package grpc

import (
	"context"

	v1 "guiltmachine/internal/proto/gen"
	"guiltmachine/internal/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type EntryHandler struct {
	v1.UnimplementedEntryServiceServer
	svc *services.EntryService
}

func NewEntryHandler(svc *services.EntryService) *EntryHandler {
	return &EntryHandler{svc: svc}
}

func (h *EntryHandler) CreateEntry(ctx context.Context, req *v1.CreateEntryRequest) (*v1.CreateEntryResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}
	if req.Text == "" {
		return nil, status.Error(codes.InvalidArgument, "text required")
	}
	if req.Level < 0 {
		return nil, status.Error(codes.InvalidArgument, "level must be >= 0")
	}

	e, err := h.svc.CreateEntry(ctx, req.SessionId, req.Text, req.Level)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &v1.CreateEntryResponse{
		EntryId:   e.ID.String(),
		SessionId: e.SessionID.String(),
		Text:      nullableText(e.EntryText),
		Level:     int32(e.GuiltLevel.Int32),
		CreatedAt: timestamppb.New(e.CreatedAt),
	}, nil
}

func (h *EntryHandler) ListEntries(ctx context.Context, req *v1.ListEntriesRequest) (*v1.ListEntriesResponse, error) {
	if req.SessionId == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}

	entries, err := h.svc.ListEntries(ctx, req.SessionId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	items := make([]*v1.EntryItem, 0, len(entries))
	for _, e := range entries {
		items = append(items, &v1.EntryItem{
			EntryId:   e.ID.String(),
			Text:      nullableText(e.EntryText),
			Level:     int32(e.GuiltLevel.Int32),
			CreatedAt: timestamppb.New(e.CreatedAt),
		})
	}

	return &v1.ListEntriesResponse{Entries: items}, nil
}

func nullableText(v string) string {
	return v
}
