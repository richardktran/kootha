package workers

import "context"

type UserJoinedWorker struct {
	repo quizSessionRepository
}

func NewUserJoinedWorker(repo quizSessionRepository) *UserJoinedWorker {
	return &UserJoinedWorker{
		repo: repo,
	}
}

func (w *UserJoinedWorker) JoinQuiz(ctx context.Context, sessionId, userId string) error {
	return w.repo.JoinQuiz(ctx, sessionId, userId)
}
