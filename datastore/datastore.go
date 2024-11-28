package datastore

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"quiz/configs"
	"quiz/core/models"
)

var client = redis.NewClient(&redis.Options{
	Addr: configs.RedisAddress,
})

var locker = redislock.New(client)

var lockByQuizId = map[models.QuizId]*redislock.Lock{}

var mutex sync.Mutex

var ErrQuizInProgress = errors.New("quiz already in progress")

func MarkQuizAsInProgress(ctx context.Context, quizId models.QuizId) error {
	lockKey := fmt.Sprintf("quiz_in_progress:%d", quizId)
	lock, err := locker.Obtain(ctx, lockKey, configs.QuizMaxDuration, nil)
	if errors.Is(err, redislock.ErrNotObtained) {
		return ErrQuizInProgress
	} else if err != nil {
		return err
	}
	mutex.Lock()
	defer mutex.Unlock()
	lockByQuizId[quizId] = lock
	return nil
}

func MarkQuizAsFinished(ctx context.Context, id models.QuizId) error {
	// handle race condition
	mutex.Lock()
	defer mutex.Unlock()

	lock := lockByQuizId[id]
	if lock == nil {
		fmt.Println("quiz not locked")
		return nil
	}
	delete(lockByQuizId, id)

	err := lock.Release(ctx)
	if err != nil {
		fmt.Println("failed to release quiz:", err)
	}
	return nil
}

func ReleaseAllLocks() {
	for _, lock := range lockByQuizId {
		lock.Release(context.Background())
	}
}

func AddOrUpdateUserScore(ctx context.Context, quizId models.QuizId, userId models.UserId, dScore int) (int, error) {
	newScore, err := client.ZIncrBy(ctx, quizId.GetLeaderboardKey(), float64(dScore), userId.String()).Result()
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
			UserId: models.UserIdFromStr(item.Member.(string)),
			Score:  models.Score(item.Score),
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

func GetPlayerRank(ctx context.Context, quizId models.QuizId, userId models.UserId) (rank int64, score float64, err error) {
	score, err = client.ZScore(ctx, quizId.GetLeaderboardKey(), userId.String()).Result()
	if err != nil {
		return 0, 0, err
	}

	rank, err = client.ZRevRank(ctx, quizId.GetLeaderboardKey(), userId.String()).Result()
	if err != nil {
		return 0, 0, err
	}

	return rank + 1, score, nil
}
