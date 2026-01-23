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
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.ConnectConfig(ctx, cfg)
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
	query := `
	UPDATE users SET role = $1 WHERE username = $2
	`

	_, err := s.pool.Exec(ctx, query, user.Role, user.Username)
	return err
}

func (s *Storage) CheckRole(ctx context.Context, user *models.UserModel) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1 AND role = $2)
	`

	var hasRole bool
	err := s.pool.QueryRow(ctx, query, user.Username, user.Role).Scan(&hasRole)
	if err != nil {
		return false, err
	}
	return hasRole, nil
}

func (s *Storage) AddGroup(ctx context.Context, user *models.UserModel) error {
	query := `
	UPDATE users SET student_group = $1 WHERE username = $2
	`

	_, err := s.pool.Exec(ctx, query, user.Group, user.Username)
	return err
}
