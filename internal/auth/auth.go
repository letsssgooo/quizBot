package auth

import (
	"context"
	"errors"

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
	err = st.SaveFullName(ctx, username, fullName)
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
	err = st.AddRole(ctx, username, role)
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
	err = st.AddGroup(ctx, username, group)
	if err != nil {
		return err
	}
	return nil
}

func (q *BotAuth) CheckRole(ctx context.Context, st storage.Storage, username, role string) (bool, error) {
	hasRole, err := st.CheckRole(ctx, username, role)
	if err != nil {
		return false, err
	}
	return hasRole, nil
}
