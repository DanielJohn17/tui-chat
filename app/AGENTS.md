# Project Rules: TUI Chat

## Go-TUI & GSX Standards
- **Generation & Build**: After any edit to `.gsx` files, always run:
  `go run github.com/grindlemire/go-tui/cmd/tui@v0.19.0 generate ./... && go build ./...`
- **Keymap Dispatch Safety**: Never register duplicate `OnStop` or `On` key patterns on the same component or conflicting tree paths. Use centralized, mutually exclusive `KeyMap()` routing in `app.gsx` for modals and sub-views.
- **Text Truncation**: Always apply `truncate()` to dynamic text fields in fixed-width containers (e.g. sidebar items) to avoid line wrapping and visual overlap.
- **Focus Discipline**: Do not auto-focus input fields on app launch; preserve top-level keyboard navigation (`j/k`, `n`, `p`, `h`, `q`) on startup.
- **Terminal Aesthetics**: Maintain OpenCode / Posting.sh terminal styling with rich cyberpunk color accents (`text-magenta`, `text-cyan`, `text-yellow`, `text-green`) and minimal box border clutter.
