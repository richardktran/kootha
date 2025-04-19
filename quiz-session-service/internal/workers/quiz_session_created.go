package workers

import (
	"context"

	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

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
