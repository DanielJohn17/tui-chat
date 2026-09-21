---
name: go-tui
description: Best practices, architecture patterns, GSX templating, keymap management, and troubleshooting for building Terminal User Interfaces in Go with github.com/grindlemire/go-tui.
---

# Go-TUI & GSX Engineering Skill

This skill provides patterns, architecture guidelines, best practices, and troubleshooting runbooks for building rich Terminal User Interfaces (TUIs) in Go using `github.com/grindlemire/go-tui` and its GSX templating engine.

---

## 1. Code Generation & Build Workflow

GSX files (`*.gsx`) compile into standard Go source files (`*_gsx.go`).

### Generator Command
Always run code generation before building or running tests:
```bash
go run github.com/grindlemire/go-tui/cmd/tui@v0.19.0 generate ./...
```

### Verification Pipeline
```bash
go run github.com/grindlemire/go-tui/cmd/tui@v0.19.0 generate ./... && go build ./... && go test ./...
```

---

## 2. Keymap Architecture & Dispatch Rules

`go-tui` constructs an internal dispatch table from the component tree on each render cycle. Improper key bindings will cause runtime dispatch errors that corrupt the terminal display.

### Critical Rule: No Duplicate Key Handlers
> **Never** register duplicate `OnStop` or `On` handlers for the same key or rune pattern at the same component or position in the tree.

If duplicates are registered, `go-tui` outputs:
```text
tui: dispatch table error: conflicting stop handlers for key pattern {...} at tree positions X and Y
```
This error writes directly to stderr while the raw terminal mode is active, causing cursor displacement, line wrapping, and ghosting artifacts.

### Pattern: Centralized Modal & View Keymaps
Manage keymaps at the top-level container component using mutually exclusive branches based on active state:

```go
func (a *app) KeyMap() tui.KeyMap {
    // 1. Help Modal Active - capture modal-specific keys exclusively
    if a.showHelp.Get() {
        return tui.KeyMap{
            tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
            tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
            tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
            tui.OnStop(tui.Rune('?'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
            tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
        }
    }

    // 2. Input/Action Modal Active - allow inputs while handling cancel
    if a.showNewDM.Get() {
        return tui.KeyMap{
            tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
            tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.cancelNewDM() }),
        }
    }

    // 3. Sub-View Active (e.g., Profile View)
    if a.view.Get() == viewProfile {
        return tui.KeyMap{
            tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
            tui.OnStop(tui.Rune('c'), func(ke tui.KeyEvent) { a.view.Set(viewChats) }),
            tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.view.Set(viewChats) }),
            tui.OnStop(tui.Rune('e'), func(ke tui.KeyEvent) { a.profileEdit.Set(true) }),
            tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
        }
    }

    // 4. Base View (Navigation & Global Hotkeys)
    return tui.KeyMap{
        tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnStop(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
        tui.OnStop(tui.KeyTab.Shift(), func(ke tui.KeyEvent) { ke.App().FocusPrev() }),
        tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
        tui.OnStop(tui.Rune('n'), func(ke tui.KeyEvent) { a.showNewDM.Set(true) }),
        tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) { a.view.Set(viewProfile) }),
        tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
        tui.OnStop(tui.Rune('?'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
        tui.On(tui.KeyUp, moveUp),
        tui.On(tui.Rune('k'), moveUp),
        tui.On(tui.KeyDown, moveDown),
        tui.On(tui.Rune('j'), moveDown),
    }
}
```

---

## 3. Layout, Bounded Columns & Text Truncation

Fixed-width containers (such as sidebars or metadata columns) will overflow or wrap lines if string lengths exceed the inner width, breaking alignment.

### Inner Width Calculation
$$\text{Inner Width} = \text{Total Width} - (\text{Border Width: 2}) - (\text{Horizontal Padding: } 2 \times \text{px})$$
For `width={30}`, `border-rounded`, `px-1`: $\text{Inner Width} = 30 - 2 - 2 = 26 \text{ cells}$.

### Truncation Helper
Always truncate dynamic strings (names, usernames, preview messages) in bounded views:
```go
func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    if maxLen <= 3 {
        return s[:maxLen]
    }
    return s[:maxLen-3] + "…"
}
```

---

## 4. Focus Management & Mouse Support

### Enabling Mouse Interactions
In application initialization (`app.go`), enable mouse support for wheel scrolling and click-to-focus:
```go
app := tui.NewApp(views.App(c), tui.WithMouse())
```

### Initial Focus Strategy
- **Do not** add `autoFocus={true}` to chat message inputs on initial load. This allows global navigation keys (`j`, `k`, `n`, `p`, `h`, `q`) to work immediately without requiring the user to press Escape.
- Add `autoFocus={true}` **only** in transient modal popups (e.g., `NewDMModal`) where typing is the immediate and sole expected user action.

---

## 5. UI Aesthetics & Styling Guidelines

Follow the OpenCode / Posting.sh terminal design philosophy:
- **Clean Message Streams**: Avoid nested border boxes around each message. Use open message streams with colored identity badges and subtle horizontal dividers (`<hr />`).
- **Curated Color Tokens**:
  - `text-magenta` / `scrollbar-magenta`: Primary accents, commands, and selected indicators (`▶`).
  - `text-cyan`: User identities, tags, version indicators.
  - `text-green`: Online status indicator (`●`), success states, confirmed actions.
  - `text-yellow`: Alerts, system notifications (`◆ SYSTEM`), help hotkey prompts.
  - `font-dim`: Secondary metadata, timestamps, offline status (`○`), inactive shortcuts.
- **Literal Key Labels**: In status bars and help popups, use literal keys (`q`, `n`, `p`, `Tab`, `Enter`, `j/k`, `h/?`) instead of caret symbols like `^p`.
