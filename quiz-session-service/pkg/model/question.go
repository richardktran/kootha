package model

type Question struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	Options       []string `json:"options"`
	CorrectAnswer int      `json:"correctAnswer"`
	TimeLimit     int      `json:"timeLimit"`
}
