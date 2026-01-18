package fetcher

import "github.com/letsssgooo/quizBot/internal/client"

// Fetcher  определяет основной интерфейс для получения сообщений.
type Fetcher interface {
	// GetUpdates получает слайс Update, учитывая timeout
	GetUpdates(timeout int) ([]client.Update, error)
}
