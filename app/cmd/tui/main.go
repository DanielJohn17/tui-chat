package main

import (
	"fmt"
	"os"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize HTTP Client (connecting to Gin API with seamless mock fallback)
	c := client.NewHTTPClient("")

	model := tui.NewApp(c)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
