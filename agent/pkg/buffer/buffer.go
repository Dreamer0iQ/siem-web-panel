package buffer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"siem-project/agent/pkg/types"
)

// кольцевой буфер для хранения событий
type RingBuffer struct {
	events   []*types.Event
	size     int
	head     int
	tail     int
	count    int
	diskPath string
	mu       sync.Mutex
}

func NewRingBuffer(size int, diskPath string) *RingBuffer {
	rb := &RingBuffer{
		events:   make([]*types.Event, size),
		size:     size,
		diskPath: diskPath,
	}

	if diskPath != "" {
		if err := rb.LoadFromDisk(); err == nil {
			fmt.Printf("Восстановлено %d событий из буфера на диске\n", rb.count)
		}
	}

	return rb
}

func (rb *RingBuffer) Add(event *types.Event) error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	// буфер полон, сбрасываем на диск и очищаем
	if rb.count >= rb.size {
		if err := rb.flushToDiskUnsafe(); err != nil {
			return fmt.Errorf("failed to flush buffer: %w", err)
		}
		rb.count = 0
		rb.head = 0
		rb.tail = 0
	}

	rb.events[rb.head] = event
	rb.head = (rb.head + 1) % rb.size
	rb.count++

	return nil
}

func (rb *RingBuffer) GetBatch(n int) []*types.Event {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	batchSize := n
	if batchSize > rb.count {
		batchSize = rb.count
	}

	batch := make([]*types.Event, batchSize)
	for i := 0; i < batchSize; i++ {
		idx := (rb.tail + i) % rb.size
		batch[i] = rb.events[idx]
	}

	return batch
}

func (rb *RingBuffer) Remove(n int) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if n > rb.count {
		n = rb.count
	}

	rb.tail = (rb.tail + n) % rb.size
	rb.count -= n
}

func (rb *RingBuffer) Size() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

func (rb *RingBuffer) FlushToDisk() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.flushToDiskUnsafe()
}

func (rb *RingBuffer) flushToDiskUnsafe() error {
	if rb.diskPath == "" || rb.count == 0 {
		return nil
	}

	// Собираем все события
	events := make([]*types.Event, rb.count)
	for i := 0; i < rb.count; i++ {
		idx := (rb.tail + i) % rb.size
		events[i] = rb.events[idx]
	}

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal events: %w", err)
	}

	// Создаём директорию, если её нет
	// diskPath может быть "./buffer" или "./buffer/events.json"
	// Используем filepath.Dir для безопасного получения директории
	dir := filepath.Dir(rb.diskPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Формируем полный путь к файлу
	bufferFile := rb.diskPath
	if stat, err := os.Stat(rb.diskPath); err == nil && stat.IsDir() {
		bufferFile = filepath.Join(rb.diskPath, "buffer.json")
	}

	if err := os.WriteFile(bufferFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write to disk: %w", err)
	}

	return nil
}

func (rb *RingBuffer) LoadFromDisk() error {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.diskPath == "" {
		return nil
	}

	// Определяем путь к файлу буфера
	bufferFile := rb.diskPath
	if stat, err := os.Stat(rb.diskPath); err == nil && stat.IsDir() {
		bufferFile = filepath.Join(rb.diskPath, "buffer.json")
	}

	// Проверяем существование файла
	if _, err := os.Stat(bufferFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(bufferFile)
	if err != nil {
		return fmt.Errorf("failed to read from disk: %w", err)
	}

	var events []*types.Event
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed to unmarshal events: %w", err)
	}

	for _, event := range events {
		if rb.count >= rb.size {
			break
		}
		rb.events[rb.head] = event
		rb.head = (rb.head + 1) % rb.size
		rb.count++
	}

	// Удаляем файл после загрузки
	os.Remove(bufferFile)

	return nil
}

func (rb *RingBuffer) Clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.head = 0
	rb.tail = 0
	rb.count = 0
}
