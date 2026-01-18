package engine

import (
	"context"
	"time"
)

// Quiz представляет загруженный квиз.
type Quiz struct {
	ID        string
	OwnerID   int64
	Title     string
	Settings  Settings
	Questions []Question
	CreatedAt time.Time
}

// Settings содержит настройки квиза.
type Settings struct {
	TimePerQuestion  int      `json:"time_per_question"`
	ShuffleQuestions bool     `json:"shuffle_questions"`
	ShuffleAnswers   bool     `json:"shuffle_answers"`
	MaxParticipants  int      `json:"max_participants"`
	Registration     []string `json:"registration"`
}

// Question представляет вопрос квиза.
type Question struct {
	Text        string   `json:"text"`
	Options     []string `json:"options"`
	Correct     int      `json:"correct"`
	Explanation string   `json:"explanation"`
	Points      int      `json:"points"`
	Time        int      `json:"time"`
	Shuffle     *bool    `json:"shuffle"`
}

// QuizRun представляет запуск квиза.
type QuizRun struct { //nolint:revive
	ID           string
	QuizID       string
	Status       RunStatus
	Participants map[int64]*Participant
	Answers      map[int64][]Answer
	StartedAt    time.Time
	FinishedAt   time.Time
}

// RunStatus — статус запуска квиза.
type RunStatus string

const (
	RunStatusLobby    RunStatus = "lobby"
	RunStatusRunning  RunStatus = "running"
	RunStatusFinished RunStatus = "finished"
)

// Participant представляет участника квиза.
type Participant struct {
	TelegramID int64
	Username   string
	FirstName  string
	LastName   string
	RegData    map[string]string
	JoinedAt   time.Time
}

// Answer представляет ответ участника на вопрос.
type Answer struct {
	QuestionIdx int
	AnswerIdx   int
	IsCorrect   bool
	Points      int
	AnsweredAt  time.Time
}

// QuizResults содержит результаты квиза.
type QuizResults struct { //nolint:revive
	RunID       string
	QuizTitle   string
	Leaderboard []LeaderboardEntry
	TotalTime   time.Duration
}

// LeaderboardEntry — запись в таблице лидеров.
type LeaderboardEntry struct {
	Participant  *Participant
	Score        int
	CorrectCount int
	TotalTime    time.Duration
	Rank         int
}

// QuizEngine определяет основной интерфейс для работы с квизами.
type QuizEngine interface { //nolint:revive
	// LoadQuiz парсит JSON и создаёт квиз.
	LoadQuiz(data []byte) (*Quiz, error)

	// StartRun создаёт новый запуск квиза в статусе "лобби".
	StartRun(ctx context.Context, quiz *Quiz) (*QuizRun, error)

	// JoinRun добавляет участника в запуск квиза.
	JoinRun(ctx context.Context, runID string, participant *Participant) error

	// GetParticipantCount возвращает текущее количество участников.
	GetParticipantCount(runID string) int

	// StartQuiz запускает квиз (переводит из лобби в running).
	// Возвращает канал для уведомлений о событиях квиза.
	StartQuiz(ctx context.Context, runID string) (<-chan QuizEvent, error)

	// ShuffleAnswers перемешивает порядок ответов на вопрос.
	ShuffleAnswers(ctx context.Context, runID string, event QuizEvent) error

	// SubmitAnswer регистрирует ответ участника по индексу (0-based).
	SubmitAnswer(
		ctx context.Context,
		runID string,
		participantID int64,
		questionIdx int,
		answerIdx int,
	) error

	// SubmitAnswerByLetter регистрирует ответ участника по букве (A, B, C, D, E, F).
	// Это основной способ ответа — участник пишет букву в чат.
	SubmitAnswerByLetter(
		ctx context.Context,
		runID string,
		participantID int64,
		letter string,
	) error

	// GetCurrentQuestion возвращает текущий номер вопроса для участника.
	// Возвращает -1 если квиз не запущен или завершён.
	GetCurrentQuestion(runID string) int

	// GetResults возвращает результаты завершённого квиза.
	GetResults(runID string) (*QuizResults, error)

	// ExportCSV экспортирует результаты в формате CSV.
	ExportCSV(runID string) ([]byte, error)

	// GetRun возвращает запуск по ID.
	GetRun(runID string) (*QuizRun, error)
}

// QuizEvent представляет событие квиза.
type QuizEvent struct { //nolint:revive
	Type        EventType
	QuestionIdx int
	Question    *Question
	TimeLeft    time.Duration
}

// EventType — тип события квиза.
type EventType string

const (
	EventTypeQuestion EventType = "question"
	EventTypeTimeUp   EventType = "time_up"
	EventTypeFinished EventType = "finished"
)

// MaxCountOfEvents - лимит событий в квизе.
const MaxCountOfEvents int64 = 1000

// AnswerLetters — допустимые буквы для ответов (A-F для до 6 вариантов).
var AnswerLetters = []string{"A", "B", "C", "D", "E", "F"}

// LetterToIndex преобразует букву в индекс (A=0, B=1, ...).
func LetterToIndex(letter string) (int, bool) {
	for i, l := range AnswerLetters {
		if l == letter {
			return i, true
		}
	}

	return -1, false
}

// IndexToLetter преобразует индекс в букву (0=A, 1=B, ...).
func IndexToLetter(idx int) string {
	if idx >= 0 && idx < len(AnswerLetters) {
		return AnswerLetters[idx]
	}

	return ""
}
