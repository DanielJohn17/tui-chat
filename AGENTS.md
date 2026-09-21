# Project Rules: TUI Chat

## 1. Make-First Workflow
Always use `make` targets for development, building, testing, and formatting:
- **Build**: `make build` (builds both `bin/api` and `bin/tui`), `make build-api`, `make build-tui`
- **Testing & Verification**: `make test`, `make test-api`, `make test-tui`, `make test-cover`
- **Code Quality**: `make fmt` (formats code), `make vet` (runs go vet), `make tidy`
- **Database & Migrations**: `make migrate` (runs goose migrations), `make seed` (populates database), `make sqlc` (generates query models)
- **Execution**: `make run-api` (starts Gin server), `make run-tui` (launches TUI), `make dev-api` (live reload)

## 2. Terminal UI (Bubble Tea & Lipgloss) Standards
- **UTF-8 & Emoji Safety**: Always use `theme.Truncate()` for text truncation in fixed-width containers to count visual cell width and avoid slicing multi-byte runes.
- **Layout Dimensions**: Calculate container widths and heights precisely taking borders (`2 cells`) and padding (`2 cells`) into account so total row width equals terminal width without line wrapping.
- **Focus Discipline**: Do not auto-focus chat inputs on app launch; preserve top-level keyboard navigation (`j/k`, `i`, `n`, `p`, `?`, `q`).
- **Terminal Aesthetics**: Maintain clean cyberpunk dark palette (`theme.ColorMagenta`, `theme.ColorCyan`, `theme.ColorGreen`, `theme.ColorYellow`, `theme.ColorDimText`).
