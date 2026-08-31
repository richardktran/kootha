package main

import (
	"log"
	"time"

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

	var last error
	for i := 1; i <= 30; i++ {
		last = kafka.EnsureTopics(allTopics, 1, 1)
		if last == nil {
			log.Println("Kafka topics ready")
			return
		}
		log.Printf("ensure topics (attempt %d/30): %v", i, last)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("ensure topics: %v", last)
}
