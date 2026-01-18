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
	for { // long polling
		panic("not implemented")
		// TODO: fetcher получает новые апдейты, учитывая offset, limit, timeout, и отдает их

		// TODO: HandleUpdate проверяет права (пр: студент не может изменять квиз) и определяет
		// TODO: тип апдейта и что с ним должен сделать engine

		// TODO: новые апдейты попадают в engine, он проверяет их на корректность, обрабатывает
		// TODO: и отдает готовый результат в sender

		// TODO: sender опредяеляет, кому надо отдать обработанные апдейты, и отдает
	}
}

// HandleUpdate обрабатывает одно обновление.
func (b *Bot) HandleUpdate(update client.Update) error {
	panic("not implemented")

	// TODO: проверить права доступа и определить дальнейший маршрут для update
}
