package model

import "github.com/richardktran/realtime-quiz/gen"

type QuizSession struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	// Participants []model.User `json:"participants"`
}

func QuizSessionToProto(qs *QuizSession) *gen.QuizSession {
	return &gen.QuizSession{
		Id:       qs.ID,
		Name:     qs.Name,
		Duration: int32(qs.Duration),
	}
}

func QuizSessionFromProto(qs *gen.QuizSession) QuizSession {
	return QuizSession{
		ID:       qs.Id,
		Name:     qs.Name,
		Duration: int(qs.Duration),
	}
}
