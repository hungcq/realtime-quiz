package configs

import (
	"time"
)

var KafkaBrokerAddress = []string{"127.0.0.1:9091", "127.0.0.1:9092"}

const RedisAddress = "127.0.0.1:6379"

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
	AnswerChecked SocketEvent = "answer_checked"
	QuestionEnded SocketEvent = "question_ended"
	ScoreUpdated  SocketEvent = "score_updated"
	QuizEnded     SocketEvent = "quiz_ended"
	QuizData      SocketEvent = "quiz_data"
)
