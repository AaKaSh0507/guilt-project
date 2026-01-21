package grpc

import (
	"context"
	"database/sql"
	"encoding/json"

	v1 "guiltmachine/internal/proto/gen"
	"guiltmachine/internal/services"
	"github.com/sqlc-dev/pqtype"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PreferencesHandler struct {
	v1.UnimplementedPreferencesServiceServer
	svc *services.PreferencesService
}

func NewPreferencesHandler(svc *services.PreferencesService) *PreferencesHandler {
	return &PreferencesHandler{svc: svc}
}

func (h *PreferencesHandler) UpsertPreferences(ctx context.Context, req *v1.UpsertPreferencesRequest) (*v1.UpsertPreferencesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	var themePtr *string
	if req.Theme != "" {
		t := req.Theme
		themePtr = &t
	}

	pref, err := h.svc.UpsertPreferences(ctx, req.UserId, themePtr, req.Notifications, req.MetadataJson)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &v1.UpsertPreferencesResponse{
		UserId:        pref.UserID.String(),
		Theme:         nullableStringPref(pref.Theme),
		Notifications: pref.NotificationsEnabled,
		MetadataJson:  rawMetaPref(pref.Metadata),
	}, nil
}

func (h *PreferencesHandler) GetPreferences(ctx context.Context, req *v1.GetPreferencesRequest) (*v1.GetPreferencesResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	pref, err := h.svc.GetPreferences(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &v1.GetPreferencesResponse{
		UserId:        pref.UserID.String(),
		Theme:         nullableStringPref(pref.Theme),
		Notifications: pref.NotificationsEnabled,
		MetadataJson:  rawMetaPref(pref.Metadata),
	}, nil
}

// helpers

func nullableStringPref(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func rawMetaPref(m pqtype.NullRawMessage) string {
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
