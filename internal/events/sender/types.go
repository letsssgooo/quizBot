package sender

import "github.com/letsssgooo/quizBot/internal/client"

// Sender определяет основной интерфейс для отправки сообщений.
type Sender interface {
	// Message отправляет текстовое сообщение.
	Message(chatID int64, text string, opts *client.SendOptions) (*client.Message, error)

	// Document отправляет файл как документ.
	Document(chatID int64, fileName string, data []byte) error
}
