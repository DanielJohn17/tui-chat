---
name: go-tui
description: Best practices, architecture patterns, layout math, keymap management, styling tokens, and testing runbooks for building rich Terminal User Interfaces (TUIs) in Go with Charm (Bubble Tea, Lipgloss, and Bubbles).
---

# Go TUI Engineering Skill (Charm: Bubble Tea, Lipgloss & Bubbles)

This skill provides patterns, architecture guidelines, layout mathematics, keymap handling, and testing runbooks for building rich, high-performance Terminal User Interfaces (TUIs) in Go using the **Charm ecosystem** (`github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, and `github.com/charmbracelet/bubbles`).

---

## 1. The Elm Architecture & Model Decomposition

Bubble Tea implements the **Elm Architecture** (Model-View-Update). Scale applications cleanly by decomposing large state machines into modular, focused sub-models.

```
                  ┌──────────────────────┐
                  │    AppModel (Root)   │
                  │   Init/Update/View   │
                  └──────────┬───────────┘
                             │
       ┌─────────────────────┼─────────────────────┐
       ▼                     ▼                     ▼
┌──────────────┐      ┌──────────────┐      ┌──────────────┐
│  AuthView    │      │ SidebarView  │      │   ChatView   │
│ (Login/Reg)  │      │ (Chat List)  │      │ (Msg Stream) │
└──────────────┘      └──────────────┘      └──────────────┘
```

### Core Interface Contract
Every interactive component or sub-view should implement the standard Bubble Tea lifecycle:
- `Init() tea.Cmd`: Returns initial batch commands (e.g. timers, cursor blinks, data fetches).
- `Update(tea.Msg) (tea.Model, tea.Cmd)`: Handles incoming messages (key events, window resizing, async payloads) and returns state mutations and next commands.
- `View() string`: Pure rendering function returning styled text for terminal output.

### Asynchronous Event Channels & Command Dispatch
Never block the `Update` loop with network calls or locks. Bridge background events (WebSocket, HTTP, timers) into Bubble Tea commands:

```go
// Bridge an external channel (e.g. WebSocket events) into Bubble Tea message loop
func (m AppModel) listenWSCmd() tea.Cmd {
    return func() tea.Msg {
        if m.wsEventsChan == nil {
            return nil
        }
        evt, ok := <-m.wsEventsChan
        if !ok {
            return nil
        }
        return WSIncomingMsg{Event: evt}
    }
}
```

---

## 2. Terminal Dimensions & Lipgloss Layout Math

Terminal layouts must adapt dynamically to terminal resizing without text wrapping or layout overflow.

### Precise Box Dimension Calculations
Lipgloss border and padding tokens consume visual character cells:
$$\text{Inner Content Width} = \text{Box Width} - (\text{Border Width: } 2) - (\text{Horizontal Padding: } 2 \times \text{PaddingX})$$
$$\text{Inner Content Height} = \text{Box Height} - (\text{Border Height: } 2) - (\text{Vertical Padding: } 2 \times \text{PaddingY})$$

**Example**: For a container with `Width(38)`, `Border(lipgloss.RoundedBorder())` (`2`), `Padding(0, 1)` (`2`):
$$\text{Inner Content Width} = 38 - 2 - 2 = 34 \text{ cells}$$

### Total Row Alignment (Zero-Wrapping Rule)
The sum of horizontal columns joined with `lipgloss.JoinHorizontal(lipgloss.Top, ...)` must strictly match the terminal width:
$$\text{Sidebar Width} + \text{Chat View Width} = \text{Terminal Width}$$
If the total width exceeds `msg.Width`, the terminal will automatically soft-wrap lines, causing vertical tearing and cursor displacement.

### Multi-byte UTF-8, CJK & Emoji Safety
> **Critical**: Never slice raw Go strings (`s[:maxLen]`) or count bytes (`len(s)`) for layout truncation. Multi-byte runes and emojis will be sliced mid-byte or consume 2 terminal cells, corrupting output.

Always use visual cell width truncation (via `golang.org/x/term` or `runewidth`):
```go
func Truncate(s string, maxCells int) string {
    if runewidth.StringWidth(s) <= maxCells {
        return s
    }
    if maxCells <= 1 {
        return "…"
    }
    var b strings.Builder
    currentW := 0
    for _, r := range s {
        rw := runewidth.RuneWidth(r)
        if currentW+rw+1 > maxCells {
            b.WriteRune('…')
            break
        }
        b.WriteRune(r)
        currentW += rw
    }
    return b.String()
}
```

---

## 3. Keybinding Architecture & Focus Discipline

### Hotkey Precedence Hierarchy
Handle key events in strict priority order inside `Update(msg tea.Msg)`:
1. **Modal Overlays**: If a modal is visible (`ModalHelp`, `ModalNewDM`), intercept keys (`Esc`, `Enter`, `?`, `q`) and block underlying views.
2. **Global Navigation Overrides**: Priority key combinations like `Ctrl+H` / `F1` (Help) or `Ctrl+C` (Quit) should work globally even when typing in text inputs.
3. **Input Mode vs Navigation Mode**:
   - **Input Mode** (`i` to enter, `Esc` to exit): Routes keystrokes directly to text area / text inputs.
   - **Navigation Mode** (`j`/`k` up/down, `n` new DM, `p` profile, `r` retry, `?` help, `q` quit): Top-level navigation shortcuts.

### Focus Discipline Rule
- **Never** auto-focus text inputs on application initialization. This preserves instant keyboard navigation (`j/k`, `n`, `p`, `?`, `q`) on launch without forcing users to press `Escape`.
- Auto-focus text inputs **only** inside dedicated interactive modal dialogs (such as `NewDMModal`) where typing is the immediate and sole expected user action.

---

## 4. Mouse Support & Click Targets

### Enabling Mouse Motion
Initialize Bubble Tea with mouse tracking:
```go
p := tea.NewProgram(initialModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
```

### Click Target Detection
Calculate item bounding boxes during `View()` or store line index ranges (`StartY`, `EndY`) to dispatch clicks reliably:
```go
type ClickableItem struct {
    ChatIndex int
    StartY    int
    EndY      int
}

// In mouse handler:
if mouseMsg.Action == tea.MouseActionPress && mouseMsg.Button == tea.MouseButtonLeft {
    for _, item := range m.clickableItems {
        if mouseMsg.Y >= item.StartY && mouseMsg.Y <= item.EndY && mouseMsg.X < sidebarWidth {
            m.SelectIndex(item.ChatIndex)
            return m, nil
        }
    }
}
```

---

## 5. UI Aesthetics & Cyberpunk Color Palette

Maintain a high-contrast dark cyberpunk palette (`lipgloss.Color`):

| Token | Hex / Color | Semantic Role |
|---|---|---|
| `ColorMagenta` | `#FF007F` / ANSI 201 | Primary brand accent, commands, selection indicators (`▶`), active tab highlight |
| `ColorCyan` | `#00FFFF` / ANSI 51 | User handles (`@user`), sender tags, room headings, active titles |
| `ColorGreen` | `#00FF87` / ANSI 48 | Online status (`●`), unread badge pill backgrounds, success states |
| `ColorYellow` | `#FFDF00` / ANSI 220 | Alerts, notification banners (`🔔`), reconnecting countdowns, spinner frames |
| `ColorRed` | `#FF5F87` / ANSI 203 | Errors, failed connections, offline indicators (`○`) |
| `ColorDimText` | `#585858` / ANSI 240 | Timestamps, inactive shortcut hints, subtle borders, secondary metadata |

---

## 6. Unit Testing Bubble Tea Models

Bubble Tea models are deterministic state machines that can be tested directly without terminal emulators:

```go
func TestChatInteraction(t *testing.T) {
    client := newTestClient()
    app := tui.NewApp(client)

    // 1. Send Window Resize
    model, _ := app.Update(tea.WindowSizeMsg{Width: 120, Height: 35})

    // 2. Simulate Keypresses
    model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

    // 3. Assert on Rendered View Output
    view := model.View()
    assert.Contains(t, view, "@bob")
    assert.Contains(t, view, "LIVE")
}
```

---

## 7. Make-First Verification Workflow

Always verify all TUI changes using standard `make` targets:
- **Format Code**: `make fmt`
- **Lint & Vet**: `make vet`
- **Run Tests**: `make test` (or `make test-tui`)
- **Build Binaries**: `make build` (or `make build-tui`)
