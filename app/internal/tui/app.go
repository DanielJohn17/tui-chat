package tui

import (
	tui "github.com/grindlemire/go-tui"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/views"
)

func New() *tui.App {
	c := client.NewMock()
	app, err := tui.NewApp(
		tui.WithRootComponent(views.App(c)),
	)
	if err != nil {
		panic(err)
	}
	return app
}