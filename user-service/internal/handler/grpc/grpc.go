package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/user-service/internal/service/user"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedUserServiceServer
	service *user.Service
}

func New(svc *user.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) CreateUser(ctx context.Context, req *gen.CreateUserRequest) (*gen.CreateUserResponse, error) {
	if req == nil || req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty name")
	}

	if err := h.service.CreateUser(ctx, req.GetName()); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &gen.CreateUserResponse{}, nil
}

func (h *Handler) GetUserById(ctx context.Context, req *gen.GetUserByIdRequest) (*gen.GetUserByIdResponse, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty id")
	}

	user, err := h.service.GetUser(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	return &gen.GetUserByIdResponse{
		User: model.UserToProto(user),
	}, nil
}
