package main

import (
	"fmt"
	"os"

	"github.com/DanielJohn17/tui-chat/app/internal/tui"
)

func main() {
	app := tui.New()
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
