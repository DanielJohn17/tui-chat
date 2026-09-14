package test

import (
	"testing"
	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAuthPassThrough(t *testing.T) {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Test Enter passes straight to chat app without API call
	newModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from Enter")
	}

	// Execute cmd
	msg := cmd()
	newModel, _ = newModel.Update(msg)
	view := newModel.View()
	if !contains(view, "CONVERSATIONS") {
		t.Fatalf("expected view to transition to chat app, got:\n%s", view)
	}
}

func TestAuthHelpModal(t *testing.T) {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 2. Test F1 opens Help Desk modal
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyF1})
	view := newModel.View()
	if !contains(view, "HELP DESK") {
		t.Fatalf("expected Help Desk modal, got:\n%s", view)
	}

	// 3. Test Esc closes Help Desk modal and returns to auth card
	newModel, _ = newModel.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = newModel.View()
	if !contains(view, "LOGIN") {
		t.Fatalf("expected return to Login card, got:\n%s", view)
	}
}

func TestAuthEscQuits(t *testing.T) {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 4. Test Esc in auth quits the app
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected quit cmd on Esc")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
