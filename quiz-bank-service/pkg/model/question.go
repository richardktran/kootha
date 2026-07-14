package model

import "github.com/richardktran/realtime-quiz/gen"

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correct_answer"`
	TimeLimit     int      `json:"time_limit"`
}

func QuestionToProto(q *Question) *gen.BankQuestion {
	return &gen.BankQuestion{
		Id:            q.ID,
		Question:      q.Question,
		Options:       q.Options,
		CorrectAnswer: int32(q.CorrectAnswer),
		TimeLimit:     int32(q.TimeLimit),
	}
}

func QuestionsToProto(questions []Question) []*gen.BankQuestion {
	out := make([]*gen.BankQuestion, len(questions))
	for i := range questions {
		out[i] = QuestionToProto(&questions[i])
	}
	return out
}
