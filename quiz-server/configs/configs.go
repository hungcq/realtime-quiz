package configs

import (
	"os"
	"strings"
	"time"
)

var KafkaBrokerAddress = strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

var RedisAddress = os.Getenv("REDIS_ADDRESS")

var TemporalAddress = os.Getenv("TEMPORAL_ADDRESS")

func init() {
	if KafkaBrokerAddress[0] == "" {
		KafkaBrokerAddress = []string{"localhost:9092"}
	}
	if RedisAddress == "" {
		RedisAddress = "localhost:6379"
	}
	if TemporalAddress == "" {
		TemporalAddress = "localhost:7233"
	}
}

const (
	QuizProgressedTopic = "quiz_progressed"
	ScoreUpdatedTopic   = "score_updated"
)

const (
	DefaultQuestionTime = 10 * time.Second
	QuizMaxDuration     = 5 * time.Minute
	LeaderboardSize     = 5
)

type SocketEvent string

const (
	// inbound events
	JoinQuiz       SocketEvent = "join_quiz"
	AnswerQuestion SocketEvent = "answer_question"
	// outbound events
	AnswerChecked   SocketEvent = "answer_checked"
	QuestionStarted SocketEvent = "question_started"
	ScoreUpdated    SocketEvent = "score_updated"
	QuizEnded       SocketEvent = "quiz_ended"
	QuizData        SocketEvent = "quiz_data"
	Error           SocketEvent = "quiz_error"
)
