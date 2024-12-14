package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/id-generation-service/internal/service/idgeneration"
	"github.com/richardktran/realtime-quiz/id-generation-service/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedIdGenerationServiceServer
	service *idgeneration.Service
}

func New(svc *idgeneration.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) GenerateId(ctx context.Context, req *gen.IdGenerationRequest) (*gen.IdGenerationResponse, error) {
	if req == nil || req.Entity == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty entity")
	}

	id := h.service.GenerateId(ctx, req.Entity)

	generatedID := &model.IDGenerator{
		ID:     id,
		Entity: req.Entity,
	}

	return &gen.IdGenerationResponse{
		IdGenerator: model.IDGeneratorToProto(generatedID),
	}, nil
}
