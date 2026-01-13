.PHONY: help run-server run-agent run-frontend install-deps build-all clean docker-build docker-up docker-down docker-logs docker-clean docker-rebuild

help:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🎯 SIEM Platform - Команды для запуска"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📦 Установка:"
	@echo "  make install-deps    - Установить все зависимости"
	@echo ""
	@echo "🚀 Запуск (нужно 3 терминала):"
	@echo "  make run-server      - Запустить Backend API (терминал 1)"
	@echo "  make run-frontend    - Запустить Frontend (терминал 2)"
	@echo "  make run-agent       - Запустить Agent (терминал 3)"
	@echo ""
	@echo "🐳 Docker команды:"
	@echo "  make docker-build    - Собрать все Docker образы"
	@echo "  make docker-up       - Запустить все сервисы"
	@echo "  make docker-down     - Остановить все сервисы"
	@echo "  make docker-logs     - Показать логи всех сервисов"
	@echo "  make docker-clean    - Удалить все контейнеры и volumes"
	@echo "  make docker-rebuild  - Пересобрать и перезапустить"
	@echo ""
	@echo "📡 Запуск Agent с динамическим IP:"
	@echo "  make run-agent-auto  - С автоопределением IP сервера"
	@echo "  ./agent/run-agent-auto.sh  - То же самое (скрипт)"
	@echo "  ./agent/run-agent.sh       - Интерактивный запуск"
	@echo "  make run-agent-with HOST=192.168.1.100 PORT=8080"
	@echo ""
	@echo "💡 Примеры:"
	@echo "  SIEM_SERVER_HOST=192.168.1.100 make run-agent"
	@echo "  cd agent && go run ./cmd/agent -server-host 192.168.1.100"
	@echo ""
	@echo "🔨 Сборка:"
	@echo "  make build-all       - Собрать все компоненты"
	@echo ""
	@echo "🧹 Очистка:"
	@echo "  make clean           - Удалить артефакты сборки"
	@echo ""
	@echo "📘 Документация:"
	@echo "  DYNAMIC_CONFIG.md    - Полная инструкция по динамической конфигурации"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Установка зависимостей
install-deps:
	@echo "📦 Установка зависимостей..."
	@echo "→ Agent dependencies"
	cd agent && go mod tidy
	@echo "→ Server dependencies"
	cd backend && go mod tidy
	@echo "→ Frontend dependencies"
	cd frontend && npm install
	@echo "Все зависимости установлены"

# Запуск Backend API Server
run-server:
	@echo "🌐 Запуск Backend API Server на порту 8080..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	cd backend && go run ./cmd/api -port 8080 -data ./data

# Запуск Frontend
run-frontend:
	@echo "💻 Запуск Frontend на порту 5173..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Откройте браузер: http://localhost:5173"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	cd frontend && npm run dev

# Запуск SIEM Agent
run-agent:
	@echo "📡 Запуск SIEM Agent..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	cd agent && go run ./cmd/agent -config ./configs/agent.yaml

# Запуск агента с автоопределением IP сервера
run-agent-auto:
	@echo "📡 Запуск SIEM Agent с автоопределением IP..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@./agent/run-agent-auto.sh

# Запуск агента с кастомными параметрами (пример)
run-agent-custom:
	@echo "📡 Запуск SIEM Agent с кастомными параметрами..."
	@echo "Используйте: make run-agent-with HOST=192.168.1.100 PORT=8080"
	@echo "Или напрямую: SIEM_SERVER_HOST=192.168.1.100 SIEM_SERVER_PORT=8080 make run-agent"

# Запуск агента с параметрами из переменных окружения
run-agent-with:
	@echo "📡 Запуск SIEM Agent..."
	@echo "   Server: $(HOST):$(PORT)"
	cd agent && SIEM_SERVER_HOST=$(HOST) SIEM_SERVER_PORT=$(PORT) go run ./cmd/agent -config ./configs/agent.yaml

# Сборка всех компонентов
build-all:
	@echo "🔨 Сборка всех компонентов..."
	@mkdir -p bin
	cd agent && go build -o ../bin/siem-agent ./cmd/agent
	cd backend && go build -o ../bin/siem-server ./cmd/api
	cd frontend && npm run build
	@echo "Сборка завершена"
	@echo "Бинарники:"
	@ls -lh bin/

# Очистка
clean:
	@echo "🧹 Очистка..."
	rm -rf bin/
	rm -rf agent/.offsets/
	cd frontend && rm -rf dist/ || true
	@echo "Очистка завершена"

# Быстрый запуск сервера (для тестирования)
quick-start:
	@echo "🚀 Быстрый запуск Backend API..."
	cd backend && go run ./cmd/api -port 8080 -data ./data

# Docker команды
docker-build:
	@echo "🐳 Сборка Docker образов..."
	docker-compose build

docker-up:
	@echo "🚀 Запуск всех сервисов в Docker..."
	docker-compose up -d
	@echo ""
	@echo "✅ Сервисы запущены!"
	@echo "🌐 Веб-интерфейс: http://localhost"
	@echo "📡 Backend API: http://localhost/api"
	@echo "📊 Прямой доступ к Backend: http://localhost:8080"
	@echo ""
	@echo "Для просмотра логов: make docker-logs"

docker-down:
	@echo "🛑 Остановка всех сервисов..."
	docker-compose down

docker-logs:
	@echo "📋 Логи сервисов (Ctrl+C для выхода)..."
	docker-compose logs -f

docker-clean:
	@echo "🧹 Удаление всех контейнеров, сетей и volumes..."
	docker-compose down -v
	docker system prune -f

docker-rebuild:
	@echo "🔄 Пересборка и перезапуск..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d
	@echo "✅ Готово!"
