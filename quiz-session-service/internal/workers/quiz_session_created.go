package workers

import (
	"context"
	"errors"

	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var ErrNotFound = errors.New("not found")

type quizSessionRepository interface {
	CreateQuizSession(context.Context, *model.QuizSession) (*model.QuizSession, error)
}

type QuizSessionCreatedWorker struct {
	repo quizSessionRepository
}

func NewQuizCreatedWorker(repo quizSessionRepository) *QuizSessionCreatedWorker {
	return &QuizSessionCreatedWorker{
		repo: repo,
	}
}

func (w *QuizSessionCreatedWorker) StoreQuizSession(ctx context.Context, session *model.QuizSession) (*model.QuizSession, error) {
	session, err := w.repo.CreateQuizSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}
