package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	idGenerationModel "github.com/richardktran/realtime-quiz/id-generation-service/pkg/model"
	"github.com/richardktran/realtime-quiz/user-service/internal/service/user"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type idGenerationGateway interface {
	GenerateId(context.Context, string) (*idGenerationModel.IDGenerator, error)
}

type Handler struct {
	gen.UnimplementedUserServiceServer
	idGenerationGateway idGenerationGateway
	service             *user.Service
}

func New(svc *user.Service, idGenerationGateway idGenerationGateway) *Handler {
	return &Handler{
		service:             svc,
		idGenerationGateway: idGenerationGateway,
	}
}

func (h *Handler) CreateUser(ctx context.Context, req *gen.CreateUserRequest) (*gen.CreateUserResponse, error) {
	if req == nil || req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty name")
	}

	generatedId, err := h.idGenerationGateway.GenerateId(ctx, "user")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate id: %v", err)
	}

	user, err := h.service.CreateUser(ctx, generatedId.ID, req.GetName())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	return &gen.CreateUserResponse{
		User: model.UserToProto(user),
	}, nil
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
