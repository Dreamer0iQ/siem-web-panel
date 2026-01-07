package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Event struct {
	ID          string                 `json:"id"`
	Timestamp   string                 `json:"timestamp"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Host        string                 `json:"host"`
	Severity    string                 `json:"severity"`
	Process     string                 `json:"process"`
	Description string                 `json:"description"`
	User        string                 `json:"user,omitempty"`
	Details     map[string]interface{} `json:"details"`
}

func (e *Event) UnmarshalJSON(data []byte) error {
	type Alias Event
	aux := &struct {
		FileID        string `json:"_id"`
		FileEventType string `json:"event_type"`
		FileHostname  string `json:"hostname"`
		FileRawLog    string `json:"raw_log"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if e.ID == "" {
		e.ID = aux.FileID
	}
	if e.Type == "" {
		e.Type = aux.FileEventType
	}
	if e.Host == "" {
		e.Host = aux.FileHostname
	}
	if e.Description == "" {
		e.Description = aux.FileRawLog
	}

	return nil
}

type Storage struct {
	dataDir string
	events  []*Event
	mu      sync.RWMutex
}

func NewStorage(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	storage := &Storage{
		dataDir: dataDir,
		events:  make([]*Event, 0),
	}

	// Загрузка существующих событий
	if err := storage.loadEvents(); err != nil {
		return nil, err
	}

	return storage, nil
}

// AddEvent добавляет новое событие
func (s *Storage) AddEvent(event *Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}

	s.events = append(s.events, event)

	return s.saveEvents()
}

func (s *Storage) AddEvents(events []*Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, event := range events {
		if event.ID == "" {
			event.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
		}
		s.events = append(s.events, event)
	}

	return s.saveEvents()
}

func (s *Storage) GetEvents(filter EventFilter) ([]*Event, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Event, 0)

	for _, event := range s.events {
		if filter.Matches(event) {
			result = append(result, event)
		}
	}

	total := len(result)

	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})

	if filter.Limit > 0 {
		start := 0
		if filter.Page > 1 {
			start = (filter.Page - 1) * filter.Limit
		}

		if start >= len(result) {
			result = []*Event{}
		} else {
			end := start + filter.Limit
			if end > len(result) {
				end = len(result)
			}
			result = result[start:end]
		}
	}

	return result, total, nil
}

func (s *Storage) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total_events": len(s.events),
	}

	severityCount := make(map[string]int)
	sourceCount := make(map[string]int)
	typeCount := make(map[string]int)

	for _, event := range s.events {
		severityCount[event.Severity]++
		sourceCount[event.Source]++
		typeCount[event.Type]++
	}

	stats["by_severity"] = severityCount
	stats["by_source"] = sourceCount
	stats["by_type"] = typeCount

	if len(s.events) > 0 {
		stats["last_event"] = s.events[len(s.events)-1].Timestamp
	}

	return stats
}

func (s *Storage) DeleteOldEvents(olderThan time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]*Event, 0)
	deleted := 0

	for _, event := range s.events {
		eventTime, err := time.Parse(time.RFC3339, event.Timestamp)
		if err != nil || eventTime.After(olderThan) {
			filtered = append(filtered, event)
		} else {
			deleted++
		}
	}

	s.events = filtered

	if deleted > 0 {
		return s.saveEvents()
	}

	return nil
}

type EventFilter struct {
	Source   string
	Severity string
	Hostname string
	Type     string // Filter by Event Type
	User     string // Filter by User
	Process  string // Filter by Process
	From     string
	To       string
	Limit    int
	Page     int
}

func (f *EventFilter) Matches(event *Event) bool {
	if f.Source != "" && event.Source != f.Source {
		return false
	}

	if f.Severity != "" && event.Severity != f.Severity {
		return false
	}

	if f.Hostname != "" && event.Host != f.Hostname {
		return false
	}

	if f.Type != "" && event.Type != f.Type {
		return false
	}

	if f.User != "" && event.User != f.User {
		return false
	}

	if f.Process != "" && event.Process != f.Process {
		return false
	}

	if f.From != "" && event.Timestamp < f.From {
		return false
	}

	if f.To != "" && event.Timestamp > f.To {
		return false
	}

	return true
}

// loadEvents загружает события из файла
func (s *Storage) loadEvents() error {
	filePath := filepath.Join(s.dataDir, "security", "security_events.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read events file: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	if err := json.Unmarshal(data, &s.events); err == nil {
		return nil
	}

	eventsMap := make(map[string]*Event)
	if err := json.Unmarshal(data, &eventsMap); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}

	s.events = make([]*Event, 0, len(eventsMap))
	for _, event := range eventsMap {
		s.events = append(s.events, event)
	}

	sort.Slice(s.events, func(i, j int) bool {
		return s.events[i].Timestamp < s.events[j].Timestamp
	})
	return nil
}

func (s *Storage) saveEvents() error {
	dirPath := filepath.Join(s.dataDir, "security")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "security_events.json")

	eventsMap := make(map[string]*Event)
	for _, event := range s.events {
		eventsMap[event.ID] = event
	}

	data, err := json.MarshalIndent(eventsMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write events file: %w", err)
	}

	return nil
}
