package agent

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"siem-project/agent/pkg/buffer"
	"siem-project/agent/pkg/collector"
	"siem-project/agent/pkg/config"
	"siem-project/agent/pkg/sender"
	"siem-project/agent/pkg/types"
)

type Agent struct {
	cfg        *config.Config
	collectors []*collector.LogCollector
	buffer     *buffer.RingBuffer
	sender     *sender.Sender
	stopCh     chan struct{}
	wg         sync.WaitGroup
	logFile    *os.File
}

func NewAgent(cfg *config.Config) (*Agent, error) {
	logFile, err := setupLogging(cfg.Logging.File)
	if err != nil {
		return nil, fmt.Errorf("failed to setup logging: %w", err)
	}

	// буфер
	buf := buffer.NewRingBuffer(cfg.Buffer.MemorySize, cfg.Buffer.DiskPath)

	snd := sender.NewSender(cfg)

	agent := &Agent{
		cfg:        cfg,
		collectors: make([]*collector.LogCollector, 0),
		buffer:     buf,
		sender:     snd,
		stopCh:     make(chan struct{}),
		logFile:    logFile,
	}

	// коллекторы для каждого источника
	if err := agent.initCollectors(); err != nil {
		return nil, err
	}

	return agent, nil
}

func (a *Agent) initCollectors() error {
	for _, source := range a.cfg.Sources {
		if !source.Enabled {
			log.Printf("Источник %s отключен, пропускаем", source.Type)
			continue
		}

		var parser collector.LogParser
		switch source.Type {
		case "bash_history":
			parser = collector.NewBashHistoryParser()
		case "syslog":
			parser = collector.NewSyslogParser("syslog")
		case "auth":
			parser = collector.NewSyslogParser("auth")
		case "auditd":
			parser = collector.NewAuditdParser()
		default:
			log.Printf("Неизвестный тип источника: %s, пропускаем", source.Type)
			continue
		}

		coll, err := collector.NewCollector(source, parser, a.cfg.Agent.Hostname)
		if err != nil {
			log.Printf("Не удалось создать коллектор для %s: %v", source.Type, err)
			continue
		}

		a.collectors = append(a.collectors, coll)
		log.Printf("Инициализирован коллектор для %s (%s)", source.Type, source.Path)
	}

	if len(a.collectors) == 0 {
		return fmt.Errorf("не удалось инициализировать ни одного коллектора")
	}

	return nil
}

func (a *Agent) Start() error {
	log.Printf("Запуск SIEM агента %s", a.cfg.Agent.ID)
	log.Printf("Сервер: %s:%d", a.cfg.Server.Host, a.cfg.Server.Port)
	log.Printf("База данных: %s/%s", a.cfg.Server.Database, a.cfg.Server.Collection)

	log.Printf("Проверка подключения к серверу...")
	if err := a.sender.TestConnection(); err != nil {
		log.Printf("Не удалось подключиться к серверу: %v", err)
		log.Printf("ℹПродолжаем работу, будем пытаться переподключиться при отправке")
	} else {
		log.Printf("Подключение к серверу успешно")
	}

	for _, coll := range a.collectors {
		if err := coll.Start(); err != nil {
			log.Printf("Ошибка запуска коллектора: %v", err)
			continue
		}
	}

	a.wg.Add(1)
	go a.processEvents()

	// периодическая отправка
	a.wg.Add(1)
	go a.periodicSend()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Получен сигнал %v, завершение работы...", sig)
		a.Stop()
	}()

	log.Printf("Агент успешно запущен")

	return nil
}

func (a *Agent) processEvents() {
	defer a.wg.Done()

	eventCh := make(chan *types.Event, 100)

	for _, coll := range a.collectors {
		a.wg.Add(1)
		go func(c *collector.LogCollector) {
			defer a.wg.Done()
			for {
				select {
				case event, ok := <-c.Events():
					if !ok {
						return
					}
					select {
					case eventCh <- event:
					case <-a.stopCh:
						return
					}
				case <-a.stopCh:
					return
				}
			}
		}(coll)
	}

	for {
		select {
		case event := <-eventCh:
			if err := a.buffer.Add(event); err != nil {
				log.Printf("Ошибка добавления события в буфер: %v", err)
			}

		case <-a.stopCh:
			log.Printf("Остановка обработки событий")
			return
		}
	}
}

// периодически отправляет события из буфера
func (a *Agent) periodicSend() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.cfg.Sender.SendInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.sendBatch()

		case <-a.stopCh:
			log.Printf("Остановка периодической отправки")
			a.sendBatch()
			return
		}
	}
}

// отправляет пакет событий
func (a *Agent) sendBatch() {
	bufferSize := a.buffer.Size()
	if bufferSize == 0 {
		return
	}

	log.Printf("Отправка событий из буфера (размер буфера: %d)", bufferSize)

	batch := a.buffer.GetBatch(a.cfg.Sender.MaxBatchSize)
	if len(batch) == 0 {
		return
	}

	if err := a.sender.SendEvents(batch); err != nil {
		log.Printf("Ошибка отправки: %v", err)
		log.Printf("События остаются в буфере для повторной отправки")
		return
	}

	a.buffer.Remove(len(batch))
	log.Printf("Успешно отправлено %d событий", len(batch))
}

func (a *Agent) Stop() {
	log.Printf("Остановка агента...")

	close(a.stopCh)

	// останавливаем коллекторы
	for _, coll := range a.collectors {
		coll.Stop()
	}

	a.wg.Wait()

	log.Printf("Сохранение буфера на диск...")
	if err := a.buffer.FlushToDisk(); err != nil {
		log.Printf("Ошибка сохранения буфера: %v", err)
	}

	if a.logFile != nil {
		a.logFile.Close()
	}

	log.Printf("Агент остановлен")
}

func (a *Agent) Wait() {
	a.wg.Wait()
}

func setupLogging(logPath string) (*os.File, error) {
	if err := os.MkdirAll("./logs", 0755); err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	return logFile, nil
}
