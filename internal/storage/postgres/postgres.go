package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/letsssgooo/quizBot/internal/domain/models"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) SaveFullName(ctx context.Context, user *models.UserModel) error {
	query := `
	INSERT INTO users (username, full_name, created_at) VALUES ($1, $2, $3)
	`

	_, err := s.pool.Exec(ctx, query, user.Username, user.FullName, user.CreatedAt)

	return err
}

func (s *Storage) AddRole(ctx context.Context, user *models.UserModel) error {
	// add role
	return nil
}

func (s *Storage) CheckRole(ctx context.Context, user *models.UserModel) (bool, error) {
	// check role
	return false, nil
}

func (s *Storage) AddGroup(ctx context.Context, user *models.UserModel) error {
	// add group
	return nil
}
