package sender

import (
	"github.com/letsssgooo/quizBot/internal/client"
)

// TelegramSender реализует Sender через Telegram Bot API.
type TelegramSender struct {
	client client.Client
}

// NewTelegramSender создает новый объект структуры TelegramSender.
func NewTelegramSender(client client.Client) *TelegramSender {
	return &TelegramSender{client: client}
}

// Message отправляет текстовое сообщение.
func (s *TelegramSender) Message(chatID int64, text string, opts *client.SendOptions) (*client.Message, error) {
	return s.client.SendMessage(chatID, text, opts)
}

// Document отправляет файл как документ.
func (s *TelegramSender) Document(chatID int64, fileName string, data []byte) error {
	return s.client.SendDocument(chatID, fileName, data)
}
