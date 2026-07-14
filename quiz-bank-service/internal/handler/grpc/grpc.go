package grpc

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/internal/service/quizbank"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/pkg/model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type quizBankService interface {
	GetRandomQuestions(context.Context, int) ([]model.Question, error)
	GetQuestionsByIds(context.Context, []string) ([]model.Question, error)
}

type Handler struct {
	gen.UnimplementedQuizBankServiceServer
	service quizBankService
}

func New(svc *quizbank.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) GetRandomQuestions(ctx context.Context, req *gen.GetRandomQuestionsRequest) (*gen.GetRandomQuestionsResponse, error) {
	if req == nil || req.GetCount() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or invalid count")
	}

	questions, err := h.service.GetRandomQuestions(ctx, int(req.GetCount()))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get random questions: %v", err)
	}

	return &gen.GetRandomQuestionsResponse{
		Questions: model.QuestionsToProto(questions),
	}, nil
}

func (h *Handler) GetQuestionsByIds(ctx context.Context, req *gen.GetQuestionsByIdsRequest) (*gen.GetQuestionsByIdsResponse, error) {
	if req == nil || len(req.GetIds()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "nil request or empty ids")
	}

	questions, err := h.service.GetQuestionsByIds(ctx, req.GetIds())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get questions by ids: %v", err)
	}

	return &gen.GetQuestionsByIdsResponse{
		Questions: model.QuestionsToProto(questions),
	}, nil
}
