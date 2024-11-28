package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"quiz/configs"
	"quiz/consumers"
	"quiz/core/managers"
	"quiz/datastore"
	"quiz/httphandlers"
	"quiz/websocket"
	"quiz/websocket/socket"
)

func main() {
	quizSessionManager := managers.NewQuizSessionManager()

	consumers.Consume(configs.QuizProgressedTopic, consumers.NewQuizProgressedEventHandler(quizSessionManager))
	consumers.Consume(configs.ScoreUpdatedTopic, consumers.NewScoreUpdatedEventHandler(quizSessionManager))

	httphandlers.StartHttpServer(quizSessionManager)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		fmt.Println("Signal received, releasing lock")
		datastore.ReleaseAllLocks()
		os.Exit(0)
	}()

	server := socket.StartServer()
	websocket.ListenAndHandleEvent(quizSessionManager, server)
}
