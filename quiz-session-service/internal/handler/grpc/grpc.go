package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/service/quizsession"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
	"golang.org/x/exp/rand"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	gen.UnimplementedQuizSessionServiceServer
	service *quizsession.Service
}

func New(svc *quizsession.Service) *Handler {
	return &Handler{
		service: svc,
	}
}

func (h *Handler) CreateQuizSession(ctx context.Context, req *gen.CreateQuizSessionRequest) (*gen.CreateQuizSessionResponse, error) {
	if req == nil || req.GetName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty name")
	}

	id := fmt.Sprintf("%d", rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())

	newSession := &model.QuizSession{
		ID:       id,
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
