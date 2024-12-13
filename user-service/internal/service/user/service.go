package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/richardktran/realtime-quiz/user-service/internal/repository"
	"github.com/richardktran/realtime-quiz/user-service/pkg/model"
	"golang.org/x/exp/rand"
)

// ErrorNotFound is returned when a requested record is not found.
var ErrNotFound = errors.New("not found")

type userRepository interface {
	CreateUser(context.Context, *model.User) error
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

func (s *Service) CreateUser(ctx context.Context, name string) error {
	// hard code for now
	id := fmt.Sprintf("%d", rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int())

	user := &model.User{
		ID:   id,
		Name: name,
	}

	return s.repo.CreateUser(ctx, user)
}

func (s *Service) GetUser(ctx context.Context, id string) (*model.User, error) {
	res, err := s.repo.GetByID(ctx, id)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	return res, err
}
