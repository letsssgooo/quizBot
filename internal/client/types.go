//go:build !change

package client

import (
	"context"
	"time"
)

// Update представляет обновление от Telegram.
type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

// Message представляет сообщение.
type Message struct {
	MessageID int       `json:"message_id"`
	From      *User     `json:"from"`
	Chat      *Chat     `json:"chat"`
	Text      string    `json:"text"`
	Document  *Document `json:"document"`
}

// User представляет пользователя Telegram.
type User struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// Chat представляет чат.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// Document представляет документ (файл).
type Document struct {
	FileID   string `json:"file_id"`
	FileName string `json:"file_name"`
	MimeType string `json:"mime_type"`
	FileSize int    `json:"file_size"`
}

// CallbackQuery представляет callback от inline кнопки.
type CallbackQuery struct {
	ID      string   `json:"id"`
	From    *User    `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

// InlineKeyboardMarkup представляет inline клавиатуру.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// InlineKeyboardButton представляет кнопку inline клавиатуры.
type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

// SendOptions содержит опции отправки сообщения.
type SendOptions struct {
	ParseMode   string                `json:"parse_mode,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

// Client определяет интерфейс Telegram клиента.
type Client interface {
	// SendMessage отправляет сообщение.
	SendMessage(chatID int64, text string, opts *SendOptions) (*Message, error)

	// EditMessage редактирует сообщение.
	EditMessage(chatID int64, messageID int, text string, opts *SendOptions) error

	// DeleteMessage удаляет сообщение.
	DeleteMessage(chatID int64, messageID int) error

	// AnswerCallback отвечает на callback query.
	AnswerCallback(callbackID string, text string) error

	// GetUpdates получает обновления (long polling).
	GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error)

	// GetFile получает информацию о файле.
	GetFile(fileID string) (string, error)

	// DownloadFile скачивает файл по пути.
	DownloadFile(filePath string) ([]byte, error)

	// SendDocument отправляет файл как документ.
	SendDocument(chatID int64, fileName string, data []byte) error
}

// Таймауты
const (
	timeoutSend = 3 * time.Second
	timeoutDownload = 5 * time.Second
)