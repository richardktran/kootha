package memory

import (
	"context"
	"sync"

	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

type Repository struct {
	sync.RWMutex
	data map[string]*model.QuizSession
}

func New() *Repository {
	return &Repository{
		data: make(map[string]*model.QuizSession),
	}
}

func (r *Repository) CreateQuizSession(_ context.Context, session *model.QuizSession) (*model.QuizSession, error) {
	r.Lock()
	defer r.Unlock()

	r.data[session.ID] = session

	return session, nil
}

func (r *Repository) GetSessionById(_ context.Context, id string) (*model.QuizSession, error) {
	r.RLock()
	defer r.RUnlock()

	session, ok := r.data[id]

	if !ok {
		return nil, repository.ErrNotFound
	}

	return session, nil
}
