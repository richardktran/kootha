package workers

import (
	"context"
	"errors"

	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

var ErrNotFound = errors.New("not found")

type quizSessionRepository interface {
	CreateQuizSession(context.Context, *model.QuizSession) (*model.QuizSession, error)
	JoinQuiz(ctx context.Context, sessionId, userId string) error
}
