package fetcher

import "github.com/letsssgooo/quizBot/internal/client"

// TelegramFetcher реализует Fetcher через Telegram Bot API.
type TelegramFetcher struct {
	client client.Client
	offset int
}

func NewTelegramFetcher(client client.Client) *TelegramFetcher {
	return &TelegramFetcher{
		client: client,
		offset: 0,
	}
}

// GetUpdates получает слайс Update, учитывая timeout
func (f *TelegramFetcher) GetUpdates(timeout int) ([]client.Update, error) {
	updates, err := f.client.GetUpdates(f.offset, timeout)
	if err != nil {
		return nil, err
	}

	if len(updates) != 0 {
		f.offset = updates[len(updates)-1].UpdateID + 1
	}

	return updates, nil
}
