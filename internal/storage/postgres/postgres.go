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

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
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
	UPDATE users SET full_name = $1, user_group = $2 WHERE telegram_id = $3;
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

// CheckRole возвращает роль пользователся в БД. Возвращает nil, если роли нет.
func (s *Storage) CheckRole(ctx context.Context, user *models.UserModel) (*string, error) {
	query := `
		SELECT role FROM users WHERE telegram_id = $1
	`

	var role *string

	err := s.pool.QueryRow(ctx, query, user.TelegramID).Scan(&role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// AddQuiz добавляет информацию о новом квизе в БД
func (s *Storage) AddQuiz(ctx context.Context, quizInfo models.InfoModel) error {
	query := `
	INSERT INTO quizzes_info (name, file, creator) VALUES ($1, $2, $3) ON CONFLICT name DO NOTHING
	`

	cmtTag, err := s.pool.Exec(ctx, query, quizInfo.Name, quizInfo.File, quizInfo.AuthorUsername)
	if cmtTag.RowsAffected() == 0 {
		return storage.ErrQuizAlreadyExists
	}

	return err
}

// EditQuiz редактирует существующий квиз
func (s *Storage) EditQuiz(
	ctx context.Context,
	requestUser *models.UserModel,
	quizInfo models.InfoModel,
) error {
	query := `
	UPDATE quizzes_info SET file = $1 WHERE name = $2
	`

	_, err := s.pool.Exec(ctx, query, quizInfo.File, quizInfo.Name)
	return err
}
