package test

import (
	"strings"
	"testing"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
)

func authenticatedApp() tea.Model {
	c := client.NewMock()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	// Authenticate
	model, _ = model.Update(auth.AuthSuccessMsg{
		Profile: c.Profile(),
	})
	return model
}

func TestProfileEditPageFlow(t *testing.T) {
	model := authenticatedApp()

	// 1. Press 'p' to open Profile Edit page
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	view := model.View()
	if !strings.Contains(view, "PROFILE EDIT") {
		t.Fatalf("expected Profile Edit page to be open, got view:\n%s", view)
	}

	// 2. Press Esc to return to Chat
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = model.View()
	if !strings.Contains(view, "CONVERSATIONS") {
		t.Fatalf("expected Chat view after Esc, got:\n%s", view)
	}

	// 3. Re-open Profile Edit page and Save
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		msg := cmd()
		model, _ = model.Update(msg)
	}
	view = model.View()
	if !strings.Contains(view, "CONVERSATIONS") {
		t.Fatalf("expected Chat view after saving profile, got:\n%s", view)
	}
}

func TestSidebarBottomUserInfoAndScrolling(t *testing.T) {
	model := authenticatedApp()
	view := model.View()

	// Verify user handle appears in the bottom sidebar
	if !strings.Contains(view, "@alice") {
		t.Fatalf("expected user handle in sidebar, got:\n%s", view)
	}

	// Verify mock contacts (at least 10+ in list)
	c := client.NewMock()
	chats := c.Chats()
	if len(chats) < 10 {
		t.Fatalf("expected at least 10 mock chats for scroll testing, got %d", len(chats))
	}

	// Verify messages in Chat 1 is at least 20+
	msgs := c.Messages(1)
	if len(msgs) < 20 {
		t.Fatalf("expected at least 20 messages in chat 1 for scrolling, got %d", len(msgs))
	}

	// Test mouse wheel down on sidebar scrolls down
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonWheelDown,
		Type:   tea.MouseWheelDown,
	})
	view = model.View()
	if !strings.Contains(view, "CONVERSATIONS") {
		t.Fatalf("expected Chat view after mouse wheel, got:\n%s", view)
	}
}

func TestHelpKeybindings(t *testing.T) {
	model := authenticatedApp()

	// 1. 'h' should NOT open Help Desk anymore
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	view := model.View()
	if strings.Contains(view, "HELP DESK") {
		t.Fatalf("'h' key should not open Help Desk modal; user requested ? and ctrl+h")
	}

	// 2. '?' MUST open Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	view = model.View()
	if !strings.Contains(view, "HELP DESK") {
		t.Fatalf("'?' key should open Help Desk modal")
	}

	// 3. Esc closes Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	view = model.View()
	if strings.Contains(view, "HELP DESK") {
		t.Fatalf("Esc should close Help Desk")
	}

	// 4. Ctrl+H MUST open Help Desk
	model, _ = model.Update(tea.KeyMsg{Type: tea.KeyCtrlH})
	view = model.View()
	if !strings.Contains(view, "HELP DESK") {
		t.Fatalf("Ctrl+H should open Help Desk modal")
	}
}

func TestMouseClickActions(t *testing.T) {
	model := authenticatedApp()

	// Initial active chat is Bob Martin
	initView := model.View()
	lines := strings.Split(initView, "\n")
	for idx, line := range lines {
		if idx < 20 {
			t.Logf("LINE %d: %s", idx, line)
		}
	}

	// 1. Mouse click on second chat (Sarah Connor, line 6)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      6, // Sarah Connor row
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view := model.View()
	if !strings.Contains(view, "@sarah") {
		t.Fatalf("expected active chat to switch to @sarah after clicking, got:\n%s", view)
	}

	// 2. Mouse click on bottom profile row in sidebar opens profile edit page
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      31, // Bottom row of sidebar (height 32)
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view = model.View()
	if !strings.Contains(view, "PROFILE EDIT") {
		t.Fatalf("clicking bottom sidebar profile should open profile edit page, got:\n%s", view)
	}
}

func TestSidebarClickWhenItemExpands(t *testing.T) {
	model := authenticatedApp()

	// Initial render: Bob Martin is selected (lines 3-5). Sarah is lines 6-7. Charlie Zhang is lines 8-9.
	_ = model.View()

	// 1. Click on Charlie Zhang (line 8)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      8,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view := model.View()
	if !strings.Contains(view, "@charlie") {
		t.Fatalf("expected @charlie to be active after click, got:\n%s", view)
	}

	// Now Charlie Zhang is expanded to 3 lines (lines 7, 8, 9).
	// Alex Rivera is now at lines 10, 11!
	// 2. Click on Alex Rivera (line 10)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      10,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view = model.View()
	if !strings.Contains(view, "@alex") {
		t.Fatalf("expected @alex to be active after click below expanded item, got:\n%s", view)
	}

	// Now Alex Rivera is expanded (lines 9-11).
	// Bob Martin above is 2 lines (lines 3-4).
	// 3. Click on Bob Martin (line 3)
	model, _ = model.Update(tea.MouseMsg{
		X:      10,
		Y:      3,
		Button: tea.MouseButtonLeft,
		Action: tea.MouseActionPress,
	})
	view = model.View()
	if !strings.Contains(view, "@bob") {
		t.Fatalf("expected @bob to be active after clicking back on shrunk item, got:\n%s", view)
	}
}
