package models

import (
	"fmt"
	"sync"

	socketio "github.com/karagenc/socket.io-go"
)

type Quiz struct {
	Id        QuizId     `json:"id"`
	Questions []Question `json:"questions"`
}

type Question struct {
	Content            string `json:"content"`
	CorrectAnswerIndex int    `json:"correct_answer_index"`
}

type OngoingQuiz struct {
	Id                   QuizId
	Participants         map[Username]*UserSession
	CurrentQuestionIndex int
}

type UserSession struct {
	Socket            socketio.ServerSocket
	AnsweredQuestions map[int]bool
	Mutex             sync.Mutex
}

type QuizProgressedEvent struct {
	QuizId        QuizId      `json:"quiz_id"`
	QuestionIndex int         `json:"question_index"`
	EventType     EventType   `json:"event_type"`
	Leaderboard   []UserScore `json:"leaderboard"`
}

type ScoreUpdatedEvent struct {
	QuizId      QuizId      `json:"quiz_id"`
	Username    Username    `json:"username"`
	Leaderboard []UserScore `json:"leaderboard"`
}

type QuestionAnsweredPayload struct {
	QuizId        QuizId `json:"quiz_id"`
	QuestionIndex int    `json:"question_index"`
	AnswerIndex   int    `json:"answer_index"`
}

type UserScore struct {
	Username Username `json:"username"`
	Score    Score    `json:"score"`
}

type AnswerQuestionResult struct {
	CorrectAnswerIndex int
	NewScore           Score
	Leaderboard        []UserScore
}

type Username string

type Score int

type QuizId int

type EventType int

const (
	QuizStarted EventType = iota + 1
	QuestionStarted
	QuizEnded
)

func (q *Quiz) FilterAnswers() *Quiz {
	res := &Quiz{
		Id: q.Id,
	}
	for _, question := range q.Questions {
		res.Questions = append(res.Questions, Question{
			Content:            question.Content,
			CorrectAnswerIndex: -1,
		})
	}
	return res
}

func (u Username) String() string {
	return string(u)
}

func (q QuizId) String() string {
	return fmt.Sprintf("%d", q)
}

func (q QuizId) GetLeaderboardKey() string {
	return fmt.Sprintf("quiz:%d", q)
}

func (q QuizId) GetLockKey() string {
	return fmt.Sprintf("quiz_in_progress:%d", q)
}
