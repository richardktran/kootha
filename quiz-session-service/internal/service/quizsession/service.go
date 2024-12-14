package quizsession

import (
	"context"
	"errors"

	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var ErrNotFound = errors.New("not found")

type quizSessionRepository interface {
	CreateQuizSession(context.Context, *model.QuizSession) (*model.QuizSession, error)
	GetSessionById(context.Context, string) (*model.QuizSession, error)
}

type Service struct {
	repo quizSessionRepository
}

func New(repo quizSessionRepository) *Service {
	return &Service{
		repo: repo,
	}
}
