package auth

import (
	"context"
	"errors"

	"github.com/letsssgooo/quizBot/internal/storage"
)

// Auth определяет интерфейс для авторизации
type Auth interface {
	// CreateUser создает нового пользотеля в БД
	CreateUser(ctx context.Context, st storage.Storage, telegramID int64) error

	// UpdateStudentData обновляет данные студента (фио и группа)
	UpdateStudentData(ctx context.Context, st storage.Storage, telegramID int64, message []string) error

	// AddRole добавляет пользователю роль
	AddRole(ctx context.Context, st storage.Storage, telegramID int64, message string) error

	// CheckRole проверяет роль пользователя
	CheckRole(ctx context.Context, st storage.Storage, telegramID int64, role string) (bool, error)
}

// Ошибки авторизации
var ErrValidation = errors.New("validation error")
