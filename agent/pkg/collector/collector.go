package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"

	"siem-project/agent/pkg/config"
	"siem-project/agent/pkg/types"
)

// интерфейс для парсинга логов
type LogParser interface {
	Parse(line string, hostname string) (*types.Event, error)
	GetSourceType() string
}

type LogCollector struct {
	source     config.SourceConfig
	parser     LogParser
	hostname   string
	offsetFile string
	offset     int64
	file       *os.File
	watcher    *fsnotify.Watcher
	events     chan *types.Event
	mu         sync.Mutex
	stopCh     chan struct{}
}

// создает новый коллектор
func NewCollector(source config.SourceConfig, parser LogParser, hostname string) (*LogCollector, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	offsetFile := filepath.Join(".offsets", fmt.Sprintf("%s.offset", source.Type))
	os.MkdirAll(".offsets", 0755)

	collector := &LogCollector{
		source:     source,
		parser:     parser,
		hostname:   hostname,
		offsetFile: offsetFile,
		watcher:    watcher,
		events:     make(chan *types.Event, 100),
		stopCh:     make(chan struct{}),
	}

	collector.loadOffset()

	return collector, nil
}

func (c *LogCollector) Start() error {
	if err := c.openFile(); err != nil {
		return err
	}

	if err := c.watcher.Add(c.source.Path); err != nil {
		return fmt.Errorf("failed to watch file: %w", err)
	}

	go c.readExisting()
	go c.watchChanges()

	return nil
}

func (c *LogCollector) Stop() {
	close(c.stopCh)
	c.watcher.Close()
	if c.file != nil {
		c.file.Close()
	}
	close(c.events)
}

func (c *LogCollector) Events() <-chan *types.Event {
	return c.events
}

func (c *LogCollector) openFile() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file != nil {
		c.file.Close()
	}

	file, err := os.Open(c.source.Path)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", c.source.Path, err)
	}

	c.file = file
	return nil
}

func (c *LogCollector) readExisting() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file == nil {
		return
	}

	c.file.Seek(c.offset, io.SeekStart)

	scanner := bufio.NewScanner(c.file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		event, err := c.parser.Parse(line, c.hostname)
		if err != nil {
			continue
		}

		select {
		case c.events <- event:
		case <-c.stopCh:
			return
		}

		c.offset, _ = c.file.Seek(0, io.SeekCurrent)
		c.saveOffset()
	}
}

// отслеживает изменения в файле
func (c *LogCollector) watchChanges() {
	for {
		select {
		case event, ok := <-c.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				c.readNewLines()
			}

			if event.Op&fsnotify.Rename == fsnotify.Rename || event.Op&fsnotify.Remove == fsnotify.Remove {
				time.Sleep(100 * time.Millisecond)
				c.reopenFile()
			}

		case err, ok := <-c.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)

		case <-c.stopCh:
			return
		}
	}
}

func (c *LogCollector) readNewLines() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.file == nil {
		return
	}

	scanner := bufio.NewScanner(c.file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		event, err := c.parser.Parse(line, c.hostname)
		if err != nil {
			continue
		}

		select {
		case c.events <- event:
		case <-c.stopCh:
			return
		default:
			// канал заполнен, пропускаем
		}

		c.offset, _ = c.file.Seek(0, io.SeekCurrent)
		c.saveOffset()
	}
}

func (c *LogCollector) reopenFile() {
	c.openFile()
	c.offset = 0
	c.saveOffset()
}

func (c *LogCollector) loadOffset() {
	data, err := os.ReadFile(c.offsetFile)
	if err != nil {
		c.offset = 0
		return
	}
	fmt.Sscanf(string(data), "%d", &c.offset)
}

func (c *LogCollector) saveOffset() {
	os.WriteFile(c.offsetFile, []byte(fmt.Sprintf("%d", c.offset)), 0644)
}
