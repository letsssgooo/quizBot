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

	// GetUserID возвращает ID пользователя по его TelegramID. Возвращает 0, если пользователя нет в БД.
	GetUserID(ctx context.Context, user *models.UserModel) (int, error)

	// UpdateStudentData обновляет данные студента (фио и группа)
	UpdateStudentData(ctx context.Context, user *models.UserModel) error

	// AddRole добавляет пользотелю роль
	AddRole(ctx context.Context, user *models.UserModel) error

	// CheckRole возвращает роль у существующего пользователя. Возвращает nil, если роли нет.
	CheckRole(ctx context.Context, user *models.UserModel) (*string, error)

	// CheckQuizExistence проверяет существования квиза
	CheckQuizExistence(
		ctx context.Context,
		quizInfo *models.InfoModel,
		user *models.UserModel,
	) bool

	// UpdateQuizInfo обвновляет информацию о новом квизе в БД
	UpdateQuizInfo(ctx context.Context, quizInfo *models.InfoModel) error

	// GetQuizInfo возвращает информацию о квизе по его названию и ID владельца
	GetQuizInfo(
		ctx context.Context,
		quizInfo *models.InfoModel,
	) ([]byte, error)

	// GetQuizzesNames возвращает список названий квизов, созданных пользователем
	GetQuizzesNames(
		ctx context.Context,
		quizInfo *models.InfoModel,
	) ([]string, error)

	// DeleteQuiz удаляет квиз из БД по названию и ID владельца. Возвращает storage.ErrQuizNotFound, если квиза нет в БД.
	DeleteQuiz(ctx context.Context, quizInfo *models.InfoModel) error

	// GetQuizID возвращает ID квиза по названию и ID владельца. Возвращает 0, если квиза нет в БД.
	GetQuizID(ctx context.Context, quizInfo *models.InfoModel) (int, error)

	// SetRun сохраняет запуск квиза в БД и возвращает ID запуска. Возвращает storage.ErrStatisticAlreadyExists, если статистика по квизу уже есть в БД.
	SetRun(ctx context.Context, run *models.HistoryModel) (int, error)

	// GetRuns возвращает запуск по ID. Возвращает storage.ErrQuizNotFound, если статистики по квизу нет в БД.
	GetRuns(ctx context.Context, name string) ([]*models.HistoryModel, error)

	// GetRunsStatistic возвращает список со статистикой запусков квиза по названию квиза
	GetRunsStatistic(
		ctx context.Context,
		user *models.UserModel,
	) ([]*models.StatisticModel, error)

	// SaveRunStatistic сохраняет статистику запуска квиза в БД. Возвращает storage.ErrStatisticAlreadyExists, если статистика по квизу уже есть в БД.
	SaveRunStatistic(
		ctx context.Context,
		statistic *models.StatisticModel,
		scores map[int]int,
	) error
}

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrIncorrectOwner         = errors.New("not the owner of the quiz")
	ErrQuizNotFound           = errors.New("quiz not found")
	ErrStatisticAlreadyExists = errors.New("statistic already exists")
)
