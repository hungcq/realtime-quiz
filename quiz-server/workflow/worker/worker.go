package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"quiz/configs"
	"quiz/workflow"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{
		HostPort: configs.TemporalAddress,
	})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, workflow.QuizTaskQueue, worker.Options{})
	w.RegisterWorkflow(workflow.QuizSessionWorkflow)
	w.RegisterActivity(workflow.StartQuiz)
	w.RegisterActivity(workflow.StartNewQuestion)
	w.RegisterActivity(workflow.EndQuiz)

	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
