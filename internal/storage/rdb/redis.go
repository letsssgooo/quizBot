package rdb

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/letsssgooo/quizBot/internal/storage"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr        string
	Password    string
	User        string
	DB          int
	MaxRetries  int
	DialTimeout time.Duration
	Timeout     time.Duration
}

type RedisConfig struct {
	raw    Config
	client *redis.Client
}

func NewClient(ctx context.Context, cfg Config) (*RedisConfig, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		Username:     cfg.User,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisConfig{client: client, raw: cfg}, nil
}

func getKey(quizID string) string {
	return fmt.Sprintf(storage.CacheKeyPrefix, quizID)
}

func (r *RedisConfig) Close() error {
	return r.client.Close()
}

// SetStudentScore устанавливает ответ студента
func (r *RedisConfig) SetStudentScore(
	ctx context.Context,
	quizID string,
	studentID int64,
	score int,
	ttl time.Duration,
) error {
	key := getKey(quizID)
	if err := r.client.HSet(ctx, key, studentID, score).Err(); err != nil {
		return err
	}

	return r.client.ExpireNX(ctx, key, ttl).Err()
}

// GetQuizScores возвращает все ответы студентов
func (r *RedisConfig) GetQuizScores(
	ctx context.Context,
	quizID string,
) (map[int]int, error) {
	key := getKey(quizID)

	results, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	scores := make(map[int]int, len(results))

	for studentIDStr, scoreStr := range results {
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid studentID in redis: %w", err)
		}

		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			return nil, fmt.Errorf("invalid score in redis: %w", err)
		}

		slog.Debug("cache data", "tg_id", studentID, "score", score)

		scores[studentID] = score
	}

	return scores, nil
}
