package quizsession

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/utils/grpcutil"
)

var serviceName = "quiz-session"

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{
		registry: registry,
	}
}

func (g *Gateway) CreateQuizSession(ctx context.Context, name, hostID string) (*gen.QuizSession, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)

	response, err := client.CreateQuizSession(ctx, &gen.CreateQuizSessionRequest{
		Name:   name,
		HostId: hostID,
	})
	if err != nil {
		return nil, err
	}

	return response.GetQuizSession(), nil
}

func (g *Gateway) GetQuizSessionById(ctx context.Context, id string) (*gen.QuizSession, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)

	response, err := client.GetQuizSessionById(ctx, &gen.GetQuizSessionByIdRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}

	return response.GetQuizSession(), nil
}

func (g *Gateway) JoinQuiz(ctx context.Context, quizSessionID, userID, name string) (*gen.QuizSession, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizSessionServiceClient(conn)

	response, err := client.JoinQuiz(ctx, &gen.JoinQuizRequest{
		QuizSessionId: quizSessionID,
		UserId:        userID,
		Name:          name,
	})
	if err != nil {
		return nil, err
	}

	return response.GetQuizSession(), nil
}
