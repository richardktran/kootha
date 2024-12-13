package user

import (
	"context"
	"errors"

	"github.com/richardktran/realtime-quiz/user-service/internal/repository"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
)

// ErrorNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type userRepository interface {
	CreateUser(context.Context, *model.User) (*model.User, error)
	GetByID(context.Context, string) (*model.User, error)
}

type Service struct {
	repo userRepository
}

func New(repo userRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateUser(ctx context.Context, id, name string) (*model.User, error) {
	user := &model.User{
		ID:   id,
		Name: name,
	}

	user, err := s.repo.CreateUser(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) GetUser(ctx context.Context, id string) (*model.User, error) {
	res, err := s.repo.GetByID(ctx, id)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	return res, err
}
