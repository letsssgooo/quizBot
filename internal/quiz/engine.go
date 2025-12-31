//go:build !solution

package quiz

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Engine реализует QuizEngine.
type Engine struct {
	quizzes                       map[string]*Quiz
	activeQuizzesRun              map[string]*QuizRun
	runIDToQuestionNumber         map[string]int
	participantIDToQuestionNumber map[int64]map[int]struct{}
	startTimeOfQuestion           map[int]time.Time
	mu                            sync.RWMutex
}

// NewEngine создаёт новый QuizEngine.
func NewEngine() *Engine {
	return &Engine{
		quizzes:                       make(map[string]*Quiz),
		activeQuizzesRun:              make(map[string]*QuizRun),
		runIDToQuestionNumber:         make(map[string]int),
		participantIDToQuestionNumber: make(map[int64]map[int]struct{}),
		startTimeOfQuestion:           make(map[int]time.Time),
	}
}

// LoadQuiz парсит JSON и создаёт квиз.
func (e *Engine) LoadQuiz(data []byte) (*Quiz, error) {
	quiz := &Quiz{}
	if err := json.Unmarshal(data, quiz); err != nil {
		return nil, err
	}

	if err := isCorrectQuiz(quiz); err != nil {
		return nil, fmt.Errorf("can not load quiz, %w", err)
	}

	id := uuid.NewString()
	e.quizzes[id] = quiz
	quiz.ID = id

	return quiz, nil
}

func isCorrectQuiz(quiz *Quiz) error {
	if quiz.Title == "" {
		return fmt.Errorf("missing field title")
	}

	if quiz.Settings.TimePerQuestion == 0 {
		return fmt.Errorf("missing field time_per_question")
	}

	if quiz.Questions == nil {
		return fmt.Errorf("missing field questions")
	}

	if len(quiz.Questions) == 0 {
		return fmt.Errorf("need at least one question")
	}

	for i, question := range quiz.Questions {
		if question.Text == "" {
			return fmt.Errorf("missing field text of %d question", i)
		}

		if question.Options == nil {
			return fmt.Errorf("missing field options of %d question", i)
		}

		if len(question.Options) < 2 {
			return fmt.Errorf("amount of options must be at least two in %d question", i)
		}

		if question.Correct < 0 {
			return fmt.Errorf("index of correct answer must not be negative in %d question", i)
		}

		if question.Correct >= len(question.Options) {
			return fmt.Errorf("index of correct answer in %d question is out of range", i)
		}
	}

	return nil
}

// StartRun создаёт новый запуск квиза.
func (e *Engine) StartRun(ctx context.Context, quiz *Quiz) (*QuizRun, error) {
	activeQuiz := &QuizRun{
		ID:           uuid.NewString(),
		QuizID:       quiz.ID,
		Status:       RunStatusLobby,
		Participants: make(map[int64]*Participant),
		Answers:      make(map[int64][]Answer),
		StartedAt:    time.Now(),
	}
	slog.Debug("start id: ", activeQuiz.ID)
	e.activeQuizzesRun[activeQuiz.ID] = activeQuiz

	return activeQuiz, nil
}

// JoinRun добавляет участника в запуск квиза.
func (e *Engine) JoinRun(ctx context.Context, runID string, participant *Participant) error {
	activeQuiz, ok := e.activeQuizzesRun[runID]
	slog.Debug("join run id: ", runID)
	if !ok {
		return errors.New("lobby of current quiz does not launched")
	}

	if _, ok = activeQuiz.Participants[participant.TelegramID]; ok {
		return errors.New("participant already joined")
	}

	activeQuiz.Participants[participant.TelegramID] = participant

	return nil
}

// GetParticipantCount возвращает текущее количество участников.
func (e *Engine) GetParticipantCount(runID string) int {
	activeQuiz := e.activeQuizzesRun[runID]
	return len(activeQuiz.Participants)
}

// StartQuiz запускает квиз.
func (e *Engine) StartQuiz(ctx context.Context, runID string) (<-chan QuizEvent, error) {
	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok {
		return nil, fmt.Errorf("quiz %s is not found", runID)
	}

	if activeQuizRun.Status != RunStatusLobby {
		return nil, errors.New(`can not start quiz, it is not in status "lobby"`)
	}
	activeQuizRun.Status = RunStatusRunning

	quiz := e.quizzes[activeQuizRun.QuizID]

	quizEvents := make(chan QuizEvent)

	go func() {
		defer close(quizEvents)

		for i, question := range quiz.Questions {
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

			timer := time.NewTimer(time.Duration(timePerQuestion) * time.Second)

			select {
			case <-timer.C:
				questionIsAnswered := false

				e.mu.RLock()

				for _, participant := range activeQuizRun.Participants {
					if _, ok = e.participantIDToQuestionNumber[participant.TelegramID][i]; ok {
						questionIsAnswered = true
						break
					}
				}

				e.mu.RUnlock()

				if !questionIsAnswered {
					timeEvent := QuizEvent{
						Type:        EventTypeTimeUp,
						QuestionIdx: i,
						Question:    &question,
					}
					quizEvents <- timeEvent
				}
			case <-ctx.Done():
				break
			}
		}

		e.mu.Lock()

		activeQuizRun.Status = RunStatusFinished

		e.mu.Unlock()

		quizResults, err := e.GetResults(runID)
		if err != nil {
			slog.Debug("error while getting quiz results:", "err", err)
			return
		}

		quizEvents <- QuizEvent{
			Type:    EventTypeFinished,
			Results: quizResults,
		}
	}()

	return quizEvents, nil
}

// SubmitAnswer регистрирует ответ участника.
func (e *Engine) SubmitAnswer(
	ctx context.Context,
	runID string,
	participantID int64,
	questionIdx int,
	answerIdx int,
) error {
	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusRunning {
		return fmt.Errorf("quiz %s not running", runID)
	}

	questionsLength := len(e.quizzes[activeQuizRun.QuizID].Questions)
	if questionIdx < 0 || questionIdx >= questionsLength {
		return errors.New("invalid index of answer")
	}

	optionsLength := len(e.quizzes[activeQuizRun.QuizID].Questions[questionIdx].Options)
	if answerIdx < 0 || answerIdx >= optionsLength {
		return errors.New("invalid index of answer")
	}

	if _, ok = e.participantIDToQuestionNumber[participantID][questionIdx]; ok {
		return nil
	}

	isCorrect := false

	question := e.quizzes[activeQuizRun.QuizID].Questions[questionIdx]
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

	if e.participantIDToQuestionNumber[participantID] == nil {
		e.participantIDToQuestionNumber[participantID] = make(map[int]struct{})
	}

	e.participantIDToQuestionNumber[participantID][questionIdx] = struct{}{}
	activeQuizRun.Answers[participantID] = append(activeQuizRun.Answers[participantID], answer)

	return nil
}

// SubmitAnswerByLetter регистрирует ответ участника по букве.
func (e *Engine) SubmitAnswerByLetter(
	ctx context.Context,
	runID string,
	participantID int64,
	letter string,
) error {
	answerIndex, ok := LetterToIndex(letter)
	if !ok {
		return errors.New("can not convert letter to index, invalid input")
	}

	return e.SubmitAnswer(ctx, runID, participantID, e.GetCurrentQuestion(runID), answerIndex)
}

// GetCurrentQuestion возвращает текущий номер вопроса.
func (e *Engine) GetCurrentQuestion(runID string) int {
	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusRunning {
		return -1
	}

	return e.runIDToQuestionNumber[runID]
}

// GetResults возвращает результаты квиза.
func (e *Engine) GetResults(runID string) (*QuizResults, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	activeQuizRun, ok := e.activeQuizzesRun[runID]
	if !ok || activeQuizRun.Status != RunStatusFinished {
		return nil, fmt.Errorf("quiz %s is not finished", runID)
	}

	results := &QuizResults{
		RunID:       runID,
		QuizTitle:   e.quizzes[activeQuizRun.QuizID].Title,
		Leaderboard: make([]LeaderboardEntry, 0, len(activeQuizRun.Participants)),
		TotalTime:   time.Since(activeQuizRun.StartedAt),
	}

	for participantTelegramID, participant := range activeQuizRun.Participants {
		participantResult := 0
		correctCount := 0

		var timeResult time.Duration

		answers := activeQuizRun.Answers[participantTelegramID]
		for _, answer := range answers {
			if answer.IsCorrect {
				if answer.Points == 0 {
					participantResult += 1
				} else {
					participantResult += answer.Points
				}

				correctCount++
			}

			timeResult += answer.AnsweredAt.Sub(e.startTimeOfQuestion[answer.QuestionIdx])
		}

		results.Leaderboard = append(results.Leaderboard, LeaderboardEntry{
			Participant:  participant,
			Score:        participantResult,
			CorrectCount: correctCount,
			TotalTime:    timeResult,
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

	leaderboard := make([][]string, len(quizResults.Leaderboard)+1)
	leaderboard[0] = []string{
		"Rank",
		"TelegramID",
		"Username",
		"FirstName",
		"LastName",
		"Score",
		"CorrectCount",
		"TotalTime",
	}
	for i, line := range quizResults.Leaderboard {
		rank := fmt.Sprintf("%d", line.Rank)
		telegramID := fmt.Sprintf("%d", line.Participant.TelegramID)
		username := line.Participant.Username
		firstName := line.Participant.FirstName
		lastName := line.Participant.LastName
		score := fmt.Sprintf("%d", line.Score)
		correctCount := fmt.Sprintf("%d", line.CorrectCount)
		totalTime := fmt.Sprintf("%v", line.TotalTime.Seconds())
		leaderboard[i+1] = []string{
			rank,
			telegramID,
			username,
			firstName,
			lastName,
			score,
			correctCount,
			totalTime,
		}
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	err = w.WriteAll(leaderboard)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetRun возвращает запуск по ID.
func (e *Engine) GetRun(runID string) (*QuizRun, error) {
	_, ok := e.activeQuizzesRun[runID]
	if !ok {
		return nil, fmt.Errorf("quiz %s is not found", runID)
	}

	return e.activeQuizzesRun[runID], nil
}
