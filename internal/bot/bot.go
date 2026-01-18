package bot

import (
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
	"github.com/letsssgooo/quizBot/internal/events/fetcher"
	"github.com/letsssgooo/quizBot/internal/events/sender"
)

const updatesTimeout = 30

// Bot реализует Telegram бота для квизов.
type Bot struct {
	client      client.Client
	fetcher     fetcher.Fetcher
	sender      sender.Sender
	engine      engine.QuizEngine
	botUsername string // Username бота для формирования ссылок (например, "my_quiz_bot")
}

// NewBot создаёт нового бота.
// botUsername — username бота без @ (например, "my_quiz_bot").
// Используется для формирования ссылок: https://t.me/<botUsername>?start=join_<runID>
func NewBot(
	client client.Client,
	fetcher fetcher.Fetcher,
	sender sender.Sender,
	engine engine.QuizEngine,
	botUsername string,
) *Bot {
	return &Bot{
		client:      client,
		fetcher:     fetcher,
		sender:      sender,
		engine:      engine,
		botUsername: botUsername,
	}
}

// Run запускает бота (long polling).
func (b *Bot) Run() error {
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

		// TODO: sender опредяеляет, кому надо отдать обработанные апдейты, и отдает
	}
}

// HandleUpdate обрабатывает одно обновление.
func (b *Bot) HandleUpdate(update client.Update) error {
	panic("not implemented")

	// TODO: HandleUpdate проверяет права (пр: студент не может изменять квиз) и определяет
	// TODO: тип апдейта и что с ним должен сделать engine

	// TODO: новые апдейты попадают в engine, он проверяет их на корректность, обрабатывает
	// TODO: и отдает готовый результат в sender
}
