package auth

import (
	"errors"
	"time"

	"github.com/letsssgooo/quizBot/internal/storage"
)

// Auth определяет интерфейс для авторизации
type Auth interface {
	// CreateUser создает нового пользотеля в БД
	CreateUser(st storage.Storage, telegramID int64) error

	// UpdateStudentData обновляет данные студента (фио и группа)
	UpdateStudentData(st storage.Storage, telegramID int64, message []string) error

	// AddRole добавляет пользователю роль
	AddRole(st storage.Storage, telegramID int64, message string) error

	// CheckRole возвращает роль у существующего пользователя. Возвращает nil, если роли нет.
	CheckRole(st storage.Storage, telegramID int64) (*string, error)
}

// Ошибки авторизации
var ErrValidation = errors.New("validation error")

// Роли
const (
	RoleLecturer = "lecturer"
	RoleStudent  = "student"
)

// Таймаут
const timeoutAuth = 500 * time.Millisecond
