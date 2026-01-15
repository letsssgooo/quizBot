//go:build !solution

package telegram

import (
	"github.com/letsssgooo/quizBot/internal/quiz"
)

// Bot реализует Telegram бота для квизов.
type Bot struct {
	client      Client
	engine      quiz.QuizEngine
	botUsername string // Username бота для формирования ссылок (например, "my_quiz_bot")
	// TODO: добавьте необходимые поля
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
	panic("not implemented")
}

// HandleUpdate обрабатывает одно обновление.
func (b *Bot) HandleUpdate(update Update) error {
	panic("not implemented")
}
