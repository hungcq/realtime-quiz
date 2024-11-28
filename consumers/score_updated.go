package consumers

import (
	"encoding/json"
	"fmt"

	"github.com/IBM/sarama"
	"quiz/core/managers"
	"quiz/core/models"
)

type ScoreUpdatedEventHandler struct {
	quizSessionManager *managers.QuizSession
}

func NewScoreUpdatedEventHandler(quizSessionManager *managers.QuizSession) *ScoreUpdatedEventHandler {
	return &ScoreUpdatedEventHandler{quizSessionManager: quizSessionManager}
}

func (q *ScoreUpdatedEventHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println("received score updated event:", string(msg.Value))
		event := &models.ScoreUpdatedEvent{}
		if err := json.Unmarshal(msg.Value, event); err != nil {
			fmt.Println("error parsing score updated event", err, string(msg.Value))
		}
		if err := q.handleMsg(event); err != nil {
			fmt.Println("error handling score updated event", err, string(msg.Value))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}

func (q *ScoreUpdatedEventHandler) handleMsg(event *models.ScoreUpdatedEvent) error {
	return q.quizSessionManager.OnScoreUpdated(event)
}

func (q *ScoreUpdatedEventHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (q *ScoreUpdatedEventHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}
