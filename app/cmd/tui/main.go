package main

import (
	"fmt"
	"os"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

// DefaultAPIURL can be customized at compile time via -ldflags:
//
//	go build -ldflags "-X main.DefaultAPIURL=https://api.yourdomain.com" ./cmd/tui
var DefaultAPIURL = "http://localhost:8080"

func main() {
	// Attempt to load .env from current directory or parent directories
	_ = godotenv.Load(".env", "../../.env", "../../../.env")

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	c := client.NewHTTPClient(apiURL)

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
