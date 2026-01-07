// prac3
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"siem-project/agent/pkg/agent"
	"siem-project/agent/pkg/config"
)

func main() {
	configPath := flag.String("config", "configs/agent.yaml", "Путь к конфигурационному файлу")
	serverHost := flag.String("server-host", "", "Адрес сервера (переопределяет config)")
	serverPort := flag.Int("server-port", 0, "Порт сервера (переопределяет config)")
	agentID := flag.String("agent-id", "", "ID агента (переопределяет config)")
	flag.Parse()

	if *serverHost != "" {
		os.Setenv("SIEM_SERVER_HOST", *serverHost)
	}
	if *serverPort > 0 {
		os.Setenv("SIEM_SERVER_PORT", fmt.Sprintf("%d", *serverPort))
	}
	if *agentID != "" {
		os.Setenv("SIEM_AGENT_ID", *agentID)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Printf("Ошибка загрузки конфигурации: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Starting SIEM Agent\n")
	fmt.Printf("Agent ID: %s\n", cfg.Agent.ID)
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Println()

	ag, err := agent.NewAgent(cfg)
	if err != nil {
		log.Fatalf("Ошибка создания агента: %v", err)
	}

	if err := ag.Start(); err != nil {
		log.Fatalf("Ошибка запуска агента: %v", err)
	}

	select {}
}
