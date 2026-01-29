package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

const apiURL = "https://api.telegram.org/bot%s/%s"

// HTTPClient реализует Client через HTTP API Telegram.
type HTTPClient struct {
	token      string
	httpClient *http.Client
}

// NewHTTPClient создаёт нового HTTP клиента Telegram по переданному токену
func NewHTTPClient(token string) *HTTPClient {
	return &HTTPClient{
		token:      token,
		httpClient: &http.Client{},
	}
}

// SendMessage отправляет сообщение text в чат chatID.
// Возвращает указатель на структуру Message в случае успеха.
func (c *HTTPClient) SendMessage(
	chatID int64,
	text string,
	opts *SendOptions,
) (*Message, error) {
	params := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}

	if opts != nil {
		if opts.ParseMode != "" {
			params["parse_mode"] = opts.ParseMode
		}

		if opts.ReplyMarkup != nil {
			params["reply_markup"] = opts.ReplyMarkup
		}
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	rawResp, err := c.doRequest(ctx, "SendMessage", params)
	if err != nil {
		return nil, err
	}

	var message Message
	if err = json.Unmarshal(rawResp, &message); err != nil {
		return nil, err
	}

	return &message, nil
}

// EditMessage изменяет сообщение messageID на text в чате chatID.
// Возвращает nil в случае успеха.
func (c *HTTPClient) EditMessage(
	chatID int64,
	messageID int,
	text string,
	opts *SendOptions,
) error {
	params := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"message_id": messageID,
	}

	if opts != nil {
		if opts.ParseMode != "" {
			params["parse_mode"] = opts.ParseMode
		}

		if opts.ReplyMarkup != nil {
			params["reply_markup"] = opts.ReplyMarkup
		}
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	_, err := c.doRequest(ctx, "editMessageText", params)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMessage удаляет сообщение messageID в чате chatID.
// Возращает nil в случае успеха.
func (c *HTTPClient) DeleteMessage(chatID int64, messageID int) error {
	params := map[string]interface{}{
		"chat_id":    chatID,
		"message_id": messageID,
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	_, err := c.doRequest(ctx, "deleteMessage", params)
	if err != nil {
		return err
	}

	return nil
}

// AnswerCallback отвечает уведомлением в верхней части экрана чата (см. документацию
// telegram api) на callback query с идентификатором callbackID.
// Возращает nil в случае успеха.
func (c *HTTPClient) AnswerCallback(callbackID string, text string) error {
	params := map[string]interface{}{
		"callback_query_id": callbackID,
		"text":              text,
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	_, err := c.doRequest(ctx, "answerCallbackQuery", params)
	if err != nil {
		return err
	}

	return nil
}

// GetUpdates получает обновления.
// Если новых обновлений нет, ждёт до timeout секунд.
// Возвращает слайс Update.
// Для продолжения обработки нужно передать offset = lastUpdateID + 1.
func (c *HTTPClient) GetUpdates(ctx context.Context, offset int, timeout int) ([]Update, error) {
	params := map[string]interface{}{
		"offset":  offset,
		"timeout": timeout,
	}

	rawResp, err := c.doRequest(ctx, "getUpdates", params)
	if err != nil {
		return nil, err
	}

	var updates []Update
	if err = json.Unmarshal(rawResp, &updates); err != nil {
		return nil, err
	}

	return updates, nil
}

// GetFile получает информацию о файле с идентификатором fileID.
// Возращает путь файла в случае успеха.
func (c *HTTPClient) GetFile(fileID string) (string, error) {
	params := map[string]interface{}{
		"file_id": fileID,
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	rawResp, err := c.doRequest(ctx, "getFile", params)
	if err != nil {
		return "", err
	}

	var file struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size"`
		FilePath     string `json:"file_path"`
	}

	if err = json.Unmarshal(rawResp, &file); err != nil {
		return "", err
	}

	return file.FilePath, nil
}

// DownloadFile скачивает файл с путем filePath.
// Возращает содержимое файла в случае успеха.
func (c *HTTPClient) DownloadFile(filePath string) ([]byte, error) {
	linkForDownload := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, filePath)

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutDownload)
	defer cancelFunc()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, linkForDownload, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download the request from the link %s: %w",
			linkForDownload,
			err,
		)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"unexpected response status code %d for link %s",
			resp.StatusCode,
			linkForDownload,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body in DownloadFile: %w", err)
	}

	return data, nil
}

// SendDocument отправляет файл с названием fileName и содержимым data в чат chatID как документ.
// Возвращает nil в случае успеха.
func (c *HTTPClient) SendDocument(
	chatID int64,
	fileName string,
	data []byte,
) error {
	var buf bytes.Buffer

	writer := multipart.NewWriter(&buf)

	defer func() {
		_ = writer.Close()
	}()

	err := writer.WriteField("chat_id", strconv.FormatInt(chatID, 10))
	if err != nil {
		return fmt.Errorf("failed to add chat_id field to multipart form: %w", err)
	}

	multipartWriter, err := writer.CreateFormFile("document", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err = multipartWriter.Write(data); err != nil {
		return fmt.Errorf("failed to write data to multipart form: %w", err)
	}

	url := fmt.Sprintf(apiURL, c.token, "sendDocument")

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeoutSend)
	defer cancelFunc()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to do post request for url %s: %w", url, err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body in SendDocument: %w", err)
	}

	var result struct {
		OK    bool   `json:"ok"`
		Error string `json:"description"`
	}
	if err = json.Unmarshal(respData, &result); err != nil {
		return err
	}

	if !result.OK {
		return fmt.Errorf("client api error: %s", result.Error)
	}

	return nil
}

// doRequest выполняет запрос к Telegram API.
// Возвращает результат запроса в случае успеха.
func (c *HTTPClient) doRequest(
	ctx context.Context,
	method string,
	params map[string]interface{},
) (json.RawMessage, error) {
	url := fmt.Sprintf(apiURL, c.token, method)

	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
		Error  string          `json:"description"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	if !result.OK {
		return nil, fmt.Errorf("client api error: %s", result.Error)
	}

	return result.Result, nil
}
