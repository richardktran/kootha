package grpc

import (
	"context"
	"errors"

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
		HostID:   req.GetHostId(),
		Status:   model.StatusWaiting,
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
	if errors.Is(err, quizsession.ErrNotFound) {
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

	session, err := h.service.JoinQuiz(ctx, req.GetQuizSessionId(), req.GetUserId(), req.GetName())
	if errors.Is(err, quizsession.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, "session not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to join: %v", err)
	}

	return &gen.JoinQuizResponse{
		QuizSession: model.QuizSessionToProto(session),
	}, nil
}

func (h *Handler) StartSession(ctx context.Context, req *gen.StartSessionRequest) (*gen.StartSessionResponse, error) {
	if req == nil || req.GetSessionId() == "" || req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionId and userId required")
	}

	session, question, err := h.service.StartSession(ctx, req.GetSessionId(), req.GetUserId(), req.GetQuestionCount())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &gen.StartSessionResponse{
		QuizSession: model.QuizSessionToProto(session),
		Question:    model.PublicQuestionToProto(question),
	}, nil
}

func (h *Handler) SubmitAnswer(ctx context.Context, req *gen.SubmitAnswerRequest) (*gen.SubmitAnswerResponse, error) {
	if req == nil || req.GetSessionId() == "" || req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionId and userId required")
	}

	err := h.service.SubmitAnswer(ctx, req.GetSessionId(), req.GetUserId(), req.GetQuestionId(), int(req.GetSelectedOption()), int(req.GetTimeToAnswer()))
	if errors.Is(err, quizsession.ErrAlreadyAnswered) {
		return &gen.SubmitAnswerResponse{Accepted: false}, nil
	}
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &gen.SubmitAnswerResponse{Accepted: true}, nil
}

func (h *Handler) NextQuestion(ctx context.Context, req *gen.NextQuestionRequest) (*gen.NextQuestionResponse, error) {
	if req == nil || req.GetSessionId() == "" || req.GetUserId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionId and userId required")
	}

	finished, question, index, err := h.service.NextQuestion(ctx, req.GetSessionId(), req.GetUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	resp := &gen.NextQuestionResponse{
		Finished:      finished,
		QuestionIndex: int32(index),
	}
	if !finished {
		resp.Question = model.PublicQuestionToProto(question)
	}
	return resp, nil
}

func (h *Handler) EndSession(ctx context.Context, req *gen.EndSessionRequest) (*gen.EndSessionResponse, error) {
	if req == nil || req.GetSessionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionId required")
	}

	if err := h.service.EndSession(ctx, req.GetSessionId(), req.GetUserId()); err != nil {
		return nil, mapDomainError(err)
	}

	session, err := h.service.GetSessionById(ctx, req.GetSessionId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &gen.EndSessionResponse{QuizSession: model.QuizSessionToProto(session)}, nil
}

func (h *Handler) ReassignHost(ctx context.Context, req *gen.ReassignHostRequest) (*gen.ReassignHostResponse, error) {
	if req == nil || req.GetSessionId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "sessionId required")
	}

	hostID, err := h.service.ReassignHost(ctx, req.GetSessionId(), req.GetLeavingUserId())
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &gen.ReassignHostResponse{HostId: hostID}, nil
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, quizsession.ErrNotFound):
		return status.Errorf(codes.NotFound, "%v", err)
	case errors.Is(err, quizsession.ErrUnauthorized):
		return status.Errorf(codes.PermissionDenied, "%v", err)
	case errors.Is(err, quizsession.ErrInvalidState):
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	default:
		return status.Errorf(codes.Internal, "%v", err)
	}
}
