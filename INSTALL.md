# 🚀 SIEM System - Quick Installation Guide

## One-Line Installation (Ubuntu VPS)

Install the complete SIEM system with a single command:

```bash
bash <(curl -Ls https://raw.githubusercontent.com/Dreamer0iQ/siem-web-panel/main/install.sh)
```

### What it does:
- ✅ Checks system requirements
- ✅ Installs Docker and Docker Compose (if not present)
- ✅ Downloads configuration files
- ✅ Pulls pre-built Docker images from GitHub Container Registry
- ✅ Starts all services (Backend, Frontend, Nginx, Agent)
- ✅ Configures agent to monitor host logs

### Requirements:
- Ubuntu 20.04+ (VPS or dedicated server)
- Root access (sudo)
- Minimum 2GB RAM
- 5GB free disk space

---

## Access After Installation

After successful installation, you can access:

- **Web Interface**: `http://YOUR_SERVER_IP`
- **API**: `http://YOUR_SERVER_IP/api`
- **Direct Backend**: `http://YOUR_SERVER_IP:8080`

---

## Useful Commands

All commands should be run from `/opt/siem` directory:

```bash
cd /opt/siem

# View logs
docker compose logs -f

# View specific service logs
docker logs siem-backend
docker logs siem-agent
docker logs siem-nginx

# Restart services
docker compose restart

# Stop services
docker compose down

# Update to latest version
docker compose pull
docker compose up -d

# Check service status
docker compose ps
```

---

## Manual Installation (Development)

If you want to build from source:

1. Clone the repository:
```bash
git clone https://github.com/Dreamer0iQ/siem-web-panel.git
cd siem-web-panel
```

2. Build and start:
```bash
make docker-build
make docker-up
```

---

## Architecture

```
┌─────────────┐
│   Browser   │
└──────┬──────┘
       │
       ↓
┌─────────────────┐
│  Nginx (Proxy)  │  :80
└────┬───────┬────┘
     │       │
     ↓       ↓
┌─────────┐ ┌──────────────────┐
│Frontend │ │ Backend (СУБД)   │  :8080
│ (React) │ │ File Storage     │
└─────────┘ └────┬─────────────┘
                 │
                 ↓
              ┌──────┐
              │ Agent│ (monitors /var/log)
              └──────┘
```

---

## Components

- **Backend**: Go-based API server with file-based storage
- **Frontend**: React + Vite web interface
- **Nginx**: Reverse proxy with WebSocket support
- **Agent**: Log collector that monitors host system

---

## Troubleshooting

### Services not starting
```bash
cd /opt/siem
docker compose logs
```

### Agent not collecting logs
Check if agent has access to host logs:
```bash
docker exec siem-agent ls -la /host/logs
```

### Port already in use
Stop conflicting services:
```bash
sudo lsof -ti:80 | xargs sudo kill
sudo lsof -ti:8080 | xargs sudo kill
```

---

## Uninstallation

To completely remove the SIEM system:

```bash
cd /opt/siem
docker compose down -v
cd /
rm -rf /opt/siem
```

---

## GitHub Actions

The repository includes automated Docker image building:
- Triggers on every push to `main` branch
- Builds backend, frontend, and agent images
- Pushes to GitHub Container Registry (ghcr.io)
- Images are publicly available

---

## Support

For issues and questions, please open an issue on GitHub.
