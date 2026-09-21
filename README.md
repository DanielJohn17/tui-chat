<div align="center">

# ⚡ LineTalk

### *Secure, keyboard-driven terminal chat for the modern developer.*

LineTalk is a lightweight, keyboard-driven terminal application engineered for private, zero-distraction 1-to-1 communication. Built entirely in Go using the Charm ecosystem ([Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lipgloss](https://github.com/charmbracelet/lipgloss)), it brings client-side security and real-time WebSocket communication directly to your command line.

<br/>

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev)
[![Bubble Tea](https://img.shields.io/badge/TUI-Bubble%20Tea-F25D94?style=for-the-badge&logo=terminal&logoColor=white)](https://github.com/charmbracelet/bubbletea)
[![PostgreSQL](https://img.shields.io/badge/DB-Neon%20Postgres-00E599?style=for-the-badge&logo=postgresql&logoColor=white)](https://neon.tech)
[![WebSocket](https://img.shields.io/badge/RealTime-WebSocket-00F5FF?style=for-the-badge&logo=socketdotio&logoColor=black)](https://github.com/gorilla/websocket)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/DanielJohn17/tui-chat?style=for-the-badge&color=magenta)](https://github.com/DanielJohn17/tui-chat/releases)

<br/>

[**Explore Features**](#-key-features) •
[**Quick Start**](#-quick-start-pre-built-binaries) •
[**Local Development**](#-local-development-setup) •
[**Multi-Session Testing**](#-multi-window-session-testing) •
[**Keybindings**](#-keyboard-shortcuts) •
[**Makefile**](#-makefile-reference)

---

</div>

## 📸 Interface Preview

<div align="center">

### 💬 Active Conversation & Real-Time Presence
*Real-time direct messaging with instant WebSocket synchronization, unread counters, and live status bar.*

![LineTalk Chat Dashboard](assets/chat.png)

<br/>

| 🔐 Authentication & Session Switcher | 👤 Identity & Profile Management |
| :---: | :---: |
| ![LineTalk Login Screen](assets/login.png) | ![LineTalk Profile Edit](assets/profile.png) |
| *Fast keyboard navigation with persistent login* | *Custom status bio, display name, and active connection info* |

</div>

---

## ✨ Key Features

* **🛡️ Zero-Trust Security**: Built with client-side privacy in mind so your private 1-to-1 conversations remain strictly confidential.
* **🎨 Cyberpunk Minimalist TUI**: Rendered with custom Lipgloss styling (`#00F5FF`, `#FF007F`, `#00FF66`) designed to seamlessly integrate into any dark terminal workflow.
* **⚡ Blazing Fast Go Architecture**: Powered by a lean Go API engine and Bubble Tea model for instant startup times (<10ms), minimal RAM footprint (<15MB), and zero input lag.
* **🎯 Pure 1-to-1 Focus**: Strip away channel clutter, unread notification bloat, and server hierarchies to focus purely on direct, line-by-line conversation.
* **🔄 Live WebSockets & Heartbeat Resilience**: Instant bidirectional message dispatch with automated ping/pong keep-alives and seamless auto-reconnect.
* **👥 Multi-Window Session Sync**: Test and operate multiple isolated identities (`alice`, `bob`, `charlie`) simultaneously in split terminal panes.

---

## 🏛️ System Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                                CLIENT WORKSPACE                                  │
│                                                                                  │
│   ┌────────────────────────────────┐            ┌────────────────────────────┐   │
│   │    Terminal Split 1 (Alice)    │            │  Terminal Split 2 (Bob)    │   │
│   │  [ Bubble Tea + Lipgloss TUI ] │            │ [ Bubble Tea + Lipgloss ]  │   │
│   └───────────────┬────────────────┘            └─────────────┬──────────────┘   │
└───────────────────┼───────────────────────────────────────────┼──────────────────┘
                    │ HTTPS / WSS                               │ HTTPS / WSS
                    ▼                                           ▼
┌──────────────────────────────────────────────────────────────────────────────────┐
│                       PRODUCTION BACKEND INFRASTRUCTURE                          │
│                                                                                  │
│   [ LiteSpeed / Apache Reverse Proxy (.htaccess) ]                               │
│       ├── Routes REST API traffic  ──► http://127.0.0.1:8080                     │
│       └── Upgrades WebSockets      ──► ws://127.0.0.1:8080/api/v1/ws             │
│                                                                                  │
│   [ Go API Engine (Gin + Goroutines + WS Hub) ]                                  │
│       ├── JWT Authentication & Session Resolver                                  │
│       ├── Goose Migrations (Embedded via embed.FS)                               │
│       └── sqlc Type-Safe Query Execution                                         │
│                                                                                  │
│   [ Neon Serverless PostgreSQL Database ] (TLS 1.3 Encrypted Connection)        │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Quick Start (Pre-built Binaries)

Zero dependencies required. Download and run the standalone executable for your operating system directly from **[GitHub Releases](https://github.com/DanielJohn17/tui-chat/releases)**:

### Linux (x86_64 / ARM64)
```bash
curl -L -o linetalk https://github.com/DanielJohn17/tui-chat/releases/latest/download/linetalk-linux-amd64
chmod +x linetalk
./linetalk
```

### macOS (Apple Silicon / Intel)
```bash
curl -L -o linetalk https://github.com/DanielJohn17/tui-chat/releases/latest/download/linetalk-darwin-arm64
chmod +x linetalk
./linetalk
```

### Windows (PowerShell / Command Prompt)
Download **`linetalk-windows-amd64.exe`** from [Releases](https://github.com/DanielJohn17/tui-chat/releases) and launch:
```powershell
.\linetalk-windows-amd64.exe
```

> [!TIP]
> Pre-built binaries are pre-configured with the production API endpoint out-of-the-box. No setup or `.env` configuration is needed for general users!

---

## 🛠️ Local Development Setup

### 1. Prerequisites
* **Go 1.22+**
* **PostgreSQL** (Local installation or free cloud instance on [Neon](https://neon.tech))
* **Make**

### 2. Clone Repository
```bash
git clone https://github.com/DanielJohn17/tui-chat.git
cd tui-chat
```

### 3. Environment Configuration (`.env`)
Create a `.env` file in the project root:

```env
# Neon PostgreSQL Connection URI (or local database)
DATABASE_URL=postgresql://user:password@ep-sample-pooler.neon.tech/tui_chat_db?sslmode=require

# Server & Auth Configuration
PORT=8080
GO_ENV=development
JWT_SECRET_KEY=your_development_secret_key_here
JWT_EXP=259200
```

### 4. Database Setup & Automated Migrations
Migrations are embedded inside the Go executable and applied automatically on API startup. You can also manage them via `make`:

```bash
# Apply pending schema migrations
make migrate

# Populate database with sample mock users (alice, bob, charlie, etc.) and chat history
make seed
```

### 5. Launch Development Servers
Open two terminal splits to run the API and TUI client:

```bash
# Pane 1: Start Backend API (Listening on :8080)
make run-api

# Pane 2: Launch TUI Client
make run-tui
```

---

## 👥 Multi-Window Session Testing

LineTalk includes isolated session profile switches for seamless real-time testing across multiple accounts on the same machine:

```bash
# Terminal Split 1: Launch session for Alice
make dev-user1

# Terminal Split 2: Launch session for Bob
make dev-user2

# Launch custom isolated session profile
make run-tui SESSION=charlie
```

> [!NOTE]
> Session tokens and identity states are stored locally per-profile in isolated session cache files so switching users never overwrites your default account.

---

## ⌨️ Keyboard Shortcuts

LineTalk is designed for fluid, 100% keyboard-driven interaction:

### Navigation & Chat Controls
| Keybinding | Scope | Action |
| :--- | :--- | :--- |
| <kbd>j</kbd> / <kbd>k</kbd> or <kbd>↑</kbd> / <kbd>↓</kbd> | Global | Navigate conversation list up / down |
| <kbd>i</kbd> / <kbd>Enter</kbd> | Global | Focus message input box (**Typing Mode**) |
| <kbd>Esc</kbd> | Typing Mode | Unfocus message input box (**Navigation Mode**) |
| <kbd>Enter</kbd> | Typing Mode | Send message |
| <kbd>n</kbd> | Navigation | Open **New Direct Message** modal |
| <kbd>p</kbd> | Navigation | Open **Profile & Identity** screen |
| <kbd>?</kbd> / <kbd>Ctrl+H</kbd> / <kbd>F1</kbd> | Global | Open **Interactive Keybinding Help** |
| <kbd>q</kbd> / <kbd>Ctrl+C</kbd> | Global | Quit LineTalk cleanly |

---

## 📋 Makefile Reference

| Target | Category | Description |
| :--- | :--- | :--- |
| `make build` | **Build** | Compiles host binaries for both API and TUI into `bin/` |
| `make build-api` | **Build** | Compiles standalone backend API server binary |
| `make build-tui` | **Build** | Compiles host TUI client binary with embedded `API_URL` |
| `make build-linux` | **Build** | Cross-compiles Linux TUI binaries (`amd64`, `arm64`, `386`, `arm`) |
| `make build-windows` | **Build** | Cross-compiles Windows executables (`amd64`, `arm64`, `386`) |
| `make build-all` | **Build** | Compiles all platform release targets into `bin/` |
| `make run-api` | **Execute** | Starts Go Gin API server directly |
| `make run-tui` | **Execute** | Launches TUI client with default development profile |
| `make dev-user1` | **Execute** | Launches TUI authenticated as sample user `alice` |
| `make dev-user2` | **Execute** | Launches TUI authenticated as sample user `bob` |
| `make seed` | **Database** | Populates database with sample users, conversations, and chats |
| `make migrate` | **Database** | Applies goose database migrations |
| `make migrate-down` | **Database** | Rolls back the latest database migration |
| `make test` | **Quality** | Runs all unit and integration test suites |
| `make sqlc` | **Generate** | Regenerates type-safe database query models with sqlc |
| `make fmt` / `make vet` | **Quality** | Formats code and executes `go vet` static analysis |

---

## 🚢 CI/CD & Automated Deployment

LineTalk uses a dual-pipeline GitHub Actions workflow:

* **Continuous Integration (`ci.yml`)**: Runs on every pull request to verify dependency checksums, validate test suites, and ensure clean multi-platform compilation.
* **Continuous Delivery (`release.yml`)**: Triggered automatically on version tags (`git tag v1.0.1 && git push origin v1.0.1`):
  1. Compiles standalone TUI client binaries for Linux, macOS, and Windows.
  2. Generates SHA256 checksums and publishes a **GitHub Release**.
  3. Securely deploys the private server binary to production cPanel hosting via SCP/SSH and executes a zero-downtime background restart.

---

## 📄 License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more information.

<div align="center">

*Built with ❤️ in Go and Bubble Tea.*

</div>
