package main

import (
	"quiz/configs"
	"quiz/consumers"
	"quiz/core/managers"
	"quiz/websocket"
	"quiz/websocket/socket"
	"quiz/workflow"
)

func main() {
	c := workflow.StartWorkflowClient()
	defer c.Close()

	quizSessionManager := managers.NewQuizSessionManager()

	consumers.Consume(configs.QuizProgressedTopic, consumers.NewQuizProgressedEventHandler(quizSessionManager))
	consumers.Consume(configs.ScoreUpdatedTopic, consumers.NewScoreUpdatedEventHandler(quizSessionManager))

	server := socket.StartServer()
	websocket.ListenAndHandleEvent(quizSessionManager, server)
}
