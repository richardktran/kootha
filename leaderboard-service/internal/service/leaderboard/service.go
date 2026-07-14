package leaderboard

import (
	"context"
	"fmt"
	"time"

	"github.com/richardktran/realtime-quiz/pkg/cache/redis"
	"github.com/richardktran/realtime-quiz/pkg/events"
)

type Service struct {
	redis *redis.Client
}

func New(r *redis.Client) *Service {
	return &Service{redis: r}
}

func (s *Service) HandleAnswer(ctx context.Context, evt events.AnswerSubmitted) error {
	scoredKey := scoreDedupKey(evt.SessionID, evt.QuestionID, evt.UserID)
	set, err := s.redis.SetNX(ctx, scoredKey, "1", 24*time.Hour)
	if err != nil {
		return fmt.Errorf("score dedupe: %w", err)
	}
	if !set {
		return nil
	}

	leaderboardKey := redis.LeaderboardKey(evt.SessionID)
	namesKey := leaderboardKey + ":names"

	if err := s.redis.HSet(ctx, namesKey, evt.UserID, evt.Name); err != nil {
		return fmt.Errorf("store participant name: %w", err)
	}

	increment := float64(0)
	if evt.SelectedOption == evt.CorrectOption {
		increment = 1
	}

	if _, err := s.redis.ZIncrBy(ctx, leaderboardKey, increment, evt.UserID); err != nil {
		return fmt.Errorf("update score: %w", err)
	}

	return nil
}

func (s *Service) GetRanking(ctx context.Context, sessionID string) ([]events.Participant, error) {
	leaderboardKey := redis.LeaderboardKey(sessionID)
	namesKey := leaderboardKey + ":names"

	scores, err := s.redis.ZRevRangeWithScores(ctx, leaderboardKey, 0, -1)
	if err != nil {
		return nil, fmt.Errorf("fetch scores: %w", err)
	}

	names, err := s.redis.HGetAll(ctx, namesKey)
	if err != nil {
		return nil, fmt.Errorf("fetch names: %w", err)
	}

	participants := make([]events.Participant, 0, len(scores))
	for _, entry := range scores {
		userID, ok := entry.Member.(string)
		if !ok {
			userID = fmt.Sprint(entry.Member)
		}

		participants = append(participants, events.Participant{
			ID:    userID,
			Name:  names[userID],
			Score: int(entry.Score),
		})
	}

	return participants, nil
}

func scoreDedupKey(sessionID, questionID, userID string) string {
	return redis.AnswerDedupKey(sessionID, questionID, userID) + ":leaderboard-scored"
}
