//go:build !change

package quiz

import (
	"context"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// drainEvents дочитывает все события из канала до его закрытия.
// Это необходимо в тестах, чтобы предотвратить утечки горутин.
// В реальном приложении Telegram бот всегда читает события до конца,
// но в тестах мы можем остановиться раньше, что приводит к зависанию
// горутины runQuizLoop на отправке события в канал.
func drainEvents(events <-chan QuizEvent) {
	// Дочитываем все оставшиеся события с таймаутом
	// Увеличиваем таймаут для тестов с long time_per_question
	timeout := time.After(10 * time.Second)

	for {
		select {
		case _, ok := <-events:
			if !ok {
				// Канал закрыт - это нормальное завершение
				return
			}
			// Продолжаем читать события
		case <-timeout:
			// Таймаут - выходим (это может означать проблему, но не будем паниковать)
			return
		}
	}
}

func TestLoadQuiz_Valid(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test Quiz",
		"settings": {
			"time_per_question": 20,
			"shuffle_questions": false,
			"shuffle_answers": false
		},
		"questions": [
			{
				"text": "What is 2+2?",
				"options": ["3", "4", "5", "6"],
				"correct": 1
			}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)
	require.NotNil(t, quiz)

	assert.Equal(t, "Test Quiz", quiz.Title)
	assert.Equal(t, 20, quiz.Settings.TimePerQuestion)
	assert.False(t, quiz.Settings.ShuffleQuestions)
	assert.False(t, quiz.Settings.ShuffleAnswers)
	assert.Len(t, quiz.Questions, 1)
	assert.Equal(t, "What is 2+2?", quiz.Questions[0].Text)
	assert.Equal(t, 1, quiz.Questions[0].Correct)
}

func TestLoadQuiz_InvalidJSON(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{invalid json}`)

	quiz, err := engine.LoadQuiz(data)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestLoadQuiz_MissingTitle(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"settings": {
			"time_per_question": 20
		},
		"questions": [
			{
				"text": "Question?",
				"options": ["A", "B"],
				"correct": 0
			}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestLoadQuiz_MissingQuestions(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test Quiz",
		"settings": {
			"time_per_question": 20
		},
		"questions": []
	}`)

	quiz, err := engine.LoadQuiz(data)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestLoadQuiz_InvalidCorrectIndex(t *testing.T) {
	engine := NewEngine()

	testCases := []struct {
		name string
		data string
	}{
		{
			name: "correct index negative",
			data: `{
				"title": "Test Quiz",
				"settings": {"time_per_question": 20},
				"questions": [{
					"text": "Question?",
					"options": ["A", "B", "C"],
					"correct": -1
				}]
			}`,
		},
		{
			name: "correct index out of range",
			data: `{
				"title": "Test Quiz",
				"settings": {"time_per_question": 20},
				"questions": [{
					"text": "Question?",
					"options": ["A", "B", "C"],
					"correct": 5
				}]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			quiz, err := engine.LoadQuiz([]byte(tc.data))
			assert.Error(t, err)
			assert.Nil(t, quiz)
		})
	}
}

func TestLoadQuiz_MissingTimePerQuestion(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test Quiz",
		"settings": {},
		"questions": [
			{
				"text": "Question?",
				"options": ["A", "B"],
				"correct": 0
			}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestLoadQuiz_TooFewOptions(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test Quiz",
		"settings": {"time_per_question": 20},
		"questions": [{
			"text": "Question?",
			"options": ["A"],
			"correct": 0
		}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	assert.Error(t, err)
	assert.Nil(t, quiz)
}

func TestLoadQuiz_WithExplanation(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test Quiz",
		"settings": {"time_per_question": 20},
		"questions": [{
			"text": "Question?",
			"options": ["A", "B"],
			"correct": 0,
			"explanation": "Because A is correct"
		}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)
	assert.Equal(t, "Because A is correct", quiz.Questions[0].Explanation)
}

func TestQuizFlow_SingleParticipant(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Single Participant Quiz",
		"settings": {"time_per_question": 5},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0},
			{"text": "Q2", "options": ["A", "B"], "correct": 1}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)
	require.NotEmpty(t, run.ID)
	assert.Equal(t, RunStatusLobby, run.Status)

	participant := &Participant{
		TelegramID: 12345,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
	}

	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	assert.Equal(t, 1, engine.GetParticipantCount(run.ID))

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	// Получаем первый вопрос
	event := <-events
	assert.Equal(t, EventTypeQuestion, event.Type)
	assert.Equal(t, 0, event.QuestionIdx)

	// Отвечаем на первый вопрос (правильно)
	err = engine.SubmitAnswer(ctx, run.ID, 12345, 0, 0)
	require.NoError(t, err)

	// Получаем второй вопрос
	event = <-events
	assert.Equal(t, EventTypeQuestion, event.Type)
	assert.Equal(t, 1, event.QuestionIdx)

	// Отвечаем на второй вопрос (неправильно)
	err = engine.SubmitAnswer(ctx, run.ID, 12345, 1, 0)
	require.NoError(t, err)

	// Квиз завершён
	event = <-events
	assert.Equal(t, EventTypeFinished, event.Type)

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	require.Len(t, results.Leaderboard, 1)
	assert.Equal(t, 1, results.Leaderboard[0].Score) // 1 правильный ответ
}

func TestQuizFlow_MultipleParticipants(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Multi Participant Quiz",
		"settings": {"time_per_question": 5},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participants := []*Participant{
		{TelegramID: 1, Username: "user1"},
		{TelegramID: 2, Username: "user2"},
		{TelegramID: 3, Username: "user3"},
	}

	for _, p := range participants {
		err = engine.JoinRun(ctx, run.ID, p)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, engine.GetParticipantCount(run.ID))

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events // первый вопрос

	// Все отвечают
	err = engine.SubmitAnswer(ctx, run.ID, 1, 0, 0) // правильно
	require.NoError(t, err)
	err = engine.SubmitAnswer(ctx, run.ID, 2, 0, 1) // неправильно
	require.NoError(t, err)
	err = engine.SubmitAnswer(ctx, run.ID, 3, 0, 0) // правильно
	require.NoError(t, err)

	<-events // finished

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	require.Len(t, results.Leaderboard, 3)

	// Первые два должны иметь по 1 баллу
	assert.Equal(t, 1, results.Leaderboard[0].Score)
	assert.Equal(t, 1, results.Leaderboard[1].Score)
	// Третий - 0 баллов
	assert.Equal(t, 0, results.Leaderboard[2].Score)
}

func TestQuestionTimer_Expires(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Timer Test",
		"settings": {"time_per_question": 1},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0},
			{"text": "Q2", "options": ["A", "B"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	// Первый вопрос
	event := <-events
	assert.Equal(t, EventTypeQuestion, event.Type)
	assert.Equal(t, 0, event.QuestionIdx)

	// Не отвечаем, ждём таймаут
	event = <-events
	assert.Equal(t, EventTypeTimeUp, event.Type)

	// Второй вопрос
	event = <-events
	assert.Equal(t, EventTypeQuestion, event.Type)
	assert.Equal(t, 1, event.QuestionIdx)

	// Отвечаем на второй
	err = engine.SubmitAnswer(ctx, run.ID, 1, 1, 0)
	require.NoError(t, err)

	// Квиз завершён
	event = <-events
	assert.Equal(t, EventTypeFinished, event.Type)

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, results.Leaderboard[0].Score) // только второй вопрос
}

func TestLeaderboard_Sorting(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Sorting Test",
		"settings": {"time_per_question": 10},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0},
			{"text": "Q2", "options": ["A", "B"], "correct": 0},
			{"text": "Q3", "options": ["A", "B"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participants := []*Participant{
		{TelegramID: 1, Username: "loser"},  // 0 правильных
		{TelegramID: 2, Username: "winner"}, // 3 правильных
		{TelegramID: 3, Username: "middle"}, // 1 правильный
	}

	for _, p := range participants {
		err = engine.JoinRun(ctx, run.ID, p)
		require.NoError(t, err)
	}

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	// Q1
	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 0, 1) // неправильно
	engine.SubmitAnswer(ctx, run.ID, 2, 0, 0) // правильно
	engine.SubmitAnswer(ctx, run.ID, 3, 0, 0) // правильно

	// Q2
	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 1, 1) // неправильно
	engine.SubmitAnswer(ctx, run.ID, 2, 1, 0) // правильно
	engine.SubmitAnswer(ctx, run.ID, 3, 1, 1) // неправильно

	// Q3
	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 2, 1) // неправильно
	engine.SubmitAnswer(ctx, run.ID, 2, 2, 0) // правильно
	engine.SubmitAnswer(ctx, run.ID, 3, 2, 1) // неправильно

	<-events // finished

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	require.Len(t, results.Leaderboard, 3)

	// Проверяем порядок: winner(3), middle(1), loser(0)
	assert.Equal(t, int64(2), results.Leaderboard[0].Participant.TelegramID)
	assert.Equal(t, 3, results.Leaderboard[0].Score)
	assert.Equal(t, 1, results.Leaderboard[0].Rank)

	assert.Equal(t, int64(3), results.Leaderboard[1].Participant.TelegramID)
	assert.Equal(t, 1, results.Leaderboard[1].Score)
	assert.Equal(t, 2, results.Leaderboard[1].Rank)

	assert.Equal(t, int64(1), results.Leaderboard[2].Participant.TelegramID)
	assert.Equal(t, 0, results.Leaderboard[2].Score)
	assert.Equal(t, 3, results.Leaderboard[2].Rank)
}

func TestLeaderboard_SameScore_SortByTime(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Same Score Test",
		"settings": {"time_per_question": 10},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participants := []*Participant{
		{TelegramID: 1, Username: "slow"},
		{TelegramID: 2, Username: "fast"},
	}

	for _, p := range participants {
		err = engine.JoinRun(ctx, run.ID, p)
		require.NoError(t, err)
	}

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events

	// fast отвечает первым
	engine.SubmitAnswer(ctx, run.ID, 2, 0, 0)
	time.Sleep(10 * time.Millisecond)
	// slow отвечает вторым
	engine.SubmitAnswer(ctx, run.ID, 1, 0, 0)

	<-events // finished

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	require.Len(t, results.Leaderboard, 2)

	// fast должен быть первым (меньше общее время)
	assert.Equal(t, int64(2), results.Leaderboard[0].Participant.TelegramID)
	assert.Equal(t, int64(1), results.Leaderboard[1].Participant.TelegramID)
}

func TestExportCSV_Format(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "CSV Test",
		"settings": {"time_per_question": 5},
		"questions": [
			{"text": "Q1", "options": ["A", "B"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{
		TelegramID: 1,
		Username:   "testuser",
		FirstName:  "Test",
		LastName:   "User",
	}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 0, 0)
	<-events

	csvData, err := engine.ExportCSV(run.ID)
	require.NoError(t, err)
	require.NotEmpty(t, csvData)

	reader := csv.NewReader(strings.NewReader(string(csvData)))
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Проверяем заголовок
	require.GreaterOrEqual(t, len(records), 2)
	header := records[0]
	assert.Contains(t, header, "Rank")
	assert.Contains(t, header, "Username")
	assert.Contains(t, header, "Score")

	// Проверяем данные
	assert.Equal(t, "1", records[1][0]) // Rank
}

func TestExportCSV_UTF8(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "UTF8 Test",
		"settings": {"time_per_question": 5},
		"questions": [
			{"text": "Вопрос?", "options": ["Да", "Нет"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{
		TelegramID: 1,
		Username:   "пользователь",
		FirstName:  "Иван",
		LastName:   "Иванов",
	}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 0, 0)
	<-events

	csvData, err := engine.ExportCSV(run.ID)
	require.NoError(t, err)

	// Проверяем UTF-8
	assert.Contains(t, string(csvData), "пользователь")
	assert.Contains(t, string(csvData), "Иван")
}

func TestJoinRun_AlreadyJoined(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test",
		"settings": {"time_per_question": 5},
		"questions": [{"text": "Q", "options": ["A", "B"], "correct": 0}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1}

	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	// Повторная попытка присоединиться
	err = engine.JoinRun(ctx, run.ID, participant)
	assert.Error(t, err)

	assert.Equal(t, 1, engine.GetParticipantCount(run.ID))
}

func TestSubmitAnswer_InvalidQuestionIdx(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test",
		"settings": {"time_per_question": 5},
		"questions": [{"text": "Q", "options": ["A", "B"], "correct": 0}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events

	// Неверный индекс вопроса
	err = engine.SubmitAnswer(ctx, run.ID, 1, 5, 0)
	assert.Error(t, err)
}

func TestSubmitAnswer_InvalidAnswerIdx(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test",
		"settings": {"time_per_question": 5},
		"questions": [{"text": "Q", "options": ["A", "B"], "correct": 0}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	<-events

	// Неверный индекс ответа
	err = engine.SubmitAnswer(ctx, run.ID, 1, 0, 10)
	assert.Error(t, err)
}

func TestGetRun_NotFound(t *testing.T) {
	engine := NewEngine()

	run, err := engine.GetRun("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, run)
}

func TestStartQuiz_NotInLobby(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Test",
		"settings": {"time_per_question": 5},
		"questions": [{"text": "Q", "options": ["A", "B"], "correct": 0}]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()

	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	// Первый запуск
	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	// Попытка запустить ещё раз (уже не в лобби)
	_, err = engine.StartQuiz(ctx, run.ID)
	assert.Error(t, err)

	// Завершаем квиз
	<-events
	engine.SubmitAnswer(ctx, run.ID, 1, 0, 0)
	<-events
}

func TestSubmitAnswerByLetter_Basic(t *testing.T) {
	engine := NewEngine()

	data := []byte(`{
		"title": "Letter Test",
		"settings": {"time_per_question": 10},
		"questions": [
			{
				"text": "Выберите правильный ответ",
				"options": ["Первый", "Второй", "Третий", "Четвёртый"],
				"correct": 2
			}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()
	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participant := &Participant{TelegramID: 1, Username: "test"}
	err = engine.JoinRun(ctx, run.ID, participant)
	require.NoError(t, err)

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	event := <-events
	assert.Equal(t, EventTypeQuestion, event.Type)

	// Отвечаем буквой C (индекс 2 = правильный ответ)
	err = engine.SubmitAnswerByLetter(ctx, run.ID, 1, "C")
	require.NoError(t, err)

	<-events // finished

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, results.Leaderboard[0].Score)
	assert.Equal(t, 1, results.Leaderboard[0].CorrectCount)
}

func TestQuizWithTextAnswers_FullFlow(t *testing.T) {
	// Полный flow квиза с текстовыми ответами A/B/C/D
	engine := NewEngine()

	data := []byte(`{
		"title": "Text Answers Quiz",
		"settings": {"time_per_question": 10},
		"questions": [
			{"text": "Q1: Какой ответ правильный?", "options": ["Неверно", "Верно"], "correct": 1},
			{"text": "Q2: A или B?", "options": ["A правильно", "B правильно", "C правильно"], "correct": 0}
		]
	}`)

	quiz, err := engine.LoadQuiz(data)
	require.NoError(t, err)

	ctx := context.Background()
	run, err := engine.StartRun(ctx, quiz)
	require.NoError(t, err)

	participants := []*Participant{
		{TelegramID: 1, Username: "alice"},
		{TelegramID: 2, Username: "bob"},
	}

	for _, p := range participants {
		err = engine.JoinRun(ctx, run.ID, p)
		require.NoError(t, err)
	}

	events, err := engine.StartQuiz(ctx, run.ID)
	require.NoError(t, err)

	// Cleanup: дочитываем все оставшиеся события
	defer drainEvents(events)

	// Q1
	<-events
	engine.SubmitAnswerByLetter(ctx, run.ID, 1, "B") // правильно
	engine.SubmitAnswerByLetter(ctx, run.ID, 2, "A") // неправильно

	// Q2
	<-events
	engine.SubmitAnswerByLetter(ctx, run.ID, 1, "A") // правильно
	engine.SubmitAnswerByLetter(ctx, run.ID, 2, "A") // правильно

	<-events // finished

	results, err := engine.GetResults(run.ID)
	require.NoError(t, err)

	// Alice: 2 правильных, Bob: 1 правильный
	assert.Equal(t, int64(1), results.Leaderboard[0].Participant.TelegramID)
	assert.Equal(t, 2, results.Leaderboard[0].Score)
	assert.Equal(t, int64(2), results.Leaderboard[1].Participant.TelegramID)
	assert.Equal(t, 1, results.Leaderboard[1].Score)
}
