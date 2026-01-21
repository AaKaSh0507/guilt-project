package grpc

import (
	"context"
	v1 "guiltmachine/internal/proto/gen"
	"guiltmachine/internal/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	v1.UnimplementedUserServiceServer
	svc *services.UserService
}

func NewUserHandler(svc *services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	if req.Email == "" || req.PasswordHash == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password_hash required")
	}

	u, err := h.svc.CreateUser(ctx, req.Email, req.PasswordHash)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &v1.CreateUserResponse{
		Id:        u.ID.String(),
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}

	u, err := h.svc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &v1.GetUserResponse{
		Id:        u.ID.String(),
		Email:     u.Email,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}, nil
}
