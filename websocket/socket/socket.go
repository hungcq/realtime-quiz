package socket

import (
	socketio "github.com/karagenc/socket.io-go"
	"quiz/configs"
	"quiz/core/models"
)

var server *socketio.Server

func StartServer() *socketio.Server {
	server = socketio.NewServer(&socketio.ServerConfig{})
	return server
}

func NotifyQuestionEnded(quizId models.QuizId, currentQuestionIndex int, leaderboard []models.UserScore) {
	server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.QuestionEnded), currentQuestionIndex, leaderboard)
}

func NotifyQuizEnded(quizId models.QuizId, leaderboard []models.UserScore) {
	server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.QuizEnded), leaderboard)
}

func NotifyScoreUpdated(quizId models.QuizId, leaderboard []models.UserScore) {
	server.Of("").In(socketio.Room(quizId.String())).Emit(string(configs.ScoreUpdated), leaderboard)
}
