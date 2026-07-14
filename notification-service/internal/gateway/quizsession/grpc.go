package quizsession

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/utils/grpcutil"
)

const serviceName = "quiz-session"

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (g *Gateway) JoinQuiz(ctx context.Context, sessionID, userID, name string) (*gen.JoinQuizResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)
	return client.JoinQuiz(ctx, &gen.JoinQuizRequest{
		QuizSessionId: sessionID,
		UserId:        userID,
		Name:          name,
	})
}

func (g *Gateway) StartSession(ctx context.Context, sessionID, userID string, questionCount int32) (*gen.StartSessionResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)
	return client.StartSession(ctx, &gen.StartSessionRequest{
		SessionId:     sessionID,
		UserId:        userID,
		QuestionCount: questionCount,
	})
}

func (g *Gateway) SubmitAnswer(ctx context.Context, sessionID, userID, questionID string, selectedOption, timeToAnswer int32) (*gen.SubmitAnswerResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)
	return client.SubmitAnswer(ctx, &gen.SubmitAnswerRequest{
		SessionId:      sessionID,
		UserId:         userID,
		QuestionId:     questionID,
		SelectedOption: selectedOption,
		TimeToAnswer:   timeToAnswer,
	})
}

func (g *Gateway) NextQuestion(ctx context.Context, sessionID, userID string) (*gen.NextQuestionResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)
	return client.NextQuestion(ctx, &gen.NextQuestionRequest{
		SessionId: sessionID,
		UserId:    userID,
	})
}

func (g *Gateway) ReassignHost(ctx context.Context, sessionID, leavingUserID string) (*gen.ReassignHostResponse, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)
	return client.ReassignHost(ctx, &gen.ReassignHostRequest{
		SessionId:     sessionID,
		LeavingUserId: leavingUserID,
	})
}
