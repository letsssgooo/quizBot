package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/letsssgooo/quizBot/internal/domain/models"
	"github.com/letsssgooo/quizBot/internal/storage"
)

// Storage реализует интерфейс storage.Storage на базе postgreSQL
type Storage struct {
	pool *pgxpool.Pool
}

// NewStorage создает пулл соединение и возвращает *Storage
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

// CreateUser создает нового пользователя в базе данных.
// Возвращает storage.ErrUserAlreadyExists, если пользователь уже в БД.
func (s *Storage) CreateUser(ctx context.Context, user *models.UserModel) error {
	query := `
	INSERT INTO users (telegram_id, created_at) VALUES ($1, $2) ON CONFLICT (telegram_id) DO NOTHING;
	`

	cmdTag, err := s.pool.Exec(ctx, query, user.TelegramID, user.CreatedAt)

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrUserAlreadyExists
	}

	return err
}

// UpdateStudentData обновляет данные о студенте в БД
func (s *Storage) UpdateStudentData(ctx context.Context, user *models.UserModel) error {
	query := `
	UPDATE users SET full_name = $1, group = $2 WHERE telegram_id = $3;
	`

	_, err := s.pool.Exec(ctx, query, user.FullName, user.Group, user.TelegramID)

	return err
}

// AddRole добавляет студенту роль в БД
func (s *Storage) AddRole(ctx context.Context, user *models.UserModel) error {
	query := `
	UPDATE users SET role = $1 WHERE telegram_id = $2;
	`

	_, err := s.pool.Exec(ctx, query, user.Role, user.TelegramID)

	return err
}

// CheckRole проверяет роль пользователся в БД
func (s *Storage) CheckRole(ctx context.Context, user *models.UserModel) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1 AND role = $2)
	`

	var hasRole bool

	err := s.pool.QueryRow(ctx, query, user.TelegramID, user.Role).Scan(&hasRole)
	if err != nil {
		return false, err
	}

	return hasRole, nil
}
