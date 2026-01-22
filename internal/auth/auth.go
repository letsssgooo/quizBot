package auth

import (
	"context"
	"errors"
	"time"

	"github.com/letsssgooo/quizBot/internal/domain/models"
	"github.com/letsssgooo/quizBot/internal/storage"
)

type BotAuth struct {
	Roles map[string]struct{}
}

func NewBotAuth() *BotAuth {
	return &BotAuth{
		Roles: map[string]struct{}{
			"teacher": {},
			"student": {},
		},
	}
}

func (q *BotAuth) CreateUser(ctx context.Context, st storage.Storage, username, message string) error {
	fullName, err := ParseFullName(message)
	if err != nil {
		return err
	}

	err = st.SaveFullName(ctx, &models.UserModel{
		Username:  username,
		FullName:  fullName,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (q *BotAuth) AddRole(ctx context.Context, st storage.Storage, username, message string) error {
	role, err := ParseRole(message)
	if err != nil {
		return err
	}

	if _, ok := q.Roles[role]; !ok {
		return errors.New("invalid role")
	}

	err = st.AddRole(ctx, &models.UserModel{
		Username: username,
		Role:     role,
	})
	if err != nil {
		return err
	}

	return nil
}

func (q *BotAuth) AddGroup(ctx context.Context, st storage.Storage, username, message string) error {
	group, err := ParseGroup(message)
	if err != nil {
		return err
	}

	err = st.AddGroup(ctx, &models.UserModel{
		Username: username,
		Group:    group,
	})
	if err != nil {
		return err
	}

	return nil
}

func (q *BotAuth) CheckRole(ctx context.Context, st storage.Storage, username, role string) (bool, error) {
	hasRole, err := st.CheckRole(ctx, &models.UserModel{
		Username: username,
		Role:     role,
	})
	if err != nil {
		return false, err
	}

	return hasRole, nil
}
