package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

// DefaultAPIURL and DefaultGoEnv can be customized at compile time via -ldflags:
//
//	go build -ldflags "-X main.DefaultAPIURL=https://api.yourdomain.com -X main.DefaultGoEnv=development" ./cmd/tui
var (
	DefaultAPIURL = "http://localhost:8080"
	DefaultGoEnv  = "production"
)

func main() {
	var sessionFlag string
	flag.StringVar(&sessionFlag, "session", "", "Session profile name (e.g. alice, bob) for testing multiple users")
	flag.StringVar(&sessionFlag, "s", "", "Shorthand for -session")
	flag.Parse()

	// Attempt to load .env from current directory or parent directories
	_ = godotenv.Load(".env", "../../.env", "../../../.env")

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	goEnv := os.Getenv("GO_ENV")
	if goEnv == "" {
		goEnv = DefaultGoEnv
	}

	sessionName := sessionFlag
	if sessionName == "" {
		sessionName = os.Getenv("SESSION")
		if sessionName == "" {
			sessionName = os.Getenv("TUI_SESSION")
		}
	}

	c := client.NewHTTPClient(apiURL)

	sessionPath := client.ResolveSessionPath(goEnv, sessionName)
	c.SetSessionPath(sessionPath)

	// Attempt to load existing saved session
	if savedProf, err := client.LoadSession(sessionPath); err == nil && savedProf != nil && savedProf.Token != "" {
		c.SetProfile(*savedProf)
	}

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
