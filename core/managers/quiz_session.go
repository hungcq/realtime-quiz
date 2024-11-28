package managers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

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

func (m *QuizSession) StartQuiz(ctx context.Context, quizId models.QuizId) error {
	quiz := data.QuizData[quizId]
	if quiz == nil {
		return errors.New("quiz not found")
	}

	if err := datastore.MarkQuizAsInProgress(ctx, quizId); err != nil {
		if errors.Is(err, datastore.ErrQuizInProgress) {
			return errors.New("start quiz: quiz already in progress")
		}
		return err
	}

	ongoingQuiz := &models.OngoingQuiz{
		Id:                   quizId,
		Participants:         map[models.UserId]*models.UserSession{},
		CurrentQuestionIndex: -1, // for pending period
	}
	// handle race condition
	mutex.Lock()
	m.quizzesInProgress[quizId] = ongoingQuiz
	mutex.Unlock()

	m.startCoordinatorLoop(ongoingQuiz, quiz)

	quizProgressed := &models.QuizProgressedEvent{
		QuizId:        quizId,
		QuestionIndex: -1,
		EventType:     models.QuestionStarted,
	}
	return event_publisher.Publish(configs.QuizProgressedTopic, quizId.String(), quizProgressed)
}

func (m *QuizSession) JoinQuiz(ctx context.Context, quizId models.QuizId, userId models.UserId, socket socketio.ServerSocket) (*models.Quiz, error) {
	quiz := data.QuizData[quizId]
	if quiz == nil {
		return nil, errors.New("quiz not found")
	}

	err := datastore.MarkQuizAsInProgress(ctx, quizId)
	if !errors.Is(err, datastore.ErrQuizInProgress) {
		return nil, errors.New("quiz haven't been started")
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
		return nil, fmt.Errorf("quiz not found")
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

func (m *QuizSession) OnQuizProgressed(event *models.QuizProgressedEvent) error {
	switch event.EventType {
	case models.QuizEnded:
		ongoingQuiz := m.quizzesInProgress[event.QuizId]
		if ongoingQuiz == nil {
			fmt.Println("quiz haven't been started")
			return nil
		}
		mutex.Lock()
		delete(m.quizzesInProgress, event.QuizId)
		mutex.Unlock()
		socket.NotifyQuizEnded(event.QuizId, event.Leaderboard)
		return nil
	case models.QuestionStarted:
		ongoingQuiz := m.quizzesInProgress[event.QuizId]
		if ongoingQuiz == nil {
			fmt.Println("quiz haven't been started")
			return nil
		}
		ongoingQuiz.CurrentQuestionIndex = event.QuestionIndex
		if ongoingQuiz.CurrentQuestionIndex >= 0 { // ignore pending case where there is no ongoing question
			socket.NotifyQuestionEnded(ongoingQuiz.Id, ongoingQuiz.CurrentQuestionIndex, event.Leaderboard)
		}
		return nil
	default:
		fmt.Println("unknown quiz event")
		return nil
	}
}

func (m *QuizSession) OnScoreUpdated(event *models.ScoreUpdatedEvent) error {
	ongoingQuiz := m.quizzesInProgress[event.QuizId]
	if ongoingQuiz == nil {
		fmt.Println("quiz haven't been started")
		return nil
	}
	socket.NotifyScoreUpdated(event.QuizId, event.Leaderboard)
	return nil
}

func (m *QuizSession) startCoordinatorLoop(ongoingQuiz *models.OngoingQuiz, quiz *models.Quiz) {
	const maxRetries = 5
	ticker := time.NewTicker(configs.DefaultQuestionTime)
	pendingPeriodPassed := false
	go func() {
		for range ticker.C {
			fmt.Println("quiz loop")
			if !pendingPeriodPassed {
				pendingPeriodPassed = true
				continue
			}
			for range maxRetries {
				if ongoingQuiz.CurrentQuestionIndex == len(quiz.Questions)-1 {
					if err := m.endQuiz(quiz); err == nil {
						ticker.Stop()
						break
					} else {
						fmt.Println("finish quiz error:", err)
					}
				} else {
					if err := m.startNewQuestion(ongoingQuiz); err == nil {
						break
					} else {
						fmt.Println("start new question error:", err)
					}
				}
			}
		}
	}()
}

func (m *QuizSession) startNewQuestion(ongoingQuiz *models.OngoingQuiz) error {
	ctx := context.Background()
	fmt.Println("start new question")
	topUsers, err := datastore.GetLeaderboard(ctx, ongoingQuiz.Id, configs.LeaderboardSize)
	if err != nil {
		return err
	}
	quizProgressed := &models.QuizProgressedEvent{
		QuizId:        ongoingQuiz.Id,
		QuestionIndex: ongoingQuiz.CurrentQuestionIndex + 1,
		Leaderboard:   topUsers,
		EventType:     models.QuestionStarted,
	}
	return event_publisher.Publish(configs.QuizProgressedTopic, ongoingQuiz.Id.String(), quizProgressed)
}

func (m *QuizSession) endQuiz(quiz *models.Quiz) error {
	ctx := context.Background()
	fmt.Println("end quiz")
	if err := datastore.MarkQuizAsFinished(ctx, quiz.Id); err != nil {
		return err
	}
	topUsers, err := datastore.GetLeaderboard(ctx, quiz.Id, configs.LeaderboardSize)
	if err != nil {
		return err
	}
	if err = datastore.CleanUpUserScores(ctx, quiz.Id); err != nil {
		return err
	}
	quizProgressed := &models.QuizProgressedEvent{
		QuizId:      quiz.Id,
		EventType:   models.QuizEnded,
		Leaderboard: topUsers,
	}
	return event_publisher.Publish(configs.QuizProgressedTopic, quiz.Id.String(), quizProgressed)
}
