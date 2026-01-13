#!/bin/bash

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Конфигурация
REPO_URL="https://raw.githubusercontent.com/Dreamer0iQ/siem-web-panel/main"
INSTALL_DIR="/opt/siem"

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         SIEM System - One-Click Installer                 ║"
echo "║                    Ubuntu VPS Edition                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Проверка root прав
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}❌ Пожалуйста, запустите скрипт с правами root (sudo)${NC}"
    exit 1
fi

echo -e "${YELLOW}📋 Проверка системы...${NC}"

# Проверка Ubuntu
if ! grep -q "Ubuntu" /etc/os-release; then
    echo -e "${RED}❌ Этот скрипт предназначен для Ubuntu${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Ubuntu обнаружен${NC}"

# Установка Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}🐳 Установка Docker...${NC}"
    
    # Обновление пакетов
    apt-get update -qq
    apt-get install -y -qq ca-certificates curl gnupg lsb-release
    
    # Добавление GPG ключа Docker
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    
    # Добавление репозитория Docker
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
    
    # Установка Docker
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    
    # Запуск Docker
    systemctl start docker
    systemctl enable docker
    
    echo -e "${GREEN}✅ Docker установлен${NC}"
else
    echo -e "${GREEN}✅ Docker уже установлен${NC}"
fi

# Проверка Docker Compose
if ! docker compose version &> /dev/null; then
    echo -e "${RED}❌ Docker Compose не найден${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Docker Compose доступен${NC}"

# Создание директории для установки
echo -e "${YELLOW}📁 Создание директории установки...${NC}"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# Скачивание конфигурационных файлов
echo -e "${YELLOW}⬇️  Скачивание конфигурации...${NC}"

echo -e "${BLUE}   → docker-compose.yml${NC}"
curl -fsSL "$REPO_URL/docker-compose.production.yml" -o docker-compose.yml

echo -e "${BLUE}   → nginx.conf${NC}"
curl -fsSL "$REPO_URL/nginx.conf" -o nginx.conf

echo -e "${GREEN}✅ Конфигурация скачана${NC}"

# Авторизация в GitHub Container Registry (публичные образы не требуют авторизации)
echo -e "${YELLOW}🔐 Подготовка к скачиванию образов...${NC}"

# Скачивание образов
echo -e "${YELLOW}📦 Скачивание Docker образов (это может занять несколько минут)...${NC}"
docker compose pull

# Запуск сервисов
echo -e "${YELLOW}🚀 Запуск SIEM системы...${NC}"
docker compose up -d

# Ожидание запуска
echo -e "${YELLOW}⏳ Ожидание запуска сервисов...${NC}"
sleep 10

# Проверка статуса
echo -e "${YELLOW}📊 Проверка статуса сервисов...${NC}"
docker compose ps

# Получение IP адреса
SERVER_IP=$(hostname -I | awk '{print $1}')

echo ""
echo -e "${GREEN}╔════════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║                  ✅ Установка завершена!                   ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}📍 Доступ к системе:${NC}"
echo -e "   🌐 Веб-интерфейс: ${GREEN}http://$SERVER_IP${NC}"
echo -e "   📡 API:           ${GREEN}http://$SERVER_IP/api${NC}"
echo -e "   🔧 Backend:       ${GREEN}http://$SERVER_IP:8080${NC}"
echo ""
echo -e "${BLUE}📂 Директория установки: ${GREEN}$INSTALL_DIR${NC}"
echo ""
echo -e "${YELLOW}📝 Полезные команды:${NC}"
echo -e "   Просмотр логов:     ${GREEN}cd $INSTALL_DIR && docker compose logs -f${NC}"
echo -e "   Перезапуск:         ${GREEN}cd $INSTALL_DIR && docker compose restart${NC}"
echo -e "   Остановка:          ${GREEN}cd $INSTALL_DIR && docker compose down${NC}"
echo -e "   Обновление:         ${GREEN}cd $INSTALL_DIR && docker compose pull && docker compose up -d${NC}"
echo ""
echo -e "${BLUE}🔍 Проверка работы агента:${NC}"
echo -e "   ${GREEN}docker logs siem-agent${NC}"
echo ""
echo -e "${YELLOW}⚠️  Примечание: Агент мониторит логи хоста из /var/log${NC}"
echo ""
