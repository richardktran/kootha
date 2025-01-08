package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	idModel "github.com/richardktran/realtime-quiz/id-generation-service/pkg/model"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type idGenerationGateway interface {
	GenerateId(context.Context, string) (*idModel.IDGenerator, error)
}

type Handler struct {
	gen.UnimplementedQuizSessionServiceServer
	idGenerationGateway idGenerationGateway
	service             *quizsession.Service
}

func New(svc *quizsession.Service, idGenerationGateway idGenerationGateway) *Handler {
	return &Handler{
		service:             svc,
		idGenerationGateway: idGenerationGateway,
	}
}

func (h *Handler) CreateQuizSession(ctx context.Context, req *gen.CreateQuizSessionRequest) (*gen.CreateQuizSessionResponse, error) {
	if req == nil || req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty name")
	}

	generatedId, err := h.idGenerationGateway.GenerateId(ctx, "quiz-session")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate id: %v", err)
	}

	newSession := &model.QuizSession{
		ID:       generatedId.ID,
		Name:     req.GetName(),
		Duration: int(req.GetDuration()),
	}

	session, err := h.service.CreateQuizSession(ctx, newSession)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session: %v", err)
	}

	return &gen.CreateQuizSessionResponse{
		QuizSession: model.QuizSessionToProto(session),
	}, nil
}

func (h *Handler) GetQuizSessionById(ctx context.Context, req *gen.GetQuizSessionByIdRequest) (*gen.GetQuizSessionByIdResponse, error) {
	if req == nil || req.GetId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty id")
	}

	session, err := h.service.GetSessionById(ctx, req.GetId())

	if err != nil && err == quizsession.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get session: %v", err)
	}

	return &gen.GetQuizSessionByIdResponse{
		QuizSession: model.QuizSessionToProto(session),
	}, nil
}

func (h *Handler) JoinQuiz(ctx context.Context, req *gen.JoinQuizRequest) (*gen.JoinQuizResponse, error) {
	if req == nil || req.GetQuizSessionId() == "" || req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty session id or user id")
	}

	session, err := h.service.JoinQuiz(ctx, req.GetQuizSessionId(), req.GetUserId())

	if err != nil && err == quizsession.ErrNotFound {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}

	return &gen.JoinQuizResponse{
		QuizSession: model.QuizSessionToProto(session),
	}, nil
}
