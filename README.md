<div align="center">

# ⚡ LineTalk

### *Secure, keyboard-driven terminal chat for the modern developer.*

LineTalk is a lightweight, keyboard-driven terminal application designed for secure 1-to-1 messaging. Built entirely in Go using the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss)), it brings end-to-end encryption and real-time WebSocket communication directly to your command line.

[![Go Report Card](https://goreportcard.com/badge/github.com/DanielJohn17/tui-chat)](https://goreportcard.com/report/github.com/DanielJohn17/tui-chat)
[![License: MIT](https://img.shields.io/badge/License-MIT-cyan.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/DanielJohn17/tui-chat?color=magenta)](https://github.com/DanielJohn17/tui-chat/releases)

---

</div>

## 📸 Interface Preview

<div align="center">

### 💬 Active Conversation & Real-Time Status
![LineTalk Chat Dashboard](assets/chat.png)

<br/>

| 🔐 Authentication Screen | 👤 Profile & Identity Management |
| :---: | :---: |
| ![LineTalk Login Screen](assets/login.png) | ![LineTalk Profile Edit](assets/profile.png) |

</div>

---

## ✨ Key Features

* **🛡️ Zero-Trust Security**: Built with client-side end-to-end encryption so your private 1-to-1 conversations remain completely unreadable to anyone in between.
* **🎨 Streamlined TUI**: A minimalist, distraction-free interface rendered with custom Lipgloss cyberpunk styling that blends seamlessly into any terminal workflow.
* **⚡ Blazing Fast Go Architecture**: Powered by a lean Go API and Bubble Tea model for instant startup times, minimal memory usage, and zero input lag.
* **🎯 Pure 1-to-1 Focus**: Strip away channels, server lists, and notification bloat to focus entirely on direct, line-by-line conversation.
* **🔄 Live WebSockets & Real-Time Presence**: Instant message delivery with automatic heartbeat pings and auto-reconnection.
* **👥 Multi-Session Support**: Test and run multiple isolated accounts simultaneously in split terminal windows (`alice`, `bob`, etc.).

---

## 🏗️ Tech Stack

* **TUI Client**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (ELM architecture), [Lipgloss](https://github.com/charmbracelet/lipgloss) (styling), [Bubbles](https://github.com/charmbracelet/bubbles)
* **Backend API**: [Gin Web Framework](https://github.com/gin-gonic/gin), [Gorilla WebSocket](https://github.com/gorilla/websocket)
* **Database & Persistence**: [PostgreSQL (Neon Serverless)](https://neon.tech), [pgx/v5](https://github.com/jackc/pgx), [sqlc](https://sqlc.dev/)
* **Migrations**: [Goose](https://github.com/pressly/goose) (Embedded inside binary via `embed.FS`)
* **Auth & Security**: JWT Bearer Tokens, Bcrypt Password Hashing

---

## 🚀 Quick Start (Pre-built Binaries)

Download the latest standalone binary for your platform from **[GitHub Releases](https://github.com/DanielJohn17/tui-chat/releases)**:

```bash
# Example for Linux (x86_64)
curl -L -o linetalk https://github.com/DanielJohn17/tui-chat/releases/latest/download/tui-linux-amd64
chmod +x linetalk
./linetalk

# Example for macOS (Apple Silicon)
curl -L -o linetalk https://github.com/DanielJohn17/tui-chat/releases/latest/download/tui-darwin-arm64
chmod +x linetalk
./linetalk
```

*(Windows executables `tui-windows-amd64.exe` are also available in Releases).*

---

## 🛠️ Local Development Setup

### 1. Prerequisites
* **Go 1.22+**
* **PostgreSQL** (Local or Cloud instance like [Neon](https://neon.tech))
* **Make**

### 2. Clone the Repository
```bash
git clone https://github.com/DanielJohn17/tui-chat.git
cd tui-chat
```

### 3. Environment Configuration (`.env`)
Create a `.env` file in the root directory:

```env
# Database URI (Neon PostgreSQL or Local)
DATABASE_URL=postgresql://user:password@ep-sample-pooler.neon.tech/tui_chat_db?sslmode=require

# Or individual DB parameters
DB_HOST=localhost
DB_PORT=5432
DB_NAME=tui_chat_db
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# API Configuration
PORT=8080
GO_ENV=development
JWT_SECRET_KEY=your_development_secret_key_here
JWT_EXP=259200
```

### 4. Database Migrations & Seeding
Migrations are automatically embedded into the binary and executed on API startup. You can also run them manually:

```bash
# Run migrations Up
make migrate

# Populate database with sample users (alice, bob, charlie, etc.)
make seed
```

### 5. Running the Application
Open two terminal splits to run the API server and the TUI client:

```bash
# Split 1: Start Backend API (Port :8080)
make run-api

# Split 2: Launch TUI Client
make run-tui
```

---

## 👥 Multi-Window Session Testing

To test real-time 1-to-1 messaging between two users on the same machine:

```bash
# Terminal Split 1: Logged in as Alice
make dev-user1

# Terminal Split 2: Logged in as Bob
make dev-user2

# Custom isolated session
make run-tui SESSION=charlie
```

---

## ⌨️ Keyboard Shortcuts

| Keybinding | Action |
| :--- | :--- |
| `j` / `k` or `↑` / `↓` | Navigate conversations list |
| `i` / `Enter` | Focus message input box (Typing Mode) |
| `Esc` | Unfocus input box (Command Navigation Mode) |
| `n` | Open New Direct Message modal |
| `p` | Open Profile & Identity Edit screen |
| `?` or `Ctrl+H` / `F1` | Open Interactive Help Modal |
| `q` / `Ctrl+C` | Quit LineTalk |

---

## 📋 Makefile Targets

| Target | Description |
| :--- | :--- |
| `make build` | Builds both `bin/api` and `bin/tui` host binaries |
| `make build-api` | Compiles backend API server binary |
| `make build-tui` | Compiles TUI client binary with embedded `API_URL` |
| `make build-linux` | Cross-compiles Linux TUI binaries (amd64, arm64, 386, arm) |
| `make build-windows` | Cross-compiles Windows TUI executables (amd64, arm64, 386) |
| `make build-all` | Compiles all platform targets into `bin/` |
| `make run-api` | Runs API server directly via `go run` |
| `make run-tui` | Runs TUI client directly via `go run` |
| `make seed` | Populates database with sample users and chat history |
| `make migrate` | Applies goose database migrations |
| `make test` | Runs all unit and integration tests |
| `make sqlc` | Regenerates SQL models and database queries |
| `make fmt` / `make vet` | Code formatting and linting verification |

---

## 🚢 CI/CD & Deployment

LineTalk includes fully automated GitHub Actions workflows:

* **`.github/workflows/ci.yml`**: Automatically verifies dependencies, runs tests, and validates builds on pull requests.
* **`.github/workflows/release.yml`**: Triggered on tag push (`git tag v1.0.1 && git push origin v1.0.1`).
  * Cross-compiles standalone TUI clients for Linux, Windows, and macOS.
  * Publishes a GitHub Release with SHA256 checksums.
  * Auto-deploys the private backend API binary directly to the production server via SSH.

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
