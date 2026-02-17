//go:build !change

package storage

import (
	"context"
	"errors"

	"github.com/letsssgooo/quizBot/internal/domain/models"
)

// Storage определяет интерфейс для хранения данных квизов и пользователей.
type Storage interface {
	// CreateUser сохраняет фио пользователя
	CreateUser(ctx context.Context, user *models.UserModel) error

	// UpdateStudentData обновляет данные студента (фио и группа)
	UpdateStudentData(ctx context.Context, user *models.UserModel) error

	// AddRole добавляет пользотелю роль
	AddRole(ctx context.Context, user *models.UserModel) error

	// CheckRole возвращает роль у существующего пользователя. Возвращает nil, если роли нет.
	CheckRole(ctx context.Context, user *models.UserModel) (*string, error)

	//// SaveQuiz сохраняет квиз.
	//SaveQuiz(ctx context.Context, q *engine.Quiz) error
	//
	//// GetQuiz возвращает квиз по ID.
	//GetQuiz(ctx context.Context, id string) (*engine.Quiz, error)
	//
	//// ListQuizzes возвращает список квизов пользователя.
	//ListQuizzes(ctx context.Context, ownerID int64) ([]*engine.Quiz, error)
	//
	//// DeleteQuiz удаляет квиз.
	//DeleteQuiz(ctx context.Context, id string) error
	//
	//// SaveRun сохраняет запуск квиза.
	//SaveRun(ctx context.Context, run *engine.QuizRun) error
	//
	//// GetRun возвращает запуск по ID.
	//GetRun(ctx context.Context, id string) (*engine.QuizRun, error)
	//
	//// ListRuns возвращает список запусков квиза.
	//ListRuns(ctx context.Context, quizID string) ([]*engine.QuizRun, error)
	//
	//// UpdateRun обновляет данные запуска.
	//UpdateRun(ctx context.Context, run *engine.QuizRun) error
}

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrQuizAlreadyExists = errors.New("quiz already exists")
)
