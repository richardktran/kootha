package model

import "github.com/richardktran/realtime-quiz/gen"

const (
	StatusWaiting    = "waiting"
	StatusInProgress = "in-progress"
	StatusFinished   = "finished"
)

type QuizSession struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Duration     int           `json:"duration"`
	HostID       string        `json:"hostId"`
	Status       string        `json:"status"`
	CurrentIndex int           `json:"currentIndex"`
	QuestionIDs  []string      `json:"questionIds"`
	Participants []Participant `json:"participants"`
}

type Participant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correctAnswer"`
	TimeLimit     int      `json:"timeLimit"`
}

func (q Question) ToPublic() PublicQuestion {
	return PublicQuestion{
		ID:        q.ID,
		Question:  q.Question,
		Options:   q.Options,
		TimeLimit: q.TimeLimit,
	}
}

type PublicQuestion struct {
	ID        string   `json:"id"`
	Question  string   `json:"question"`
	Options   []string `json:"options"`
	TimeLimit int      `json:"timeLimit"`
}

// SessionState is the live Redis-cached room state.
type SessionState struct {
	ID                   string                  `json:"id"`
	Name                 string                  `json:"name"`
	HostID               string                  `json:"hostId"`
	Status               string                  `json:"status"`
	CurrentQuestionIndex int                     `json:"currentQuestionIndex"`
	Questions            []Question              `json:"questions"`
	Participants         map[string]*Participant `json:"participants"`
}

func QuizSessionToProto(qs *QuizSession) *gen.QuizSession {
	if qs == nil {
		return nil
	}
	participants := make([]*gen.Participant, 0, len(qs.Participants))
	for _, p := range qs.Participants {
		participants = append(participants, &gen.Participant{
			Id:    p.ID,
			Name:  p.Name,
			Score: int32(p.Score),
		})
	}
	return &gen.QuizSession{
		Id:            qs.ID,
		Name:          qs.Name,
		Duration:      int32(qs.Duration),
		HostId:        qs.HostID,
		Status:        qs.Status,
		CurrentIndex:  int32(qs.CurrentIndex),
		QuestionIds:   qs.QuestionIDs,
		Participants:  participants,
	}
}

func QuizSessionFromProto(qs *gen.QuizSession) QuizSession {
	participants := make([]Participant, 0, len(qs.GetParticipants()))
	for _, p := range qs.GetParticipants() {
		participants = append(participants, Participant{
			ID:    p.GetId(),
			Name:  p.GetName(),
			Score: int(p.GetScore()),
		})
	}
	return QuizSession{
		ID:           qs.GetId(),
		Name:         qs.GetName(),
		Duration:     int(qs.GetDuration()),
		HostID:       qs.GetHostId(),
		Status:       qs.GetStatus(),
		CurrentIndex: int(qs.GetCurrentIndex()),
		QuestionIDs:  qs.GetQuestionIds(),
		Participants: participants,
	}
}

func PublicQuestionToProto(q PublicQuestion) *gen.PublicQuestion {
	return &gen.PublicQuestion{
		Id:        q.ID,
		Question:  q.Question,
		Options:   q.Options,
		TimeLimit: int32(q.TimeLimit),
	}
}
