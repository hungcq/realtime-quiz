package workflow

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"quiz/configs"
	"quiz/core/managers"
	"quiz/core/models"
)

var c client.Client

func StartWorkflowClient() client.Client {
	var err error
	c, err = client.Dial(client.Options{
		HostPort: configs.TemporalAddress,
	})

	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	return c
}

const QuizTaskQueue = "QUIZ_TASK_QUEUE"

func StartQuizWorkflow(ctx context.Context, quiz *models.Quiz) error {
	options := client.StartWorkflowOptions{
		TaskQueue: QuizTaskQueue,
	}

	we, err := c.ExecuteWorkflow(ctx, options, QuizSessionWorkflow, quiz)
	if err != nil {
		return err
	}

	fmt.Printf("WorkflowID: %s RunID: %s\n", we.GetID(), we.GetRunID())

	return nil
}

type newQuestionPayload struct {
	QuizId               models.QuizId
	CurrentQuestionIndex int
}

func QuizSessionWorkflow(ctx workflow.Context, quiz *models.Quiz) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: configs.DefaultQuestionTime + // pending period
			time.Duration(len(quiz.Questions))*configs.DefaultQuestionTime + // question period
			time.Minute, // timeout period
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        5 * time.Second,
			BackoffCoefficient:     1,
			MaximumAttempts:        5,
			NonRetryableErrorTypes: []string{managers.QuizInProgressError.Error()},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	if err := workflow.ExecuteActivity(ctx, StartQuiz, quiz.Id).Get(ctx, nil); err != nil {
		return err
	}
	workflow.Sleep(ctx, configs.DefaultQuestionTime)

	for i := range quiz.Questions {
		payload := &newQuestionPayload{
			QuizId:               quiz.Id,
			CurrentQuestionIndex: i,
		}
		if err := workflow.ExecuteActivity(ctx, StartNewQuestion, payload).Get(ctx, nil); err != nil {
			return err
		}
		workflow.Sleep(ctx, configs.DefaultQuestionTime)
	}
	if err := workflow.ExecuteActivity(ctx, EndQuiz, quiz.Id).Get(ctx, nil); err != nil {
		return err
	}
	return nil
}

func StartQuiz(ctx context.Context, quizId models.QuizId) error {
	return managers.StartQuiz(ctx, quizId)
}

func StartNewQuestion(ctx context.Context, payload newQuestionPayload) error {
	return managers.StartNewQuestion(ctx, payload.QuizId, payload.CurrentQuestionIndex)
}

func EndQuiz(ctx context.Context, quizId models.QuizId) error {
	return managers.EndQuiz(ctx, quizId)
}
