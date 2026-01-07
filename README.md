<div align="center">
  <h1>SIEM web panel</h1>
  <p>Personal Security Information and Event Management system.</p>
</div>

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.24.0-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)](https://reactjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5-3178C6?style=for-the-badge&logo=typescript&logoColor=white)](https://www.typescriptlang.org/)
[![Ubuntu](https://img.shields.io/badge/Ubuntu-E95420?style=for-the-badge&logo=ubuntu&logoColor=white)](https://ubuntu.com/)

</div>

---

> [!IMPORTANT]
> This project is only for personal usage and please do not use it in a production environment without proper security auditing.

## Quick Start

To instantly set up and run the SIEM environment, you can use the provided Makefile commands.

### Prerequisites
- Go 1.22+
- Node.js & npm

### Installation & Run

1.  **Install Dependencies**:
    ```bash
    make install-deps
    ```

2.  **Start the Backend**:
    ```bash
    make run-server
    ```

3.  **Start the Frontend** (in a new terminal):
    ```bash
    make run-frontend
    ```

4.  **Start the Agent** (in a new terminal):
    ```bash
    make run-agent
    ```

## Security & Features

The SIEM Platform prioritizes visibility and security monitoring:

*   **Real-time Event Collection**: Specialized agents collect system logs and security events in real-time.
*   **Centralized Dashboard**: A modern web interface to visualize threats, active agents, and top processes.
*   **Secure Communication**: Agents communicate with the server via secure JSON API.
*   **Rule-Based Analysis**: Events are categorized by severity (Low, Medium, Critical) and type.

## Interface

The system features a modern, intuitive dashboard for monitoring your infrastructure.

*   **Live Event Feed**: Watch security events as they happen.
*   **Threat Visualization**: Pie charts and graphs for severity distribution and event types.
*   **Agent Management**: Monitor the status and activity of connected agents.

## Architecture

The system is built with modularity and performance in mind:

*   **[Backend](backend)**: Powered by Go, providing high-performance API endpoints and data processing.
*   **[Frontend](frontend)**: A responsive single-page application built with React and TypeScript.
*   **[Agent](agent)**: Lightweight Go-based collector running on client machines to gather telemetry.
