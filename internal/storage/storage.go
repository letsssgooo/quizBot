//go:build !change

package storage

import (
	"github.com/letsssgooo/quizBot/internal/storage/postgres/quizzes"
	"github.com/letsssgooo/quizBot/internal/storage/postgres/users"
)

// Storage определяет интерфейс для хранения данных квизов и пользователей.
type Storage interface {
	// QuizRepo определяет интерфейс для работы с квизами
	quizzes.QuizRepo

	// UserRepo определяет интерфейс для работы с пользователями
	users.UserRepo
}
