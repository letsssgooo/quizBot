//go:build !solution

package storage

import (
	"context"

	"gitlab.com/slon/shad-go/Exam-1-QuizBot/quizbot/internal/quiz"
)

// MemoryStorage реализует Storage в памяти.
type MemoryStorage struct {
	// TODO: добавьте необходимые поля
}

// NewMemoryStorage создаёт новый MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// SaveQuiz сохраняет квиз.
func (s *MemoryStorage) SaveQuiz(ctx context.Context, q *quiz.Quiz) error {
	panic("not implemented")
}

// GetQuiz возвращает квиз по ID.
func (s *MemoryStorage) GetQuiz(ctx context.Context, id string) (*quiz.Quiz, error) {
	panic("not implemented")
}

// ListQuizzes возвращает список квизов пользователя.
func (s *MemoryStorage) ListQuizzes(ctx context.Context, ownerID int64) ([]*quiz.Quiz, error) {
	panic("not implemented")
}

// DeleteQuiz удаляет квиз.
func (s *MemoryStorage) DeleteQuiz(ctx context.Context, id string) error {
	panic("not implemented")
}

// SaveRun сохраняет запуск квиза.
func (s *MemoryStorage) SaveRun(ctx context.Context, run *quiz.QuizRun) error {
	panic("not implemented")
}

// GetRun возвращает запуск по ID.
func (s *MemoryStorage) GetRun(ctx context.Context, id string) (*quiz.QuizRun, error) {
	panic("not implemented")
}

// ListRuns возвращает список запусков квиза.
func (s *MemoryStorage) ListRuns(ctx context.Context, quizID string) ([]*quiz.QuizRun, error) {
	panic("not implemented")
}

// UpdateRun обновляет данные запуска.
func (s *MemoryStorage) UpdateRun(ctx context.Context, run *quiz.QuizRun) error {
	panic("not implemented")
}
