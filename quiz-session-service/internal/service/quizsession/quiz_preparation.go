package quizsession

import (
	"context"
	"errors"

	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

func (s *Service) CreateQuizSession(ctx context.Context, data *model.QuizSession) (*model.QuizSession, error) {
	session, err := s.repo.CreateQuizSession(ctx, data)

	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) GetSessionById(ctx context.Context, id string) (*model.QuizSession, error) {
	session, err := s.repo.GetSessionById(ctx, id)

	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}

	return session, err
}
