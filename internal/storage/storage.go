//go:build !change

package storage

import (
	"context"

	"github.com/letsssgooo/quizBot/internal/events/engine"
)

// Storage определяет интерфейс для хранения данных квизов.
type Storage interface {
	// SaveQuiz сохраняет квиз.
	SaveQuiz(ctx context.Context, q *engine.Quiz) error

	// GetQuiz возвращает квиз по ID.
	GetQuiz(ctx context.Context, id string) (*engine.Quiz, error)

	// ListQuizzes возвращает список квизов пользователя.
	ListQuizzes(ctx context.Context, ownerID int64) ([]*engine.Quiz, error)

	// DeleteQuiz удаляет квиз.
	DeleteQuiz(ctx context.Context, id string) error

	// SaveRun сохраняет запуск квиза.
	SaveRun(ctx context.Context, run *engine.QuizRun) error

	// GetRun возвращает запуск по ID.
	GetRun(ctx context.Context, id string) (*engine.QuizRun, error)

	// ListRuns возвращает список запусков квиза.
	ListRuns(ctx context.Context, quizID string) ([]*engine.QuizRun, error)

	// UpdateRun обновляет данные запуска.
	UpdateRun(ctx context.Context, run *engine.QuizRun) error
}
