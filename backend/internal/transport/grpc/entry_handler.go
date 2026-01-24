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

	entryStatus := "pending"
	if e.Status.Valid {
		entryStatus = e.Status.String
	}

	return &v1.CreateEntryResponse{
		EntryId:   e.ID.String(),
		SessionId: e.SessionID.String(),
		Text:      nullableText(e.EntryText),
		Level:     int32(e.GuiltLevel.Int32),
		CreatedAt: timestamppb.New(e.CreatedAt),
		Status:    entryStatus,
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
		entryStatus := ""
		if e.Status.Valid {
			entryStatus = e.Status.String
		}
		roastText := ""
		if e.RoastText.Valid {
			roastText = e.RoastText.String
		}
		// Get score for entry
		score, _ := h.svc.GetEntryScore(ctx, e.ID.String())

		items = append(items, &v1.EntryItem{
			EntryId:    e.ID.String(),
			Text:       nullableText(e.EntryText),
			Level:      int32(e.GuiltLevel.Int32),
			CreatedAt:  timestamppb.New(e.CreatedAt),
			Status:     entryStatus,
			RoastText:  roastText,
			GuiltScore: score,
		})
	}

	return &v1.ListEntriesResponse{Entries: items}, nil
}

func (h *EntryHandler) GetEntry(ctx context.Context, req *v1.GetEntryRequest) (*v1.GetEntryResponse, error) {
	if req.EntryId == "" {
		return nil, status.Error(codes.InvalidArgument, "entry_id required")
	}

	e, err := h.svc.GetEntry(ctx, req.EntryId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	entryStatus := "pending"
	if e.Status.Valid {
		entryStatus = e.Status.String
	}
	roastText := ""
	if e.RoastText.Valid {
		roastText = e.RoastText.String
	}

	// Get score for entry
	score, _ := h.svc.GetEntryScore(ctx, req.EntryId)

	return &v1.GetEntryResponse{
		EntryId:    e.ID.String(),
		SessionId:  e.SessionID.String(),
		Text:       nullableText(e.EntryText),
		Level:      int32(e.GuiltLevel.Int32),
		CreatedAt:  timestamppb.New(e.CreatedAt),
		Status:     entryStatus,
		RoastText:  roastText,
		GuiltScore: score,
	}, nil
}

func nullableText(v string) string {
	return v
}
