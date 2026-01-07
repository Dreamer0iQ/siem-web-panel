// prac3
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Agent   AgentConfig    `yaml:"agent"`
	Logging LoggingConfig  `yaml:"logging"`
	Sources []SourceConfig `yaml:"sources"`
	Buffer  BufferConfig   `yaml:"buffer"`
	Sender  SenderConfig   `yaml:"sender"`
}

type ServerConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
}

type AgentConfig struct {
	ID       string `yaml:"id"`
	Hostname string `yaml:"hostname"`
}

type LoggingConfig struct {
	File string `yaml:"file"`
}

type SourceConfig struct {
	Type    string `yaml:"type"`
	Path    string `yaml:"path"`
	Enabled bool   `yaml:"enabled"`
}

type BufferConfig struct {
	MemorySize int    `yaml:"memory_size"`
	DiskPath   string `yaml:"disk_path"`
}

type SenderConfig struct {
	MaxBatchSize  int `yaml:"max_batch_size"`
	SendInterval  int `yaml:"send_interval"`
	RetryInterval int `yaml:"retry_interval"`
	MaxRetries    int `yaml:"max_retries"`
}

// Load загружает конфигурацию из YAML файла
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg.applyEnvOverrides()

	if cfg.Agent.Hostname == "" {
		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}
		cfg.Agent.Hostname = hostname
	}

	for i := range cfg.Sources {
		cfg.Sources[i].Path = expandPath(cfg.Sources[i].Path)
	}
	cfg.Logging.File = expandPath(cfg.Logging.File)
	cfg.Buffer.DiskPath = expandPath(cfg.Buffer.DiskPath)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if host := os.Getenv("SIEM_SERVER_HOST"); host != "" {
		c.Server.Host = host
		fmt.Printf("Server host overridden from env: %s\n", host)
	}

	if portStr := os.Getenv("SIEM_SERVER_PORT"); portStr != "" {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil && port > 0 {
			c.Server.Port = port
			fmt.Printf("Server port overridden from env: %d\n", port)
		}
	}

	if agentID := os.Getenv("SIEM_AGENT_ID"); agentID != "" {
		c.Agent.ID = agentID
		fmt.Printf("Agent ID overridden from env: %s\n", agentID)
	}
}

func (c *Config) Validate() error {
	if c.Server.Host == "" {
		return fmt.Errorf("server.host is required")
	}
	if c.Server.Port <= 0 {
		return fmt.Errorf("server.port must be positive")
	}
	if c.Agent.ID == "" {
		return fmt.Errorf("agent.id is required")
	}
	if len(c.Sources) == 0 {
		return fmt.Errorf("at least one source must be configured")
	}
	return nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}
	return path
}
