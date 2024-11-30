package managers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	socketio "github.com/karagenc/socket.io-go"
	"quiz/configs"
	"quiz/core/data"
	"quiz/core/models"
	"quiz/datastore"
	"quiz/event_publisher"
	"quiz/websocket/socket"
)

var mutex sync.Mutex

type QuizSession struct {
	quizzesInProgress map[models.QuizId]*models.OngoingQuiz
}

func NewQuizSessionManager() *QuizSession {
	return &QuizSession{
		quizzesInProgress: make(map[models.QuizId]*models.OngoingQuiz),
	}
}

var quizNotFoundError = errors.New("quiz not found")

var QuizInProgressError = errors.New("quiz in progress")

func StartQuiz(ctx context.Context, quizId models.QuizId) error {
	if err := datastore.MarkQuizAsInProgress(ctx, quizId); err != nil {
		if errors.Is(err, datastore.ErrQuizInProgress) {
			return QuizInProgressError
		}
		return err
	}
	quizProgressed := &models.QuizProgressedEvent{
		QuizId:    quizId,
		EventType: models.QuizStarted,
	}
	return event_publisher.Publish(configs.QuizProgressedTopic, quizId.String(), quizProgressed)
}

func StartNewQuestion(ctx context.Context, quizId models.QuizId, questionIndex int) error {
	fmt.Println("start new question", quizId, questionIndex)
	topUsers, err := datastore.GetLeaderboard(ctx, quizId, configs.LeaderboardSize)
	if err != nil {
		return err
	}
	quizProgressed := &models.QuizProgressedEvent{
		QuizId:        quizId,
		QuestionIndex: questionIndex,
		Leaderboard:   topUsers,
		EventType:     models.QuestionStarted,
	}
	if err = event_publisher.Publish(configs.QuizProgressedTopic, quizId.String(), quizProgressed); err != nil {
		return err
	}
	return nil
}

func EndQuiz(ctx context.Context, quizId models.QuizId) error {
	fmt.Println("end quiz", quizId)
	if err := datastore.MarkQuizAsFinished(ctx, quizId); err != nil {
		return err
	}
	topUsers, err := datastore.GetLeaderboard(ctx, quizId, configs.LeaderboardSize)
	if err != nil {
		return err
	}
	if err = datastore.CleanUpUserScores(ctx, quizId); err != nil {
		return err
	}
	quizProgressed := &models.QuizProgressedEvent{
		QuizId:      quizId,
		EventType:   models.QuizEnded,
		Leaderboard: topUsers,
	}
	return event_publisher.Publish(configs.QuizProgressedTopic, quizId.String(), quizProgressed)
}

func (m *QuizSession) JoinQuiz(ctx context.Context, quizId models.QuizId, userId models.UserId, socket socketio.ServerSocket) (*models.Quiz, error) {
	quiz := data.QuizData[quizId]
	if quiz == nil {
		return nil, quizNotFoundError
	}

	err := datastore.CheckQuizInProgress(ctx, quizId)
	if !errors.Is(err, datastore.ErrQuizInProgress) {
		fmt.Println("check quiz in progress error", err)
		return nil, errors.New("quiz haven't been started")
	}

	if err = datastore.MarkUserAsInQuiz(ctx, quizId, userId); err != nil {
		return nil, err
	}

	mutex.Lock()
	ongoingQuiz := m.quizzesInProgress[quizId]
	if ongoingQuiz == nil {
		ongoingQuiz = &models.OngoingQuiz{
			Id:                   quiz.Id,
			CurrentQuestionIndex: -1,
			Participants:         map[models.UserId]*models.UserSession{},
		}
		m.quizzesInProgress[quizId] = ongoingQuiz
	}
	ongoingQuiz.Participants[userId] = &models.UserSession{
		Socket:            socket,
		AnsweredQuestions: map[int]bool{},
	}
	mutex.Unlock()

	return quiz.FilterAnswers(), nil
}

func (m *QuizSession) AnswerQuestion(
	s socketio.ServerSocket, quizId models.QuizId, questionIndex, answerIndex int,
) (*models.AnswerQuestionResult, error) {
	ongoingQuiz := m.quizzesInProgress[quizId]
	if ongoingQuiz == nil {
		return nil, errors.New("quiz haven't been started")
	}

	var userId models.UserId = -1
	for k, v := range ongoingQuiz.Participants {
		if v.Socket == s {
			userId = k
		}
	}
	if userId == -1 {
		return nil, errors.New("user hasn't connected")
	}

	quiz := m.quizzesInProgress[quizId]
	if quiz == nil {
		return nil, quizNotFoundError
	}
	if quiz.CurrentQuestionIndex != questionIndex {
		return nil, fmt.Errorf("question is not in progress: %d, %d", quiz.CurrentQuestionIndex, questionIndex)
	}

	session := quiz.Participants[userId]
	if session == nil {
		return nil, fmt.Errorf("user does not exist")
	}
	if session.AnsweredQuestions[questionIndex] {
		return nil, fmt.Errorf("question already answered")
	}

	question := data.QuizData[quiz.Id].Questions[questionIndex]
	dScore := 0
	if answerIndex == question.CorrectAnswerIndex {
		dScore = 1
	}
	ctx := context.Background()
	newScore, err := datastore.AddOrUpdateUserScore(ctx, quizId, userId, dScore)
	if err != nil {
		return nil, fmt.Errorf("error adding new quiz score: %w", err)
	}
	var leaderboard []models.UserScore
	if dScore > 0 {
		leaderboard, err = datastore.GetLeaderboard(ctx, quizId, configs.LeaderboardSize)
		event := &models.ScoreUpdatedEvent{
			QuizId:      quizId,
			UserId:      userId,
			Leaderboard: leaderboard,
		}
		if err = event_publisher.Publish(configs.ScoreUpdatedTopic, quizId.String(), event); err != nil {
			fmt.Println("error publishing quiz score updated event", err)
		}
	}
	return &models.AnswerQuestionResult{
		CorrectAnswerIndex: question.CorrectAnswerIndex,
		NewScore:           models.Score(newScore),
		Leaderboard:        leaderboard,
	}, nil
}

func (m *QuizSession) OnScoreUpdated(event *models.ScoreUpdatedEvent) error {
	ongoingQuiz := m.quizzesInProgress[event.QuizId]
	if ongoingQuiz == nil {
		fmt.Println("quiz haven't been started")
		return nil
	}
	socket.NotifyScoreUpdated(event.QuizId, event.UserId, event.Leaderboard)
	return nil
}

func (m *QuizSession) OnQuizProgressed(event *models.QuizProgressedEvent) error {
	switch event.EventType {
	case models.QuizStarted:
		return m.onQuizStarted(event)
	case models.QuestionStarted:
		return m.onQuestionStarted(event)
	case models.QuizEnded:
		return m.onQuizEnded(event)
	default:
		fmt.Println("unknown quiz event")
		return nil
	}
}

func (m *QuizSession) onQuizStarted(event *models.QuizProgressedEvent) error {
	ongoingQuiz := m.quizzesInProgress[event.QuizId]
	if ongoingQuiz != nil {
		fmt.Println("quiz has already been started")
		return nil
	}
	ongoingQuiz = &models.OngoingQuiz{
		Id:                   event.QuizId,
		Participants:         map[models.UserId]*models.UserSession{},
		CurrentQuestionIndex: -1, // for pending period
	}
	// handle race condition
	mutex.Lock()
	m.quizzesInProgress[event.QuizId] = ongoingQuiz
	mutex.Unlock()
	return nil
}

func (m *QuizSession) onQuestionStarted(event *models.QuizProgressedEvent) error {
	ongoingQuiz := m.quizzesInProgress[event.QuizId]
	if ongoingQuiz == nil {
		fmt.Println("quiz hasn't been started")
		return nil
	}
	ongoingQuiz.CurrentQuestionIndex = event.QuestionIndex
	socket.NotifyQuestionEnded(ongoingQuiz.Id, ongoingQuiz.CurrentQuestionIndex, event.Leaderboard)
	return nil
}

func (m *QuizSession) onQuizEnded(event *models.QuizProgressedEvent) error {
	ongoingQuiz := m.quizzesInProgress[event.QuizId]
	if ongoingQuiz == nil {
		fmt.Println("quiz haven't been started")
		return nil
	}
	for userId := range ongoingQuiz.Participants {
		if err := datastore.MarkUserAsNotInQuiz(context.Background(), event.QuizId, userId); err != nil {
			fmt.Println("error marking quiz as not-in-quiz")
		}
	}
	mutex.Lock()
	delete(m.quizzesInProgress, event.QuizId)
	mutex.Unlock()
	socket.NotifyQuizEnded(event.QuizId, event.Leaderboard)
	return nil
}
