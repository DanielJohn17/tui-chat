# Go-TUI & GSX Engineering Rules

## 1. Code Generation & Builds
- Always run `go run github.com/grindlemire/go-tui/cmd/tui@v0.19.0 generate ./...` after modifying any `.gsx` file.
- Verify changes with `go build ./...` and `go test ./...` before considering tasks complete.

## 2. Keymap Conflict Prevention
- **NEVER** register duplicate `OnStop` or `On` handlers for the same key or rune within the same component or overlapping tree positions.
- Manage modal and view keybindings centrally in `App.KeyMap()` using state-based `if` branches rather than appending or duplicating handlers in child components.
- Do not let modal dismiss handlers (`Esc`, `q`, `h`, `?`) collide with global hotkeys.

## 3. UI Layout & Text Truncation
- Always truncate dynamic strings (`truncate(s, maxLen)`) inside fixed-width containers (sidebars, metadata badges, list items) to prevent line wrapping and broken columns.
- Factor in border width (2 cells) and horizontal padding (`px-1` = 2 cells) when calculating maximum text widths.

## 4. Focus & Mouse Interaction
- Enable `tui.WithMouse()` on the root application for scroll wheel and click support.
- Do NOT set `autoFocus={true}` on main chat inputs upon startup; preserve global navigation (`j/k`, `q`, `p`, `n`, `h`) immediately on launch.
- Use `autoFocus={true}` only on explicit modal popups (such as `NewDMModal` or editing forms).

## 5. Visual Aesthetics (OpenCode / Posting.sh)
- Avoid noisy nested border boxes around each individual chat message or list item. Use clean streams with color-coded badges (`text-magenta`, `text-cyan`, `text-yellow`, `text-green`).
- Format hotkeys with literal names (`q`, `n`, `p`, `Tab`, `Enter`, `j/k`, `h/?`), never caret prefixes like `^p`.
