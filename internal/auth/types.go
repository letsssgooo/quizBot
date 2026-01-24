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

	// CheckRole возвращает роль у существующего пользователя. Возвращает nil, если роли нет.
	CheckRole(ctx context.Context, st storage.Storage, telegramID int64) (*string, error)
}

// Ошибки авторизации
var ErrValidation = errors.New("validation error")

// Роли
var (
	RoleLecturer = "lecturer"
	RoleStudent  = "student"
)
