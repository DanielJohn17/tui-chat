package test

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views/auth"
	tea "github.com/charmbracelet/bubbletea"
)

func authenticatedApp() tea.Model {
	c := newTestClient()
	model := tea.Model(tui.NewApp(c))
	model, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 35})
	model, _ = model.Update(auth.AuthSuccessMsg{
		Profile: c.Profile(),
	})
	return model
}
