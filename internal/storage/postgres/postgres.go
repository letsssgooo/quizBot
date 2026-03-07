package postgres

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
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

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
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

// GetUserID возвращает ID пользователя по его TelegramID. Возвращает 0, если пользователя нет в БД.
func (s *Storage) GetUserID(ctx context.Context, user *models.UserModel) (int, error) {
	query := `
	SELECT id FROM users WHERE telegram_id = $1;
	`

	var id int

	err := s.pool.QueryRow(ctx, query, user.TelegramID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
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
		SELECT role FROM users WHERE telegram_id = $1;
	`

	var role *string

	err := s.pool.QueryRow(ctx, query, user.TelegramID).Scan(&role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// CheckQuizExistence проверяет существования квиза
func (s *Storage) CheckQuizExistence(
	ctx context.Context,
	quizInfo *models.InfoModel,
	user *models.UserModel,
) bool {
	query := `
	SELECT EXISTS (SELECT 1 FROM quizzes_info WHERE name = $1 AND owner_id = $2);
	`

	var exists bool

	_ = s.pool.QueryRow(ctx, query, quizInfo.Name, user.TelegramID).Scan(&exists)

	return exists
}

// UpdateQuizInfo обновляет информацию о новом квизе в БД
func (s *Storage) UpdateQuizInfo(ctx context.Context, quizInfo *models.InfoModel) error {
	queryUpdate := `
    UPDATE quizzes_info
    SET file = $1
    WHERE name = $2 AND owner_id = $3
    `

	cmdTag, err := s.pool.Exec(ctx, queryUpdate, quizInfo.File, quizInfo.Name, quizInfo.OwnerID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() > 0 {
		return nil
	}

	queryCheck := `SELECT EXISTS(SELECT 1 FROM quizzes_info WHERE name = $1)`

	var check bool

	err = s.pool.QueryRow(ctx, queryCheck, quizInfo.Name).Scan(&check)
	if err != nil {
		return err
	}

	if !check {
		queryInsert := `
		INSERT INTO quizzes_info (name, file, owner_id)
		VALUES ($1, $2, $3)
		`
		_, err = s.pool.Exec(ctx, queryInsert, quizInfo.Name, quizInfo.File, quizInfo.OwnerID)

		return err
	}

	return storage.ErrIncorrectOwner
}

// GetQuizInfo возвращает информацию (файл) о квизе по названию и ID владельца
func (s *Storage) GetQuizInfo(
	ctx context.Context,
	quizInfo *models.InfoModel,
) ([]byte, error) {
	query := `
	SELECT file FROM quizzes_info WHERE name = $1 AND owner_id = $2;
	`

	var file []byte

	err := s.pool.QueryRow(ctx, query, quizInfo.Name, quizInfo.OwnerID).Scan(&file)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// GetQuizzesNames возвращает список названий квизов, созданных пользователем
func (s *Storage) GetQuizzesNames(
	ctx context.Context,
	quizInfo *models.InfoModel,
) ([]string, error) {
	query := `
	SELECT name FROM quizzes_info WHERE owner_id = $1;
	`

	rows, err := s.pool.Query(ctx, query, quizInfo.OwnerID)
	if err != nil {
		return nil, err
	}

	quizzesNames, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}

	return quizzesNames, nil
}

// DeleteQuiz удаляет квиз из БД по названию и ID владельца. Возвращает storage.ErrQuizNotFound, если квиза нет в БД.
func (s *Storage) DeleteQuiz(ctx context.Context, quizInfo *models.InfoModel) error {
	query := `
	DELETE FROM quizzes_info WHERE name = $1 AND owner_id = $2;
	`

	cmdTag, err := s.pool.Exec(ctx, query, quizInfo.Name, quizInfo.OwnerID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return storage.ErrQuizNotFound
	}

	return nil
}

// GetQuizID возвращает ID квиза по названию и ID владельца. Возвращает 0, если квиза нет в БД.
func (s *Storage) GetQuizID(ctx context.Context, quizInfo *models.InfoModel) (int, error) {
	query := `
	SELECT id FROM quizzes_info WHERE name = $1 AND owner_id = $2;
	`
	slog.Debug("here", "quizInfo.Name", quizInfo.Name, "quizInfo.OwnerID", quizInfo.OwnerID)
	var id int
	err := s.pool.QueryRow(ctx, query, quizInfo.Name, quizInfo.OwnerID).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// SetRun сохраняет запуск квиза в БД и возвращает ID запуска. Возвращает storage.ErrStatisticAlreadyExists, если статистика по квизу уже есть в БД.
func (s *Storage) SetRun(ctx context.Context, run *models.HistoryModel) (int, error) {
	query := `
	INSERT INTO runs_history (name, quiz_id, started_at, finished_at) VALUES ($1, $2, $3, $4) RETURNING id;
	`

	var id int

	err := s.pool.QueryRow(ctx, query, run.Name, run.QuizID, run.StartedAt, run.FinishedAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetRuns возвращает названия запусков по ID. Возвращает storage.ErrQuizNotFound, если статистики по квизу нет в БД.
func (s *Storage) GetRuns(ctx context.Context, name string) ([]*models.HistoryModel, error) {
	query := `
	SELECT name, started_at, finished_at
	FROM runs_history
	WHERE name = $1;
	`

	var history []*models.HistoryModel

	rows, err := s.pool.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}

	history, err = pgx.CollectRows(rows, pgx.RowTo[*models.HistoryModel])
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetRunsStatistic возвращает список со статистикой запусков квиза по названию квиза
func (s *Storage) GetRunsStatistic(
	ctx context.Context,
	user *models.UserModel,
) ([]*models.StatisticModel, error) {
	query := `
	SELECT qi.name, rh.started_at, rh.finished_at, qs.score
	FROM quizzes_statistic qs
	JOIN quizzes_info qi ON qs.quiz_id = qi.id
	JOIN runs_history rh ON rh.id = qs.run_id
	WHERE qs.student_id = $1
	ORDER BY rh.started_at DESC
	`

	rows, err := s.pool.Query(ctx, query, user.ID)
	if err != nil {
		return nil, err
	}

	statistics, err := pgx.CollectRows(rows, pgx.RowTo[*models.StatisticModel])
	if err != nil {
		return nil, err
	}

	return statistics, nil
}

// SaveRunStatistic сохраняет статистику по запуску квиза в БД. Если статистика по студенту и запуску уже есть, то обновляет ее.
func (s *Storage) SaveRunStatistic(
	ctx context.Context,
	statistic *models.StatisticModel,
	scores map[int]int,
) error {
	if len(scores) == 0 {
		return nil
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	query := `
		INSERT INTO quizzes_statistic (quiz_id, student_id, run_id, score)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (student_id, run_id) DO UPDATE
		SET score = EXCLUDED.score
	`
	if err != nil {
		return err
	}

	slog.Debug("scores", "data", scores)

	for studentID, score := range scores {
		slog.Debug("SaveRunStatistic: ", "quiz_id", statistic.QuizID, "student_id", studentID, "run_id", statistic.RunID, "score", score)
		if _, err := tx.Exec(ctx, query, statistic.QuizID, studentID, statistic.RunID, score); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
