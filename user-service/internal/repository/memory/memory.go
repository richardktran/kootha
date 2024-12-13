package memory

import (
	"context"
	"sync"

	"github.com/richardktran/realtime-quiz/user-service/internal/repository"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
)

type Repository struct {
	sync.RWMutex
	data map[string]*model.User
}

func New() *Repository {
	return &Repository{
		data: make(map[string]*model.User),
	}
}

func (r *Repository) CreateUser(_ context.Context, user *model.User) error {
	r.Lock()
	defer r.Unlock()

	r.data[user.ID] = user

	return nil
}

func (r *Repository) GetByID(_ context.Context, id string) (*model.User, error) {
	r.RLock()
	defer r.RUnlock()

	user, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}

	return user, nil
}
