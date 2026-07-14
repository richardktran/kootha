package events

// SessionCreated is published on quiz-session-created.
type SessionCreated struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Duration int    `json:"duration"`
	HostID   string `json:"hostId"`
	Status   string `json:"status"`
}

// UserJoined is published on user-joined-quiz.
type UserJoined struct {
	SessionID string `json:"sessionId"`
	UserID    string `json:"userId"`
	Name      string `json:"name"`
}

// SessionStart is published on session-start.
type SessionStart struct {
	SessionID     string          `json:"sessionId"`
	QuestionIndex int             `json:"questionIndex"`
	Question      PublicQuestion  `json:"question"`
	Participants  []Participant   `json:"participants"`
	HostID        string          `json:"hostId"`
}

// ChangeQuestion is published on change-question.
type ChangeQuestion struct {
	SessionID     string         `json:"sessionId"`
	QuestionIndex int            `json:"questionIndex"`
	Question      PublicQuestion `json:"question"`
}

// SessionEnd is published on session-end.
type SessionEnd struct {
	SessionID    string        `json:"sessionId"`
	Participants []Participant `json:"participants"`
}

// AnswerSubmitted is published on answer-submitted.
type AnswerSubmitted struct {
	SessionID      string `json:"sessionId"`
	UserID         string `json:"userId"`
	Name           string `json:"name"`
	QuestionID     string `json:"questionId"`
	SelectedOption int    `json:"selectedOption"`
	CorrectOption  int    `json:"correctOption"`
	TimeToAnswer   int    `json:"timeToAnswer"`
	QuestionIndex  int    `json:"questionIndex"`
}

// RankingUpdated is published on ranking-updated.
type RankingUpdated struct {
	SessionID    string        `json:"sessionId"`
	Participants []Participant `json:"participants"`
}

// QuestionResult is published when a question is revealed (all answers in or timer expired).
type QuestionResult struct {
	SessionID      string `json:"sessionId"`
	QuestionID     string `json:"questionId"`
	QuestionIndex  int    `json:"questionIndex"`
	CorrectAnswer  int    `json:"correctAnswer"`
	Reason         string `json:"reason"` // "all_submitted" | "timeout"
}

type Participant struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type PublicQuestion struct {
	ID        string   `json:"id"`
	Question  string   `json:"question"`
	Options   []string `json:"options"`
	TimeLimit int      `json:"timeLimit"`
}

// FanoutMessage is published on Redis pub/sub for multi-instance WS broadcast.
type FanoutMessage struct {
	SessionID string      `json:"sessionId"`
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
}
