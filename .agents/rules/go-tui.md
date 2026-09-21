# TUI Chat Engineering Rules

## 1. Make-First Commands
Always use the project `Makefile` targets:
- **Build**: `make build` (builds API and TUI binaries), `make build-api`, `make build-tui`
- **Testing & Quality**: `make fmt`, `make vet`, `make test`, `make test-api`, `make test-tui`
- **Database & Migrations**: `make migrate` (goose migrations), `make seed` (seed data), `make sqlc` (generate queries)
- **Running**: `make run-api`, `make run-tui`, `make dev-api`

## 2. Terminal UI (Bubble Tea & Lipgloss) Rules
- **UTF-8 & Emoji Safety**: Always use `theme.Truncate()` for text truncation in fixed-width containers to measure visual cell width and preserve multi-byte UTF-8 runes and emojis.
- **Layout Precision**: When configuring Lipgloss styles with `.Border()` and `.Padding(0, 1)`, set `Width(totalWidth - 2)` and `Height(totalHeight - 2)` so total rendered dimensions match terminal geometry without line wrapping.
- **Focus & Navigation**: Preserve top-level keyboard navigation on startup (`j/k`, `i`, `n`, `p`, `?`, `q`). Do not auto-focus chat inputs on app launch.
- **Visual Aesthetics**: Clean cyberpunk dark palette with minimal borders, avoiding nested box clutter. Hotkeys should use literal labels (`[Esc]`, `[Enter]`, `j/k`).
