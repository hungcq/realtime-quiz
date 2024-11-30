package main

import (
	"quiz/configs"
	"quiz/consumers"
	"quiz/core/managers"
	"quiz/httphandlers"
	"quiz/websocket"
	"quiz/websocket/socket"
	"quiz/workflow"
)

func main() {
	c := workflow.StartWorkflowClient()
	defer c.Close()
	defer socket.Server.Close()

	quizSessionManager := managers.NewQuizSessionManager()

	consumers.Consume(configs.QuizProgressedTopic, consumers.NewQuizProgressedEventHandler(quizSessionManager))
	consumers.Consume(configs.ScoreUpdatedTopic, consumers.NewScoreUpdatedEventHandler(quizSessionManager))

	httphandlers.StartHttpServer()
	websocket.ListenAndHandleEvent(quizSessionManager)
}
