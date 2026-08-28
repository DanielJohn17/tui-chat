package views

import (
	"fmt"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tui "github.com/grindlemire/go-tui"
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
		<div class="flex justify-between items-center shrink-0">
			<span class="font-bold text-cyan">{chatName}</span>
			<span class="font-dim">{fmt.Sprintf("%d messages", len(messages))}</span>
		</div>
		<hr />
		<div ref={c.msgRef} class="flex-col grow overflow-y-scroll scrollbar-cyan gap-1">
			for _, msg := range messages {
				if msg.Sender == "system" {
					<div class="flex gap-1">
						<span class="text-yellow font-bold">[system]</span>
						<span class="text-yellow font-dim">{msg.Text}</span>
					</div>
				} else if msg.Self {
					<div class="flex gap-1">
						<span class="text-cyan font-bold">❯ you:</span>
						<span class="text-white">{msg.Text}</span>
					</div>
				} else {
					<div class="flex gap-1">
						<span class="text-magenta font-bold">{msg.Sender + ":"}</span>
						<span class="text-white">{msg.Text}</span>
					</div>
				}
			}
			if len(messages) == 0 {
				<span class="font-dim">No messages yet in this channel. Send the first message!</span>
			}
		</div>
		<div class="shrink-0 pt-1">
			<input value={c.draft} onSubmit={c.onSubmit} placeholder="Type a message and press Enter..." border={tui.BorderRounded} width={100} focusColor={tui.Cyan} autoFocus={true} />
		</div>
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
