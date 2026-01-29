package fetcher

import (
	"context"

	"github.com/letsssgooo/quizBot/internal/client"
)

// TelegramFetcher реализует Fetcher через Telegram Bot API.
type TelegramFetcher struct {
	client client.Client
	offset int
}

// NewTelegramFetcher возвращает *TelegramFetcher
func NewTelegramFetcher(ctx context.Context, client client.Client) *TelegramFetcher {
	updates, _ := client.GetUpdates(ctx, 0, 0)

	offset := 0
	if len(updates) > 0 {
		offset = updates[len(updates)-1].UpdateID + 1
	}

	return &TelegramFetcher{
		client: client,
		offset: offset,
	}
}

// GetUpdates получает слайс Update, учитывая timeout
func (f *TelegramFetcher) GetUpdates(ctx context.Context, timeout int) ([]client.Update, error) {
	updates, err := f.client.GetUpdates(ctx, f.offset, timeout)
	if err != nil {
		return nil, err
	}

	if len(updates) != 0 {
		f.offset = updates[len(updates)-1].UpdateID + 1
	}

	return updates, nil
}
