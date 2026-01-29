package engine

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Engine реализует QuizEngine.
type Engine struct {
	quizzes               map[string]*Quiz    // ключ - quizID
	activeQuizzesRun      map[string]*QuizRun // ключ - runID
	runIDToEvents         map[string]chan QuizEvent
	runIDToQuestionNumber map[string]int
	startTimeOfQuestion   map[int]time.Time
	quizErrChan           map[string]chan struct{} // для выхода из горутины при ошибке
	mu                    sync.RWMutex
}

// NewEngine создаёт новый Engine.
func NewEngine() *Engine {
	return &Engine{
		quizzes:               make(map[string]*Quiz),
		activeQuizzesRun:      make(map[string]*QuizRun),
		runIDToEvents:         make(map[string]chan QuizEvent),
		runIDToQuestionNumber: make(map[string]int),
		startTimeOfQuestion:   make(map[int]time.Time),
		quizErrChan:           make(map[string]chan struct{}),
	}
}

// LoadQuiz парсит JSON и создаёт квиз.
// Возвращает указатель на загруженный квиз.
func (e *Engine) LoadQuiz(data []byte) (*Quiz, error) {
	quiz := &Quiz{}
	if err := json.Unmarshal(data, quiz); err != nil {
		return nil, err
	}

	if err := isCorrectQuiz(quiz); err != nil {
		return nil, fmt.Errorf("cannot load events, %w", err)
	}

	quizID := uuid.NewString()

	e.mu.Lock()
	e.quizzes[quizID] = quiz
	e.mu.Unlock()

	quiz.ID = quizID

	return quiz, nil
}

// StartRun создаёт новый запуск квиза.
// Возвращает указатель на запуск квиза.
func (e *Engine) StartRun(ctx context.Context, quiz *Quiz) (*QuizRun, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if quiz == nil {
		return nil, ErrNilQuiz
	}

	runID := uuid.NewString()
	activeQuizRun := &QuizRun{
		ID:           runID,
		QuizID:       quiz.ID,
		Status:       RunStatusLobby,
		Participants: make(map[int64]*Participant),
		Answers:      make(map[int64][]Answer),
		StartedAt:    time.Now(),
	}

	e.mu.Lock()
	e.activeQuizzesRun[runID] = activeQuizRun
	e.mu.Unlock()

	return activeQuizRun, nil
}

// JoinRun добавляет участника в запуск квиза.
func (e *Engine) JoinRun(ctx context.Context, runID string, participant *Participant) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if participant == nil {
		return ErrNilParticipant
	}

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		return ErrNoRunLobby
	}

	quiz := e.quizzes[activeQuizRun.QuizID]
	if quiz.Settings.MaxParticipants != 0 &&
		len(activeQuizRun.Participants) >= quiz.Settings.MaxParticipants {
		return ErrLobbyFull
	}

	if _, ok = activeQuizRun.Participants[participant.TelegramID]; ok {
		return ErrRepeatedJoin
	}

	activeQuizRun.Participants[participant.TelegramID] = participant
	activeQuizRun.Answers[participant.TelegramID] = make([]Answer, 0, len(quiz.Questions))
	participant.JoinedAt = time.Now()

	return nil
}

// GetParticipantCount возвращает текущее количество участников.
func (e *Engine) GetParticipantCount(runID string) int {
	e.mu.RLock()
	defer e.mu.RUnlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		return -1
	}

	return len(activeQuizRun.Participants)
}

// StartQuiz запускает квиз.
// Возвращает канал событий квиза.
func (e *Engine) StartQuiz(ctx context.Context, runID string) (<-chan QuizEvent, error) {
	e.mu.Lock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		e.mu.Unlock()
		return nil, fmt.Errorf("quiz with runID: %s is not found", runID)
	}

	if activeQuizRun.Status != RunStatusLobby {
		e.mu.Unlock()
		return nil, ErrNoRunningStatus
	}

	activeQuizRun.Status = RunStatusRunning

	quiz := e.quizzes[activeQuizRun.QuizID]

	e.runIDToEvents[runID] = make(chan QuizEvent, MaxCountOfEvents)
	quizEvents := e.runIDToEvents[runID]
	e.quizErrChan[runID] = make(chan struct{}, 1)

	e.mu.Unlock()

	go func() {
		defer close(quizEvents)

		for i, question := range quiz.Questions {
			select {
			case <-ctx.Done():
				return
			default:
				e.mu.Lock()

				e.runIDToQuestionNumber[runID] = i

				e.startTimeOfQuestion[i] = time.Now()

				e.mu.Unlock()

				timePerQuestion := quiz.Settings.TimePerQuestion
				if question.Time != 0 {
					timePerQuestion = question.Time
				}

				questionEvent := QuizEvent{
					Type:        EventTypeQuestion,
					QuestionIdx: i,
					Question:    &question,
					TimeLeft:    time.Duration(timePerQuestion) * time.Second,
				}
				quizEvents <- questionEvent

				ok = e.waitEndOfQuestion(ctx, activeQuizRun, i, timePerQuestion, quiz, quizEvents, e.quizErrChan[runID])
				if !ok {
					return
				}
			}
		}

		e.mu.Lock()

		activeQuizRun.Status = RunStatusFinished
		activeQuizRun.FinishedAt = time.Now()

		e.mu.Unlock()

		quizEvents <- QuizEvent{
			Type: EventTypeFinished,
		}
	}()

	return quizEvents, nil
}

// ShuffleAnswers перемешивает порядок ответов на вопрос.
func (e *Engine) ShuffleAnswers(runID string, event QuizEvent) error {
	if event.Type != EventTypeQuestion {
		return ErrNoQuestionType
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		return ErrNoRunLobby
	}

	quiz := e.quizzes[activeQuizRun.QuizID]

	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	for k := range quiz.Questions {
		options := quiz.Questions[k].Options

		correctOption := options[quiz.Questions[k].Correct]
		if quiz.Questions[k].Shuffle != nil && *quiz.Questions[k].Shuffle {
			randGen.Shuffle(len(options), func(i, j int) {
				quiz.Questions[k].Options[i], quiz.Questions[k].Options[j] = quiz.Questions[k].Options[j], quiz.Questions[k].Options[i]
			})

			for j, option := range options {
				if option == correctOption {
					quiz.Questions[k].Correct = j
					break
				}
			}
		}
	}

	return nil
}

// SubmitAnswer регистрирует ответ участника.
func (e *Engine) SubmitAnswer(
	runID string,
	participantID int64,
	questionIdx int,
	answerIdx int,
) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	quizErrChan := e.quizErrChan[runID]

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusRunning {
		close(quizErrChan)
		return fmt.Errorf("quiz with runID: %s not running", runID)
	}

	quiz := e.quizzes[activeQuizRun.QuizID]

	questionsLength := len(quiz.Questions)
	if questionIdx < 0 || questionIdx >= questionsLength {
		close(quizErrChan)
		return ErrInvalidQuestionIndex
	}

	optionsLength := len(quiz.Questions[questionIdx].Options)
	if answerIdx < 0 || answerIdx >= optionsLength {
		close(quizErrChan)
		return ErrInvalidAnswerIndex
	}

	if _, ok = activeQuizRun.Participants[participantID]; !ok {
		close(quizErrChan)
		return fmt.Errorf("no such participant with id %d", participantID)
	}

	isCorrect := false

	question := quiz.Questions[questionIdx]
	if question.Correct == answerIdx {
		isCorrect = true
	}

	answer := Answer{
		QuestionIdx: questionIdx,
		AnswerIdx:   answerIdx,
		IsCorrect:   isCorrect,
		Points:      0,
		AnsweredAt:  time.Now(),
	}
	if isCorrect {
		answer.Points += question.Points
	}

	activeQuizRun.Answers[participantID] = append(activeQuizRun.Answers[participantID], answer)

	return nil
}

// SubmitAnswerByLetter регистрирует ответ участника по букве.
func (e *Engine) SubmitAnswerByLetter(
	runID string,
	participantID int64,
	letter string,
) error {
	e.mu.Lock()
	quizErrChan := e.quizErrChan[runID]
	e.mu.Unlock()

	answerIndex, ok := LetterToIndex(letter)
	if !ok {
		close(quizErrChan)
		return ErrConvertLetterToIndex
	}

	return e.SubmitAnswer(runID, participantID, e.GetCurrentQuestion(runID), answerIndex)
}

// GetCurrentQuestion возвращает текущий номер вопроса.
func (e *Engine) GetCurrentQuestion(runID string) int {
	e.mu.Lock()
	defer e.mu.Unlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusRunning {
		return -1
	}

	return e.runIDToQuestionNumber[runID]
}

// GetResults возвращает результаты квиза.
func (e *Engine) GetResults(runID string) (*QuizResults, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusFinished {
		return nil, fmt.Errorf("events %s is not finished", runID)
	}

	results := &QuizResults{
		RunID:       runID,
		QuizTitle:   e.quizzes[activeQuizRun.QuizID].Title,
		Leaderboard: make([]LeaderboardEntry, 0, len(activeQuizRun.Participants)),
		TotalTime:   activeQuizRun.FinishedAt.Sub(activeQuizRun.StartedAt),
	}

	for participantTelegramID, participant := range activeQuizRun.Participants {
		participantScore := 0
		correctCount := 0

		var timeResult time.Duration

		answers := activeQuizRun.Answers[participantTelegramID]
		for _, answer := range answers {
			if answer.IsCorrect {
				if answer.Points == 0 {
					participantScore += 1
				} else {
					participantScore += answer.Points
				}

				correctCount++
			}

			timeResult += answer.AnsweredAt.Sub(e.startTimeOfQuestion[answer.QuestionIdx])
		}

		results.Leaderboard = append(results.Leaderboard, LeaderboardEntry{
			Participant:  participant,
			Score:        participantScore,
			CorrectCount: correctCount,
			TotalTime:    timeResult,
			Rank:         0,
		})
	}

	sort.Slice(results.Leaderboard, func(i, j int) bool {
		if results.Leaderboard[i].Score != results.Leaderboard[j].Score {
			return results.Leaderboard[i].Score > results.Leaderboard[j].Score
		}

		return results.Leaderboard[i].TotalTime < results.Leaderboard[j].TotalTime
	})

	for i := range results.Leaderboard {
		results.Leaderboard[i].Rank = i + 1
	}

	return results, nil
}

// ExportCSV экспортирует результаты в CSV.
func (e *Engine) ExportCSV(runID string) ([]byte, error) {
	quizResults, err := e.GetResults(runID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	_ = w.Write(
		[]string{
			"Rank",
			"TelegramID",
			"Username",
			"FirstName",
			"LastName",
			"Score",
			"CorrectCount",
			"TotalTime",
		},
	)

	for _, ld := range quizResults.Leaderboard {
		_ = w.Write([]string{
			strconv.Itoa(ld.Rank),
			strconv.FormatInt(ld.Participant.TelegramID, 10),
			ld.Participant.Username,
			ld.Participant.FirstName,
			ld.Participant.LastName,
			strconv.Itoa(ld.Score),
			strconv.Itoa(ld.CorrectCount),
			ld.TotalTime.String(),
		})
	}

	w.Flush()

	err = w.Error()
	if err != nil {
		return nil, fmt.Errorf("failed to flush buffer: %w", err)
	}

	return buf.Bytes(), nil
}

// GetRun возвращает запуск по ID.
func (e *Engine) GetRun(runID string) (*QuizRun, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		return nil, fmt.Errorf("quiz with runID: %s is not found", runID)
	}

	return activeQuizRun, nil
}

// waitEndOfQuestion ждет окончание вопроса.
func (e *Engine) waitEndOfQuestion(
	ctx context.Context,
	activeQuizRun *QuizRun,
	questionIndex, questionTime int,
	quiz *Quiz,
	events chan QuizEvent,
	quizErrChan chan struct{},
) bool {
	timer := time.NewTimer(time.Duration(questionTime) * time.Second)
	timeToCheckForAllAnswers := time.NewTicker(time.Second / 10)

	for {
		select {
		case <-timer.C:
			event := QuizEvent{
				Type:        EventTypeTimeUp,
				QuestionIdx: questionIndex,
				Question:    &quiz.Questions[questionIndex],
			}
			events <- event

			return true
		case <-timeToCheckForAllAnswers.C:
			e.mu.Lock()

			answeredCnt := 0

			for _, answers := range activeQuizRun.Answers {
				for _, answer := range answers {
					if answer.QuestionIdx == questionIndex {
						answeredCnt++
						break
					}
				}
			}

			if answeredCnt == len(activeQuizRun.Participants) {
				e.mu.Unlock()
				return true
			}

			e.mu.Unlock()
		case <-quizErrChan:
			return false
		case <-ctx.Done():
			return false
		}
	}
}
