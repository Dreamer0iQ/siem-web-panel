package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"siem-project/agent/pkg/config"
	"siem-project/agent/pkg/types"
)

type Sender struct {
	cfg        *config.Config
	httpClient *http.Client
	serverURL  string
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		serverURL: fmt.Sprintf("http://%s:%d/query", cfg.Server.Host, cfg.Server.Port),
	}
}

func (s *Sender) SendEvents(events []*types.Event) error {
	if len(events) == 0 {
		return nil
	}

	message := types.NewMessage(s.cfg.Agent.ID, events)
	return s.sendWithRetry(message)
}

type insertRequest struct {
	Database   string            `json:"database"`
	Collection string            `json:"collection"`
	Events     []json.RawMessage `json:"events"`
}

type serverResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (s *Sender) sendWithRetry(message *types.Message) error {
	var lastErr error

	for attempt := 0; attempt <= s.cfg.Sender.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Повторная попытка отправки (%d/%d)...", attempt, s.cfg.Sender.MaxRetries)
			time.Sleep(time.Duration(s.cfg.Sender.RetryInterval) * time.Second)
		}

		err := s.sendRequest(message)
		if err == nil {
			if attempt > 0 {
				log.Printf("Успешно отправлено после %d попыток", attempt+1)
			}
			return nil
		}

		lastErr = err
		log.Printf("Ошибка отправки: %v", err)
	}

	return fmt.Errorf("не удалось отправить после %d попыток: %w", s.cfg.Sender.MaxRetries+1, lastErr)
}

func (s *Sender) sendRequest(message *types.Message) error {
	var eventData []json.RawMessage

	for _, event := range message.Events {
		eventJSON, err := json.Marshal(event)
		if err != nil {
			log.Printf("Ошибка маршалинга: %v", err)
			continue
		}
		eventData = append(eventData, eventJSON)
	}

	reqBody := insertRequest{
		Database:   s.cfg.Server.Database,
		Collection: s.cfg.Server.Collection,
		Events:     eventData,
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", s.serverURL, bytes.NewReader(reqData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response serverResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status != "success" {
		return fmt.Errorf("server error: %s", response.Message)
	}

	log.Printf("Отправлено %d событий в %s/%s", len(message.Events), s.cfg.Server.Database, s.cfg.Server.Collection)

	return nil
}

func (s *Sender) TestConnection() error {
	httpReq, err := http.NewRequest("GET", s.serverURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("сервер вернул код %d", resp.StatusCode)
	}

	return nil
}
