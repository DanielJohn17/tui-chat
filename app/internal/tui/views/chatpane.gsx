package views

import (
	"fmt"

	tui "github.com/grindlemire/go-tui"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
)

type chatPane struct {
	client     client.Client
	selectedID *tui.State[int]
	draft      *tui.State[string]
	onSend     func()
	msgRef     *tui.Ref
	app        *tui.App
}

func ChatPane(c client.Client, selectedID *tui.State[int], draft *tui.State[string], onSend func()) *chatPane {
	return &chatPane{
		client:     c,
		selectedID: selectedID,
		draft:      draft,
		onSend:     onSend,
		msgRef:     tui.NewRef(),
	}
}

func (c *chatPane) BindApp(app *tui.App) {
	c.app = app
}

func (c *chatPane) onSubmit(text string) {
	if text == "" {
		return
	}
	c.client.Send(c.selectedID.Get(), text)
	c.draft.Set("")
	c.onSend()
	c.app.QueueUpdate(func() {
		if el := c.msgRef.El(); el != nil {
			el.ScrollToBottom()
		}
	})
}

templ (c *chatPane) Render() {
	chatName := chatNameFor(c.client, c.selectedID.Get())
	messages := c.client.Messages(c.selectedID.Get())
	<div class="flex-col grow px-1">
		<span class="font-bold text-cyan">{chatName}</span>
		<hr />
		<div class="flex-col grow overflow-y-scroll scrollbar-cyan" ref={c.msgRef}>
			for _, msg := range messages {
				if msg.Self {
					<span class="text-cyan">{"> " + msg.Text}</span>
				} else {
					<span class="font-dim">{msg.Sender + ": " + msg.Text}</span>
				}
			}
			if len(messages) == 0 {
				<span class="font-dim">No messages yet</span>
			}
		</div>
		<input value={c.draft} onSubmit={c.onSubmit} placeholder="Type a message..." border={tui.BorderRounded} autoFocus={true} />
	</div>
}

func chatNameFor(c client.Client, id int) string {
	for _, ch := range c.Chats() {
		if ch.ID == id {
			return "#" + ch.Name
		}
	}
	return fmt.Sprintf("#chat-%d", id)
}