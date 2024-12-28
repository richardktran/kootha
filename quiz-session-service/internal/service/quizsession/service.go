package quizsession

import (
	"context"
	"errors"

	messagebroker "github.com/richardktran/realtime-quiz/pkg/message-broker"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var ErrNotFound = errors.New("not found")

type quizSessionRepository interface {
	CreateQuizSession(context.Context, *model.QuizSession) (*model.QuizSession, error)
	GetSessionById(context.Context, string) (*model.QuizSession, error)
}

type Service struct {
	repo     quizSessionRepository
	producer messagebroker.Producer
}

func New(repo quizSessionRepository, producer messagebroker.Producer) *Service {
	return &Service{
		repo:     repo,
		producer: producer,
	}
}
