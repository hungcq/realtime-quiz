package socket

import (
	socketio "github.com/karagenc/socket.io-go"
	"quiz/configs"
	"quiz/core/models"
)

var Server *socketio.Server

func NotifyQuestionEnded(quizId models.QuizId, currentQuestionIndex int, leaderboard []models.UserScore) {
	Server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.QuestionStarted), currentQuestionIndex, leaderboard)
}

func NotifyQuizEnded(quizId models.QuizId, leaderboard []models.UserScore) {
	Server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.QuizEnded), leaderboard)
}

func NotifyScoreUpdated(quizId models.QuizId, userId models.UserId, leaderboard []models.UserScore) {
	Server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.ScoreUpdated), userId, leaderboard)
}
