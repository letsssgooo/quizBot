package fetcher

import (
	"context"

	"github.com/letsssgooo/quizBot/internal/client"
)

// Fetcher  определяет основной интерфейс для получения сообщений.
type Fetcher interface {
	// GetUpdates получает слайс Update, учитывая timeout
	GetUpdates(ctx context.Context, timeout int) ([]client.Update, error)
}
