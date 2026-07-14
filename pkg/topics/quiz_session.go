package topics

const (
	QuizSessionCreated = "quiz-session-created"
	UserJoinedQuiz     = "user-joined-quiz"
	SessionStart       = "session-start"
	ChangeQuestion     = "change-question"
	SessionEnd         = "session-end"
	AnswerSubmitted    = "answer-submitted"
	RankingUpdated     = "ranking-updated"
	QuestionResult     = "question-result"
	QuizCompleted      = "quiz-completed"
)

// Redis pub/sub channel for notification fan-out across instances.
const NotificationFanout = "notification:fanout"
