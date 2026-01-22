package auth

import (
	"context"

	"github.com/letsssgooo/quizBot/internal/storage"
)

// Auth определяет интерфейс для авторизации
type Auth interface {
	// CreateUser создает нового пользотеля в БД
	CreateUser(ctx context.Context, st storage.Storage, username, message string) error

	// AddRole добавляет пользователю роль
	AddRole(ctx context.Context, st storage.Storage, username, message string) error

	// AddGroup добавляет пользователю группу
	AddGroup(ctx context.Context, st storage.Storage, username, message string) error

	// CheckRole проверяет роль пользователя
	CheckRole(ctx context.Context, st storage.Storage, username, role string) (bool, error)
}
