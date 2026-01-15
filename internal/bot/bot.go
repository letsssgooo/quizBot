//go:build !solution

package bot

import (
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
)

// Bot реализует Telegram бота для квизов.
type Bot struct {
	client      client.Client
	engine      engine.QuizEngine
	botUsername string // Username бота для формирования ссылок (например, "my_quiz_bot")
	// TODO: добавьте необходимые поля
}

// NewBot создаёт нового бота.
// botUsername — username бота без @ (например, "my_quiz_bot").
// Используется для формирования ссылок: https://t.me/<botUsername>?start=join_<runID>
func NewBot(client client.Client, engine engine.QuizEngine, botUsername string) *Bot {
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
func (b *Bot) HandleUpdate(update client.Update) error {
	panic("not implemented")
}
