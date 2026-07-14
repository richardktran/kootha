package main

import (
	"log"

	"github.com/richardktran/realtime-quiz/pkg/message-broker/kafka"
	"github.com/richardktran/realtime-quiz/pkg/topics"
)

func main() {
	allTopics := []string{
		topics.QuizSessionCreated,
		topics.UserJoinedQuiz,
		topics.SessionStart,
		topics.ChangeQuestion,
		topics.SessionEnd,
		topics.AnswerSubmitted,
		topics.RankingUpdated,
		topics.QuestionResult,
		topics.QuizCompleted,
	}

	if err := kafka.EnsureTopics(allTopics, 1, 1); err != nil {
		log.Fatalf("ensure topics: %v", err)
	}
	log.Println("Kafka topics ready")
}
