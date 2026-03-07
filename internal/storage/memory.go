package storage

import (
	"context"
	"time"
)

const (
	CacheKeyPrefix = "quiz_id:%s:scores"
	CacheTime = time.Hour * 24
)

type Cache interface {
	Close() error

	// SetStudentScore устанавливает ответ студента
	SetStudentScore(
		ctx context.Context,
		quizID string,
		studentID int64,
		score int,
		ttl time.Duration,
	) error

	// GetQuizScores возвращает все ответы студентов
	GetQuizScores(
		ctx context.Context,
		quizID string,
	) (map[int]int, error)
}