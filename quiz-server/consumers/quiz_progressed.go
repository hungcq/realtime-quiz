package consumers

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"quiz/core/managers"
	"quiz/core/models"
)

type QuizProgressedEventHandler struct {
	quizSessionManager *managers.QuizSession
}

func NewQuizProgressedEventHandler(quizSessionManager *managers.QuizSession) *QuizProgressedEventHandler {
	return &QuizProgressedEventHandler{quizSessionManager: quizSessionManager}
}

func (q *QuizProgressedEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println("received quiz progressed event:", string(msg.Value))
		event := &models.QuizProgressedEvent{}
		if err := json.Unmarshal(msg.Value, event); err != nil {
			fmt.Println("error parsing quiz progressed event", err, string(msg.Value))
		}
		if err := q.handleMsg(event); err != nil {
			fmt.Println("error handling quiz progressed event", err, string(msg.Value))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (q *QuizProgressedEventHandler) handleMsg(event *models.QuizProgressedEvent) error {
	return q.quizSessionManager.OnQuizProgressed(event)
}

func (q *QuizProgressedEventHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (q *QuizProgressedEventHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
