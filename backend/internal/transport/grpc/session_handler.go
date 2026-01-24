package grpc

import (
	"context"
	"database/sql"

	v1 "guiltmachine/internal/proto/gen/v1"
	"guiltmachine/internal/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SessionHandler struct {
	v1.UnimplementedSessionServiceServer
	svc *services.SessionService
}

func NewSessionHandler(svc *services.SessionService) *SessionHandler {
	return &SessionHandler{svc: svc}
}

func (h *SessionHandler) CreateSession(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	notes := req.Notes
	if notes == "" {
		notes = ""
	}

	result, err := h.svc.CreateSessionWithJWT(ctx, req.UserId, &notes)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &v1.CreateSessionResponse{
		Id:        result.Session.ID.String(),
		UserId:    result.Session.UserID.String(),
		Notes:     nullableStringSession(result.Session.Notes),
		CreatedAt: timestamppb.New(result.Session.StartTime),
		Jwt:       result.JWT,
	}, nil
}

func (h *SessionHandler) EndSession(ctx context.Context, req *v1.EndSessionRequest) (*v1.EndSessionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}

	sess, err := h.svc.EndSession(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.EndSessionResponse{
		Id:        sess.ID.String(),
		UserId:    sess.UserID.String(),
		Notes:     nullableStringSession(sess.Notes),
		CreatedAt: timestamppb.New(sess.StartTime),
		EndedAt:   nullableTimestampSession(sess.EndTime),
	}, nil
}

func (h *SessionHandler) GetSession(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id required")
	}

	sess, err := h.svc.GetSession(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.GetSessionResponse{
		Id:        sess.ID.String(),
		UserId:    sess.UserID.String(),
		Notes:     nullableStringSession(sess.Notes),
		CreatedAt: timestamppb.New(sess.StartTime),
		EndedAt:   nullableTimestampSession(sess.EndTime),
	}, nil
}

func (h *SessionHandler) ListSessionsByUser(ctx context.Context, req *v1.ListSessionsByUserRequest) (*v1.ListSessionsByUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	sessions, err := h.svc.ListSessionsByUser(ctx, req.UserId, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	items := make([]*v1.SessionItem, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, &v1.SessionItem{
			Id:        s.ID.String(),
			UserId:    s.UserID.String(),
			Notes:     nullableStringSession(s.Notes),
			CreatedAt: timestamppb.New(s.StartTime),
			EndedAt:   nullableTimestampSession(s.EndTime),
		})
	}

	return &v1.ListSessionsByUserResponse{
		Sessions: items,
	}, nil
}

// helpers

func nullableStringSession(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func nullableTimestampSession(nt sql.NullTime) *timestamppb.Timestamp {
	if nt.Valid {
		return timestamppb.New(nt.Time)
	}
	return nil
}
