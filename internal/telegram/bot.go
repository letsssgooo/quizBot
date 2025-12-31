//go:build !solution

package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"gitlab.com/slon/shad-go/Exam-1-QuizBot/quizbot/internal/quiz"
)

// Bot реализует Telegram бота для квизов.
type Bot struct {
	client      Client
	engine      quiz.QuizEngine
	botUsername string // Username бота для формирования ссылок (например, "my_quiz_bot")
	offset      int
	token       string
	lobby       *Lobby
	quizRun     *quiz.QuizRun
}

type Lobby struct {
	tgUserIDToChatID map[int64]int64
}

// SendDocument отправляет файл как документ.
func (b *Bot) SendDocument(chatID int64, fileName string, data []byte) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("chat_id", fmt.Sprint(chatID))
	multipartWriter, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return err
	}
	if _, err = multipartWriter.Write(data); err != nil {
		return err
	}
	writer.Close()

	url := fmt.Sprintf(apiURL, b.token, "sendDocument")
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"description"`
	}

	if err = json.Unmarshal(respData, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("telegram api error: %s", result.Error)
	}

	return nil
}

// NewBot создаёт нового бота.
// botUsername — username бота без @ (например, "my_quiz_bot").
// Используется для формирования ссылок: https://t.me/<botUsername>?start=join_<runID>
func NewBot(client Client, engine quiz.QuizEngine, botUsername string) *Bot {
	return &Bot{
		client:      client,
		engine:      engine,
		botUsername: botUsername,
	}
}

// Run запускает бота (long polling).
func (b *Bot) Run() error {
	for {
		updates, err := b.client.GetUpdates(b.offset, 60)
		if err != nil {
			return err
		}
		for _, update := range updates {
			if err = b.HandleUpdate(update); err != nil {
				slog.Error("error while getting update: ", "update_id: ", update.UpdateID, "err", err)
			}
			b.offset = update.UpdateID + 1
		}
	}
}

// HandleUpdate обрабатывает одно обновление.
func (b *Bot) HandleUpdate(update Update) error {
	if update.Message != nil {
		if update.Message.From.Username == "ivans_tg" {
			return b.HandleTeacher(update)
		}
		return b.HandleStudentMessage(update)
	}
	if update.CallbackQuery != nil {
		return b.HandleTeacher(update)
	}
	return nil
}

func (b *Bot) HandleTeacher(update Update) error {
	if update.Message != nil && update.Message.Document != nil {
		filePath, err := b.client.GetFile(update.Message.Document.FileID)
		if err != nil {
			return err
		}
		data, err := b.client.DownloadFile(filePath)
		if err != nil {
			return err
		}

		newQuiz, err := b.engine.LoadQuiz(data)
		if err != nil {
			return err
		}

		ctx := context.Background()
		b.quizRun, err = b.engine.StartRun(ctx, newQuiz)
		if err != nil {
			return err
		}

		b.lobby = &Lobby{
			tgUserIDToChatID: make(map[int64]int64),
		}

		quizLink := fmt.Sprintf("https://t.me/%s?start=join_%s", b.botUsername, b.quizRun.ID)
		_, err = b.client.SendMessage(update.Message.Chat.ID, fmt.Sprintf("ссылка для подключения: %s", quizLink), nil)
		if err != nil {
			return err
		}

		keyboard := &InlineKeyboardMarkup{[][]InlineKeyboardButton{
			{
				InlineKeyboardButton{
					Text:         "Начать квиз",
					CallbackData: "кнопка начала квиза",
				},
			},
		}}

		usersConnectedMsg, err := b.client.SendMessage(update.Message.Chat.ID, "Количество подключившихся: 0", &SendOptions{ReplyMarkup: keyboard})
		if err != nil {
			return err
		}

		go func() {
			ticker := time.NewTicker(time.Second * 3)
			defer ticker.Stop()

			previousUsersCount := 0

			for range ticker.C {
				newCount := b.engine.GetParticipantCount(b.quizRun.ID)
				if newCount > previousUsersCount {
					previousUsersCount = newCount
					err = b.client.EditMessage(usersConnectedMsg.Chat.ID, usersConnectedMsg.MessageID, fmt.Sprintf("Количество подключившихся: %d", newCount), &SendOptions{ReplyMarkup: keyboard})
				}
				if err != nil {
					return
				}

				if b.quizRun.Status != quiz.RunStatusLobby {
					return
				}
			}
		}()
	}

	if update.CallbackQuery != nil && b.quizRun != nil {
		slog.Info("Квиз начат")
		ctx := context.Background()
		quizEvent, err := b.engine.StartQuiz(ctx, b.quizRun.ID)
		if err != nil {
			return err
		}

		go func() {
			for event := range quizEvent {
				for _, chatID := range b.lobby.tgUserIDToChatID {
					switch event.Type {
					case quiz.EventTypeQuestion:
						_, err = b.client.SendMessage(chatID, event.Question.Text, nil)
						if err != nil {
							return
						}
					case quiz.EventTypeTimeUp:
						_, err = b.client.SendMessage(chatID, "время на вопрос закончилось", nil)
						if err != nil {
							return
						}
					case quiz.EventTypeFinished:
						_, err = b.client.SendMessage(chatID, "квиз финишировал", nil)
						if err != nil {
							return
						}
					}
				}
			}
		}()
	}

	return nil
}

func (b *Bot) HandleStudentMessage(update Update) error {
	if strings.HasPrefix(update.Message.Text, "/start") {
		runID := strings.Split(update.Message.Text, " join_")[1]
		ctx := context.Background()
		participant := &quiz.Participant{
			TelegramID: update.Message.From.ID,
			Username:   update.Message.From.Username,
			FirstName:  update.Message.From.FirstName,
			LastName:   update.Message.From.LastName,
			JoinedAt:   time.Now(),
		}
		err := b.engine.JoinRun(ctx, runID, participant)
		if err != nil {
			return err
		}

		b.lobby.tgUserIDToChatID[participant.TelegramID] = update.Message.Chat.ID
	}

	ctx := context.Background()
	err := b.engine.SubmitAnswerByLetter(ctx, b.quizRun.ID, update.Message.From.ID, update.Message.Text)
	if err != nil {
		return err
	}
	return nil
}
