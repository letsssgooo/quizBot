package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/letsssgooo/quizBot/internal/auth"
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
	"github.com/letsssgooo/quizBot/internal/events/fetcher"
	"github.com/letsssgooo/quizBot/internal/events/sender"
	"github.com/letsssgooo/quizBot/internal/storage"
)

const updatesTimeout = 30

// Bot реализует Telegram бота для квизов.
type Bot struct {
	client              client.Client
	auth                auth.Auth
	fetcher             fetcher.Fetcher
	sender              sender.Sender
	engine              engine.QuizEngine
	storage             storage.Storage
	botUsername         string // Username бота для формирования ссылок (например, "my_quiz_bot")
	userIDToRunID       map[int64]string
	IsLecturersID       map[int64]bool
	runIDToLobbyEndChan map[string]chan struct{}
	runIDToQuiz         map[string]*engine.Quiz
	userIDToChatID      map[int64]int64
	runIDToOwnerChatID  map[string]int64
	userIDToAnswersCnt  map[int64]int
	hasLecturer         bool
	mu                  sync.Mutex
}

// NewBot создаёт нового бота.
// botUsername — username бота без @ (например, "my_quiz_bot").
// Используется для формирования ссылок: https://t.me/<botUsername>?start=join_<runID>
func NewBot(
	client client.Client,
	auth auth.Auth,
	fetcher fetcher.Fetcher,
	sender sender.Sender,
	quizEngine engine.QuizEngine,
	storage storage.Storage,
	botUsername string,
) *Bot {
	return &Bot{
		client:              client,
		auth:                auth,
		fetcher:             fetcher,
		sender:              sender,
		engine:              quizEngine,
		storage:             storage,
		botUsername:         botUsername,
		userIDToRunID:       make(map[int64]string),
		IsLecturersID:       make(map[int64]bool),
		runIDToLobbyEndChan: make(map[string]chan struct{}),
		runIDToQuiz:         make(map[string]*engine.Quiz),
		userIDToChatID:      make(map[int64]int64),
		runIDToOwnerChatID:  make(map[string]int64),
		userIDToAnswersCnt:  make(map[int64]int),
	}
}

// Run запускает бота (long polling).
func (b *Bot) Run() error {
	slog.Debug("Bot started!")

	for { // long polling
		updates, err := b.fetcher.GetUpdates(updatesTimeout)
		if err != nil {
			return err
		}

		for _, update := range updates {
			err = b.HandleUpdate(update)
			if err != nil {
				return err
			}
		}
	}
}

// HandleUpdate обрабатывает одно обновление.
func (b *Bot) HandleUpdate(update client.Update) error {
	if update.Message != nil {
		return b.handleMessageUpdate(update.Message)
	} else if update.CallbackQuery != nil {
		return b.handleCallbackUpdate(update.CallbackQuery)
	}

	return fmt.Errorf("%w, update :%v", errors.New("undefined update type"), update)
}

// handleMessageUpdate обрабатывает одно сообщение.
func (b *Bot) handleMessageUpdate(message *client.Message) error {
	ID := message.From.ID

	b.mu.Lock()
	runID, ok := b.userIDToRunID[ID]
	b.mu.Unlock()

	if ok {
		return b.handleAnswerUpdate(message.Chat.ID, message.From.ID, message.Text, runID)
	}

	if b.IsLecturersID[ID] && message.Document != nil {
		return b.handleDocumentUpdate(message)
	} else if !b.IsLecturersID[ID] && message.Document != nil { // TODO: SELECT запрос в БД на проверку прав, вместо мапы
		_, err := b.sender.Message(message.Chat.ID, msgNoRights, nil)

		return err
	}

	text := strings.Fields(message.Text)

	if len(text[0]) < 1 {
		_, err := b.sender.Message(message.Chat.ID, msgUnknownText, nil)

		return err
	}

	if text[0][:1] == "/" {
		// TODO: добавить все команды бота
		switch text[0] {
		case "/start":
			return b.handleStartCommand(message, text)
		case "/help":
			return b.handleHelpCommand(message)
		default:
			_, err := b.sender.Message(message.Chat.ID, msgUnknownCommand, nil)

			return err
		}
	} else if len(text) == 4 {
		ctx := context.Background()

		err := b.auth.UpdateStudentData(ctx, b.storage, message.From.ID, text)
		if errors.Is(err, auth.ErrValidation) {
			slog.Debug("incorrect student data: ", "error", err)

			_, err = b.sender.Message(message.Chat.ID, msgStudentsDataMistake, nil)

			return err
		} else if err != nil {
			return err
		}

		_, err = b.sender.Message(message.Chat.ID, msgStudentsSuccessfullVerification, nil)

		return err
	}

	_, err := b.sender.Message(message.Chat.ID, msgUnknownText, nil)

	return err
}

// handleStartCommand обрабатывает /start и /start join_<runID> команды.
func (b *Bot) handleStartCommand(message *client.Message, text []string) error {
	ctx := context.Background()
	err := b.auth.CreateUser(ctx, b.storage, message.From.ID)
	if errors.Is(err, storage.ErrUserAlreadyExists) {
		slog.Debug("Error while creating new user", "error", err)
	} else if err != nil {
		return err
	}

	if !errors.Is(err, storage.ErrUserAlreadyExists) {
		slog.Debug("User in database now", "username", message.From.Username)
	}

	if len(text) == 1 {
		keyboard := client.InlineKeyboardMarkup{
			InlineKeyboard: [][]client.InlineKeyboardButton{
				{
					{Text: "Студент", CallbackData: "Student"},
				},
				{
					{Text: "Преподаватель", CallbackData: "Lecturer"},
				},
			},
		}

		opts := &client.SendOptions{ReplyMarkup: &keyboard}
		_, err := b.sender.Message(message.Chat.ID, msgIdentification, opts)

		return err
	}

	// студент присоединен к квизу по ссылке
	runID := strings.Split(text[1], "_")[1]

	return b.handleStudentsJoin(message, runID)
}

// handleStudentsJoin присоединяет студента к квизу.
func (b *Bot) handleStudentsJoin(message *client.Message, runID string) error {
	b.mu.Lock()

	run, err := b.engine.GetRun(runID)
	b.mu.Unlock()

	if err != nil {
		_, err := b.client.SendMessage(message.Chat.ID, msgUnknownQuiz, nil)

		return err
	}

	if run.Status != engine.RunStatusLobby {
		_, err := b.client.SendMessage(message.Chat.ID, msgClosedLobby, nil)

		return err
	}

	ctx := context.Background()
	participant := &engine.Participant{
		TelegramID: message.From.ID,
		Username:   message.From.Username,
		FirstName:  message.From.FirstName,
		LastName:   message.From.LastName,
	}

	err = b.engine.JoinRun(ctx, runID, participant)
	if errors.Is(err, engine.ErrLobbyFull) {
		_, err = b.client.SendMessage(message.Chat.ID, msgMaxParticipantNumber, nil)

		return err
	} else if err != nil {
		return err
	}

	b.mu.Lock()
	b.userIDToRunID[message.From.ID] = runID
	b.userIDToChatID[message.From.ID] = message.Chat.ID
	b.mu.Unlock()

	_, err = b.client.SendMessage(message.Chat.ID, msgQuizJoin, nil)

	return err
}

// handleHelpCommand обрабатывает /help команду.
func (b *Bot) handleHelpCommand(message *client.Message) error {
	isLecturersID, ok := b.IsLecturersID[message.From.ID]
	if ok {
		if isLecturersID {
			_, err := b.sender.Message(message.Chat.ID, msgLecturersHelp, nil)

			return err
		}

		_, err := b.sender.Message(message.Chat.ID, msgStudentsHelp, nil)

		return err
	}

	_, err := b.sender.Message(message.Chat.ID, msgHelp, nil)

	return err
}

// handleAnswerUpdate обрабатывает ответ студента на вопрос.
func (b *Bot) handleAnswerUpdate(chatID, fromID int64, text string, runID string) error {
	ctx := context.Background()

	b.mu.Lock()

	currentQuestion := b.engine.GetCurrentQuestion(runID)
	hasAnswered := currentQuestion < b.userIDToAnswersCnt[fromID]
	b.mu.Unlock()

	if hasAnswered {
		_, err := b.sender.Message(chatID, msgRepeatedAnswer, nil)

		return err
	}

	err := b.engine.SubmitAnswerByLetter(ctx, runID, fromID, text)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.userIDToAnswersCnt[fromID]++
	b.mu.Unlock()

	_, err = b.sender.Message(chatID, msgAnswerAcceptance, nil)
	if err != nil {
		return err
	}

	return nil
}

// handleDocumentUpdate обрабатывает присланный преподавателем JSON файл.
func (b *Bot) handleDocumentUpdate(message *client.Message) error {
	filePath, err := b.client.GetFile(message.Document.FileID)
	if err != nil {
		return err
	}

	data, err := b.client.DownloadFile(filePath)
	if err != nil {
		return nil
	}

	quiz, err := b.engine.LoadQuiz(data)
	if err != nil {
		return err
	}

	quiz.OwnerID = message.From.ID
	quiz.CreatedAt = time.Now()
	ctx := context.Background()

	activeQuizRun, err := b.engine.StartRun(ctx, quiz)
	if err != nil {
		return err
	}

	b.mu.Lock()
	b.runIDToQuiz[activeQuizRun.ID] = quiz
	b.runIDToOwnerChatID[activeQuizRun.ID] = message.Chat.ID
	b.mu.Unlock()

	callbackData := fmt.Sprintf("start_quiz %s", activeQuizRun.ID)
	keyboard := client.InlineKeyboardMarkup{
		InlineKeyboard: [][]client.InlineKeyboardButton{
			{
				{Text: "Начать квиз", CallbackData: callbackData},
			},
		},
	}

	text := fmt.Sprintf(`Квиз создан.
Ссылка для студентов: %s
Количество участников: %d`, fmt.Sprintf("https://t.me/%s?start=join_%s", b.botUsername, activeQuizRun.ID), 0)
	opts := &client.SendOptions{
		ReplyMarkup: &keyboard,
	}

	botMessage, err := b.client.SendMessage(message.Chat.ID, text, opts)
	if err != nil {
		return err
	}

	go func() {
		_ = b.handleEditLecturerMessage(activeQuizRun.ID, botMessage, opts)
	}()

	return nil
}

// handleEditLecturerMessage каждые 3 сек изменяет счетчик участников в сообщении бота.
func (b *Bot) handleEditLecturerMessage(runID string, botMessage *client.Message, opts *client.SendOptions) error {
	lobbyEndChan := make(chan struct{})

	b.mu.Lock()
	b.runIDToLobbyEndChan[runID] = lobbyEndChan
	b.mu.Unlock()

	prevCnt := 0
	ticker := time.NewTicker(3 * time.Second)

	for {
		select {
		case <-ticker.C:
			cnt := b.engine.GetParticipantCount(runID)
			if cnt == prevCnt {
				continue
			}

			prevCnt = cnt

			text := fmt.Sprintf(`Квиз создан.
Ссылка для студентов: %s
Количество участников: %d`, fmt.Sprintf("https://t.me/%s?start=join_%s", b.botUsername, runID), cnt)
			_ = b.client.EditMessage(botMessage.Chat.ID, botMessage.MessageID, text, opts)
		case <-lobbyEndChan:
			return nil
		}
	}
}

// handleCallbackUpdate обрабатывает callback запрос.
func (b *Bot) handleCallbackUpdate(callback *client.CallbackQuery) error {
	if callback.Data == "Student" || callback.Data == "Lecturer" {
		return b.handleIdentificationCallbackUpdate(callback)
	}

	return b.handleQuizStartCallbackUpdate(callback)
}

// handleIdentificationCallbackUpdate определяет роль пользователя.
func (b *Bot) handleIdentificationCallbackUpdate(callback *client.CallbackQuery) error {
	switch callback.Data {
	case "Student":
		b.mu.Lock()

		b.IsLecturersID[callback.From.ID] = false // пишу это явно, чтобы понимать,
		// что этот пользователь именно студент,
		// а не тот, кто еще не выбрал роль

		b.mu.Unlock()

		_, err := b.sender.Message(callback.Message.Chat.ID, msgStudentsData, nil)

		return err
	case "Lecturer":
		b.mu.Lock()

		// TODO: запись данных преподавателя в БД (если надо, но скорее всего не понадобится)
		b.IsLecturersID[callback.From.ID] = true
		b.hasLecturer = true

		b.mu.Unlock()

		_, err := b.sender.Message(callback.Message.Chat.ID, msgLecturersSuccessfullVerification, nil)

		return err
	default:
		return nil
	}
}

// handleQuizStartCallbackUpdate запускает квиз.
func (b *Bot) handleQuizStartCallbackUpdate(callback *client.CallbackQuery) error {
	err := b.client.AnswerCallback(callback.ID, msgQuizRunning)
	if err != nil {
		return err
	}

	runID := strings.Split(callback.Data, " ")[1]
	ctx := context.Background()

	events, err := b.engine.StartQuiz(ctx, runID)
	if err != nil {
		return err
	}

	b.mu.Lock()
	close(b.runIDToLobbyEndChan[runID]) // квиз запустился => больше нет лобби => больше не запускаем студентов
	participantsCnt := b.engine.GetParticipantCount(runID)
	b.mu.Unlock()

	msg := fmt.Sprintf(`Квиз запущен. Количество участников: %d`, participantsCnt)

	_, err = b.client.SendMessage(callback.Message.Chat.ID, msg, nil)
	if err != nil {
		return nil
	}

	go func() {
		for event := range events {
			switch event.Type {
			case engine.EventTypeQuestion:
				_ = b.handleQuestionEvent(runID, event)
			case engine.EventTypeFinished:
				_ = b.handleFinishedEvent(runID)
			}
		}
	}()

	return nil
}

// handleQuestionEvent отправляет каждому студенту вопрос со счетчиком времени.
func (b *Bot) handleQuestionEvent(runID string, event engine.QuizEvent) error {
	var builder strings.Builder

	text := fmt.Sprintf("Вопрос %d", event.QuestionIdx+1) + "\n\n"
	builder.WriteString(text)
	builder.WriteString(event.Question.Text + "\n\n")

	for i, option := range event.Question.Options {
		letter := engine.IndexToLetter(i)
		text = fmt.Sprintf("%s. %s", letter, option) + "\n"
		builder.WriteString(text)
	}

	b.mu.Lock()

	var questionTime int
	if event.Question.Time > 0 {
		questionTime = event.Question.Time
	} else {
		questionTime = b.runIDToQuiz[runID].Settings.TimePerQuestion
	}

	b.mu.Unlock()

	text = "\n" + fmt.Sprintf("Время: %d секунд", questionTime) + "\n\n"
	builder.WriteString(text)
	builder.WriteString("Отправьте букву ответа (A, B, C, ...)")
	msg := builder.String()

	userIDToBotMessage := make(map[int64]*client.Message)

	for userID, runIDForUser := range b.userIDToRunID {
		if runIDForUser == runID {
			b.mu.Lock()

			chatID := b.userIDToChatID[userID]
			b.mu.Unlock()

			botMessage, err := b.client.SendMessage(chatID, msg, nil)
			if err != nil {
				return err
			}

			userIDToBotMessage[userID] = botMessage
		}
	}

	go func() {
		_ = b.handleEditUserMessage(runID, userIDToBotMessage, event)
	}()

	return nil
}

// handleEditUserMessage изменяет счетчик времени в сообщении бота.
func (b *Bot) handleEditUserMessage(runID string, userIDToBotMessage map[int64]*client.Message, event engine.QuizEvent) error {
	ticker := time.NewTicker(time.Second)

	b.mu.Lock()

	var questionTime int
	if event.Question.Time > 0 {
		questionTime = event.Question.Time
	} else {
		questionTime = b.runIDToQuiz[runID].Settings.TimePerQuestion
	}

	b.mu.Unlock()

	lim := questionTime
	for range lim {
		<-ticker.C

		questionTime--

		var str strings.Builder

		text := fmt.Sprintf("Вопрос %d", event.QuestionIdx+1) + "\n\n"
		str.WriteString(text)
		str.WriteString(event.Question.Text + "\n\n")

		for i, option := range event.Question.Options {
			letter := engine.IndexToLetter(i)
			text = fmt.Sprintf("%s. %s", letter, option) + "\n"
			str.WriteString(text)
		}

		text = "\n" + fmt.Sprintf("Время: %d секунд", questionTime) + "\n\n"
		str.WriteString(text)
		str.WriteString("Отправьте букву ответа (A, B, C, ...)")
		msg := str.String()

		b.mu.Lock()

		for userID, runIDForUser := range b.userIDToRunID {
			if runIDForUser == runID {
				_ = b.client.EditMessage(
					b.userIDToChatID[userID],
					userIDToBotMessage[userID].MessageID,
					msg,
					nil,
				)
			}
		}

		b.mu.Unlock()
	}

	return nil
}

// handleFinishedEvent отправляет студентам и преподавателю результаты квиза.
func (b *Bot) handleFinishedEvent(runID string) error {
	res, err := b.engine.GetResults(runID)
	if err != nil {
		return err
	}

	var str strings.Builder
	str.WriteString("Топ-10:\n")

	limit := min(10, len(res.Leaderboard))
	for i := range limit {
		username := fmt.Sprintf("@%s", res.Leaderboard[i].Participant.Username)
		text := fmt.Sprintf("%d. %s - %d баллов\n", i+1, username, res.Leaderboard[i].Score)
		str.WriteString(text)
	}

	idx := 0
	text := str.String()

	for userID, runIDForUser := range b.userIDToRunID {
		if runIDForUser == runID {
			endText := "Квиз %s окончен! Результаты:\n\nВаш результат: %d баллов (место %d)\n\n%s"
			msg := fmt.Sprintf(
				endText,
				res.QuizTitle,
				res.Leaderboard[idx].Score,
				res.Leaderboard[idx].Rank,
				text,
			)

			b.mu.Lock()
			chatID := b.userIDToChatID[userID]
			b.mu.Unlock()

			_, err = b.client.SendMessage(chatID, msg, nil)
			if err != nil {
				return err
			}

			idx++
		}
	}

	msg := fmt.Sprintf(`Квиз %s окончен. Результаты квиза:`, res.QuizTitle)

	b.mu.Lock()
	ownerChatID := b.runIDToOwnerChatID[runID]
	b.mu.Unlock()

	_, err = b.sender.Message(ownerChatID, msg, nil)
	if err != nil {
		return err
	}

	csvData, err := b.engine.ExportCSV(runID)
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf(`Результаты квиза "%s"`, res.QuizTitle)

	return b.sender.Document(ownerChatID, fileName, csvData)
}
