package sender

import "github.com/letsssgooo/quizBot/internal/client"

// Sender реализует отправку сообщений через Telegram Bot API.
type Sender struct {
	client client.Client
}

// NewSender создает новый объект структуры Sender.
func NewSender(client client.Client) *Sender {
	return &Sender{client: client}
}

// Message отправляет текстовое сообщение.
func (s *Sender) Message(chatID int64, text string, opts *client.SendOptions) (*client.Message, error) {
	return s.client.SendMessage(chatID, text, opts)
}

// Document отправляет файл как документ.
func (s *Sender) Document(chatID int64, fileName string, data []byte) error {
	return s.client.SendDocument(chatID, fileName, data)
}
