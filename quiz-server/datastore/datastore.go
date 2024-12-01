package datastore

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"quiz/configs"
	"quiz/core/models"
)

var client = redis.NewClient(&redis.Options{
	Addr: configs.RedisAddress,
})

var ErrQuizInProgress = errors.New("quiz already in progress")

var ErrUserInQuiz = errors.New("user already in quiz")

func MarkQuizAsInProgress(ctx context.Context, quizId models.QuizId) error {
	ok, err := client.SetNX(ctx, quizId.GetLockKey(), "locked", configs.QuizMaxDuration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrQuizInProgress
	}
	return nil
}

func MarkQuizAsFinished(ctx context.Context, quizId models.QuizId) error {
	res, err := client.Del(ctx, quizId.GetLockKey()).Result()
	if err != nil {
		return err
	}
	// If the delete result is 0, the key didn't exist (lock might have expired or not set)
	if res == 0 {
		fmt.Println("lock does not exist or already released", quizId)
		return nil
	}
	return nil
}

func CheckQuizInProgress(ctx context.Context, quizId models.QuizId) error {
	exists, err := client.Exists(ctx, quizId.GetLockKey()).Result()
	if err != nil {
		return fmt.Errorf("get lock error: %w", err)
	}
	if exists > 0 {
		return ErrQuizInProgress
	}
	return nil
}

func MarkUserAsInQuiz(ctx context.Context, quizId models.QuizId, username models.Username) error {
	key := fmt.Sprintf("user_in_quiz:%d:%s", quizId, username)
	ok, err := client.SetNX(ctx, key, "locked", configs.QuizMaxDuration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserInQuiz
	}
	return nil
}

func MarkUserAsNotInQuiz(ctx context.Context, quizId models.QuizId, username models.Username) error {
	res, err := client.Del(ctx, fmt.Sprintf("user_in_quiz:%d:%s", quizId, username)).Result()
	if err != nil {
		return err
	}
	// If the delete result is 0, the key didn't exist (lock might have expired or not set)
	if res == 0 {
		fmt.Println("lock does not exist or already released", quizId)
		return nil
	}
	return nil
}

func AddOrUpdateUserScore(ctx context.Context, quizId models.QuizId, username models.Username, dScore int) (int, error) {
	newScore, err := client.ZIncrBy(ctx, quizId.GetLeaderboardKey(), float64(dScore), username.String()).Result()
	return int(newScore), err
}

// GetLeaderboard retrieves the top N players from the leaderboard.
func GetLeaderboard(ctx context.Context, quizId models.QuizId, count int) ([]models.UserScore, error) {
	zres, err := client.ZRevRangeWithScores(ctx, quizId.GetLeaderboardKey(), 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}
	res := lo.Map(zres, func(item redis.Z, index int) models.UserScore {
		return models.UserScore{
			Username: models.Username(item.Member.(string)),
			Score:    models.Score(item.Score),
		}
	})
	res = lo.Filter(res, func(item models.UserScore, index int) bool {
		return item.Score > 0
	})
	return res, nil
}

func CleanUpUserScores(ctx context.Context, quizId models.QuizId) error {
	fmt.Println("cleaning up user scores of quiz:", quizId)
	err := client.Del(ctx, quizId.GetLeaderboardKey()).Err()
	if err != nil {
		return fmt.Errorf("error deleting leaderboard: %w", err)
	}
	return nil
}

func GetPlayerRank(ctx context.Context, quizId models.QuizId, username models.Username) (rank int64, score float64, err error) {
	score, err = client.ZScore(ctx, quizId.GetLeaderboardKey(), username.String()).Result()
	if err != nil {
		return 0, 0, err
	}

	rank, err = client.ZRevRank(ctx, quizId.GetLeaderboardKey(), username.String()).Result()
	if err != nil {
		return 0, 0, err
	}

	return rank + 1, score, nil
}
