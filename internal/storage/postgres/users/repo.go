package users

import "context"

// UserRepo определяет интерфейс для хранения данных о пользователях.
type UserRepo interface {
	// SaveFullName сохраняет фио пользователя
	SaveFullName(ctx context.Context, username string, fullName string) error

	// AddRole добавляет пользотелю роль
	AddRole(ctx context.Context, username string, role string) error

	// CheckRole проверяет есть ли у пользователя роль role
	CheckRole(ctx context.Context, username string, role string) (bool, error)

	// AddGroup добавляет пользотелю группу
	AddGroup(ctx context.Context, username string, group string) error
}
