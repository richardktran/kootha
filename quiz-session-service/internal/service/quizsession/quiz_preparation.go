package quizsession

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/richardktran/realtime-quiz/pkg/topics"
	"github.com/richardktran/realtime-quiz/quiz-session-service/internal/repository"
	"github.com/richardktran/realtime-quiz/quiz-session-service/pkg/model"
)

func (s *Service) CreateQuizSession(ctx context.Context, data *model.QuizSession) (*model.QuizSession, error) {

	// Can be removed because we will store to the database by consumer
	// session, err := s.repo.CreateQuizSession(ctx, data)
	session := data

	encodedSession, err := s.encodeSession(session)
	if err != nil {
		return nil, err
	}

	err = s.producer.Produce(ctx, topics.QuizSessionCreated, encodedSession)
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

func (s *Service) JoinQuiz(ctx context.Context, quizSessionId, userId string) (*model.QuizSession, error) {
	session, err := s.GetSessionById(ctx, quizSessionId)
	if err != nil {
		return nil, err
	}

	data := map[string]interface{}{
		"sessionId": session.ID,
		"userId":    userId,
	}

	encodedSession, err := s.encodeSession(data)
	if err != nil {
		return nil, err
	}

	err = s.producer.Produce(ctx, topics.UserJoinedQuiz, encodedSession)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Service) encodeSession(data any) ([]byte, error) {
	encodedSession, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return encodedSession, nil
}
