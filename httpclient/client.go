package httpclient

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/KharpukhaevV/postui/models"
)

// HTTPClient обрабатывает HTTP запросы
type HTTPClient struct {
	client *http.Client
}

// NewHTTPClient создает новый HTTP клиент с настройками по умолчанию
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// SendRequest отправляет HTTP запрос и возвращает данные ответа
func (c *HTTPClient) SendRequest(req *HTTPRequest) (models.ResponseData, error) {
	start := time.Now()

	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return models.ResponseData{}, fmt.Errorf("неверный URL: %w", err)
	}

	// Добавляем параметры запроса
	query := parsedURL.Query()
	for _, p := range req.Params {
		query.Add(p.Key, p.Value)
	}
	parsedURL.RawQuery = query.Encode()

	// Создаем запрос
	httpReq, err := http.NewRequest(req.Method, parsedURL.String(), bytes.NewReader(req.Body))
	if err != nil {
		return models.ResponseData{}, fmt.Errorf("не удалось создать запрос: %w", err)
	}

	// Добавляем заголовки
	for _, h := range req.Headers {
		httpReq.Header.Add(h.Key, h.Value)
	}

	// Выполняем запрос
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return models.ResponseData{}, fmt.Errorf("запрос не выполнен: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	elapsed := time.Since(start).Round(time.Millisecond)

	return models.ResponseData{
		Body:       responseBody,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Time:       elapsed.String(),
	}, nil
}

// HTTPRequest представляет HTTP запрос
type HTTPRequest struct {
	Method  string
	URL     string
	Headers []models.Header
	Params  []models.Param
	Body    []byte
}

// NewHTTPRequest создает новый HTTP запрос из модели приложения
func NewHTTPRequest(model *models.AppModel) HTTPRequest {
	var bodyBytes []byte
	if model.BodyInputValue() != "" {
		bodyBytes = []byte(model.BodyInputValue())
	}

	return HTTPRequest{
		Method:  model.GetCurrentMethod(),
		URL:     model.URLInputValue(),
		Headers: model.GetHeaders(),
		Params:  model.GetParams(),
		Body:    bodyBytes,
	}
}
