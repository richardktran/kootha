package quizbank

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"

	goredis "github.com/redis/go-redis/v9"
	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/quiz-bank-service/pkg/model"
)

var ErrNotFound = errors.New("not found")

type questionRepository interface {
	GetAll(context.Context) ([]model.Question, error)
	GetByIDs(context.Context, []string) ([]model.Question, error)
	GetRandom(context.Context, int) ([]model.Question, error)
}

type Service struct {
	repo  questionRepository
	cache *redis.Client
}

func New(repo questionRepository, cache *redis.Client) *Service {
	return &Service{
		repo:  repo,
		cache: cache,
	}
}

func (s *Service) GetRandomQuestions(ctx context.Context, count int) ([]model.Question, error) {
	if count <= 0 {
		return []model.Question{}, nil
	}

	all, err := s.getAllQuestions(ctx)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		return []model.Question{}, nil
	}

	if count >= len(all) {
		return shuffleCopy(all), nil
	}

	shuffled := shuffleCopy(all)
	return shuffled[:count], nil
}

func (s *Service) GetQuestionsByIds(ctx context.Context, ids []string) ([]model.Question, error) {
	if len(ids) == 0 {
		return []model.Question{}, nil
	}

	all, err := s.getAllQuestions(ctx)
	if err != nil {
		return nil, err
	}

	if len(all) == 0 {
		questions, err := s.repo.GetByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		return orderByIDs(questions, ids), nil
	}

	byID := make(map[string]model.Question, len(all))
	for _, q := range all {
		byID[q.ID] = q
	}

	result := make([]model.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := byID[id]; ok {
			result = append(result, q)
		}
	}

	return result, nil
}

func (s *Service) getAllQuestions(ctx context.Context) ([]model.Question, error) {
	cached, err := s.cache.Get(ctx, redis.BankQuestionsKey())
	if err == nil {
		var questions []model.Question
		if err := json.Unmarshal([]byte(cached), &questions); err == nil {
			return questions, nil
		}
	} else if !errors.Is(err, goredis.Nil) {
		return nil, err
	}

	questions, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(questions)
	if err != nil {
		return questions, nil
	}

	_ = s.cache.Set(ctx, redis.BankQuestionsKey(), string(data), 0)

	return questions, nil
}

func shuffleCopy(questions []model.Question) []model.Question {
	shuffled := make([]model.Question, len(questions))
	copy(shuffled, questions)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func orderByIDs(questions []model.Question, ids []string) []model.Question {
	byID := make(map[string]model.Question, len(questions))
	for _, q := range questions {
		byID[q.ID] = q
	}

	result := make([]model.Question, 0, len(ids))
	for _, id := range ids {
		if q, ok := byID[id]; ok {
			result = append(result, q)
		}
	}

	return result
}
