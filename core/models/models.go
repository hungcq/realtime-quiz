package models

import (
	"fmt"
	"strconv"

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
	Participants         map[UserId]*UserSession
	CurrentQuestionIndex int
}

type UserSession struct {
	Socket            socketio.ServerSocket
	AnsweredQuestions map[int]bool
}

type QuizProgressedEvent struct {
	QuizId        QuizId      `json:"quiz_id"`
	QuestionIndex int         `json:"question_index"`
	EventType     EventType   `json:"event_type"`
	Leaderboard   []UserScore `json:"leaderboard"`
}

type ScoreUpdatedEvent struct {
	QuizId      QuizId      `json:"quiz_id"`
	Leaderboard []UserScore `json:"leaderboard"`
}

type QuestionAnsweredPayload struct {
	QuizId        QuizId `json:"quiz_id"`
	QuestionIndex int    `json:"question_index"`
	AnswerIndex   int    `json:"answer_index"`
}

type UserScore struct {
	UserId UserId
	Score  Score
}

type AnswerQuestionResult struct {
	CorrectAnswerIndex int
	NewScore           Score
	Leaderboard        []UserScore
}

type UserId int

type Score int

type QuizId int

type EventType int

const (
	QuestionStarted EventType = iota + 1
	ScoreUpdated
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

func (u UserId) String() string {
	return fmt.Sprintf("%d", u)
}

func (q QuizId) String() string {
	return fmt.Sprintf("%d", q)
}

func (q QuizId) GetLeaderboardKey() string {
	return fmt.Sprintf("quiz:%d", q)
}

func UserIdFromStr(userId string) UserId {
	val, _ := strconv.Atoi(userId)
	return UserId(val)
}
