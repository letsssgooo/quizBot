package auth

import (
	"context"
	"errors"
	"time"

	"github.com/letsssgooo/quizBot/internal/domain/models"
	"github.com/letsssgooo/quizBot/internal/storage"
)

// BotAuth реализует Auth
type BotAuth struct {
	Roles map[string]struct{}
}

// NewBotAuth создает BotAuth
func NewBotAuth() *BotAuth {
	return &BotAuth{
		Roles: map[string]struct{}{
			RoleLecturer: {},
			RoleStudent:  {},
		},
	}
}

// CreateUser создает нового пользотеля
func (q *BotAuth) CreateUser(ctx context.Context, st storage.Storage, telegramID int64) error {
	err := st.CreateUser(ctx, &models.UserModel{
		TelegramID: telegramID,
		CreatedAt:  time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateStudentData обновляет данные студента у существующего пользотеля
func (q *BotAuth) UpdateStudentData(ctx context.Context, st storage.Storage, telegramID int64, message []string) error {
	studentsData, err := ParseStudentsData(message)
	if err != nil {
		return err
	}

	err = st.UpdateStudentData(ctx, &models.UserModel{
		TelegramID: telegramID,
		FullName:   studentsData[0],
		Group:      studentsData[1],
	})
	if err != nil {
		return err
	}

	return nil
}

// AddRole добавляет роль у существующего пользотеля
func (q *BotAuth) AddRole(ctx context.Context, st storage.Storage, telegramID int64, message string) error {
	role, err := ParseRole(message)
	if err != nil {
		return err
	}

	if _, ok := q.Roles[role]; !ok {
		return errors.New("invalid role")
	}

	err = st.AddRole(ctx, &models.UserModel{
		TelegramID: telegramID,
		Role:       role,
	})
	if err != nil {
		return err
	}

	return nil
}

// CheckRole возвращает роль у существующего пользователя. Возвращает nil, если роли нет.
func (q *BotAuth) CheckRole(ctx context.Context, st storage.Storage, telegramID int64) (*string, error) {
	role, err := st.CheckRole(ctx, &models.UserModel{
		TelegramID: telegramID,
	})
	if err != nil {
		return nil, err
	}

	return role, nil
}
