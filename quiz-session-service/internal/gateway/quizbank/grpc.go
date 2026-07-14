package quizbank

import (
	"context"

	"github.com/richardktran/realtime-quiz/gen"
	"github.com/richardktran/realtime-quiz/pkg/discovery"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
	"github.com/richardktran/realtime-quiz/utils/grpcutil"
)

const serviceName = "quiz-bank"

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry: registry}
}

func (g *Gateway) GetRandomQuestions(ctx context.Context, count int32) ([]model.Question, error) {
	conn, err := grpcutil.ServiceConnection(ctx, serviceName, g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewQuizBankServiceClient(conn)
	resp, err := client.GetRandomQuestions(ctx, &gen.GetRandomQuestionsRequest{Count: count})
	if err != nil {
		return nil, err
	}

	return bankQuestionsToModel(resp.GetQuestions()), nil
}

func bankQuestionsToModel(qs []*gen.BankQuestion) []model.Question {
	out := make([]model.Question, 0, len(qs))
	for _, q := range qs {
		out = append(out, model.Question{
			ID:            q.GetId(),
			Question:      q.GetQuestion(),
			Options:       q.GetOptions(),
			CorrectAnswer: int(q.GetCorrectAnswer()),
			TimeLimit:     int(q.GetTimeLimit()),
		})
	}
	return out
}
