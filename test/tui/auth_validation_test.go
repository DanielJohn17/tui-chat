package test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/internal/tui"
	"github.com/DanielJohn17/tui-chat/internal/tui/theme"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/auth"
	"github.com/DanielJohn17/tui-chat/internal/tui/views/statusbar"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestRegisterPasswordChecklistHints(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 35})

	// Switch to Register mode via Ctrl+T
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	view := model.View()

	// Initial view in Register mode has all checks unmet
	assert.Contains(t, view, "[ ] Uppercase (A-Z)")
	assert.Contains(t, view, "[ ] Number (0-9)")
	assert.Contains(t, view, "[ ] Minimum 6 characters")

	// Focus password field (focusIndex 2: DisplayName=0, Username=1, Password=2)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})

	// 1. Type lowercase only "abc"
	for _, r := range "abc" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	view = model.View()
	assert.Contains(t, view, "[ ] Uppercase (A-Z)")
	assert.Contains(t, view, "[ ] Number (0-9)")
	assert.Contains(t, view, "[ ] Minimum 6 characters")

	// 2. Type an uppercase letter 'D' -> "abcD"
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	view = model.View()
	assert.Contains(t, view, "[✔] Uppercase (A-Z)", "Uppercase requirement should be met")
	assert.Contains(t, view, "[ ] Number (0-9)")
	assert.Contains(t, view, "[ ] Minimum 6 characters")

	// 3. Type a number '9' -> "abcD9"
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}})
	view = model.View()
	assert.Contains(t, view, "[✔] Uppercase (A-Z)")
	assert.Contains(t, view, "[✔] Number (0-9)", "Number requirement should be met")
	assert.Contains(t, view, "[ ] Minimum 6 characters")

	// 4. Type one more char 'x' -> "abcD9x" (len 6)
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	view = model.View()
	assert.Contains(t, view, "[✔] Uppercase (A-Z)")
	assert.Contains(t, view, "[✔] Number (0-9)")
	assert.Contains(t, view, "[✔] Minimum 6 characters", "Length requirement should now be met")
}

func TestAuthValidationErrors(t *testing.T) {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 35})

	// --- LOGIN VALIDATION ---
	// 1. Empty username
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter your username")

	// 2. Username < 3 chars
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a', 'b'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Username must be at least 3 characters")

	// 3. Enter valid username, empty password
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}) // "abc"
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter your password")

	// --- REGISTER VALIDATION ---
	// Switch to Register mode
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlT})

	// 4. Empty display name
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter your display name")

	// 5. Fill name, empty username
	for _, r := range "Alice Smith" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // to username
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter a username")

	// 6. Fill short username
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Username must be at least 3 characters")

	// 7. Fill valid username, empty password
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y', 'z'}})
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyTab}) // to password
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Please enter a password")

	// 8. Fill password with no uppercase
	for _, r := range "password123" {
		model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Contains(t, model.View(), "Password must contain at least one capital")
}

func TestCleanAuthError(t *testing.T) {
	// Backend validator error -> real error
	rawValErr := errors.New("validation error: Key: 'RegisterUserType.Password' Error:Field validation for 'Password' failed on the 'containsany' tag")
	assert.Equal(t, "Password must contain at least one uppercase letter (A-Z) and one number (0-9).", auth.CleanAuthError(rawValErr))

	// Backend user already exists -> real error
	assert.Equal(t, "Username is already taken. Please choose another username.", auth.CleanAuthError(errors.New("user already exists")))

	// Backend incorrect credentials -> real error
	assert.Equal(t, "Incorrect username or password. Please try again.", auth.CleanAuthError(errors.New("incorrect username or password")))

	// Min length validator error
	rawMinErr := errors.New("validation error: Key: 'RegisterUserType.Username' Error:Field validation for 'Username' failed on the 'min' tag")
	assert.Equal(t, "Username must be between 3 and 20 characters.", auth.CleanAuthError(rawMinErr))

	// Connection refused
	assert.Equal(t, "Chat server is currently offline or unreachable.", auth.CleanAuthError(errors.New("dial tcp 127.0.0.1:8080: connect: connection refused")))
}

func TestEnforceDefaultBackground(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	sampleText := theme.StyleTitle.Render("LINETALK") + " Normal text"
	rendered := theme.EnforceDefaultBackground(sampleText, 40, 2)

	// Must contain the ANSI TrueColor background sequence for ColorDarkBg (#0B0C10 -> 11;12;16)
	expectedBgSeq := "\x1b[48;2;11;12;16m"
	assert.Contains(t, rendered, expectedBgSeq, "Rendered output must enforce default dark background")

	// Normal text must be preceded by background escape sequence after reset
	lines := strings.Split(rendered, "\n")
	assert.Equal(t, 2, len(lines), "Should pad to requested height")

	for _, l := range lines {
		assert.True(t, strings.HasPrefix(l, expectedBgSeq), "Every line must start with default background")
	}
}

func TestStatusBarNotificationColors(t *testing.T) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(termenv.Ascii)

	bar := statusbar.Render(nil, false, statusbar.StatusConnected, 0, "", "🔔 @alice: hello world", 100)
	enforced := theme.EnforceDefaultBackground(bar, 100, 1)

	// Green background sequence for badge: #00FF85 -> 48;2;0;255;133
	greenBgSeq := "\x1b[48;2;0;255;133m"
	assert.Contains(t, enforced, greenBgSeq, "Status bar notification must contain green background for the badge")

	// Yellow background sequence for banner: #FFE600 -> 48;2;255;230;0
	yellowBgSeq := "\x1b[48;2;255;230;0m"
	assert.Contains(t, enforced, yellowBgSeq, "Status bar notification must contain yellow background for the message banner")

	assert.Contains(t, enforced, "alice", "Status bar notification must include sender name")
	assert.Contains(t, enforced, "hello world", "Status bar notification must include message content")
}
