package views

import (
	"fmt"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tui "github.com/grindlemire/go-tui"
)

type chatPane struct {
	client        client.Client
	selectedIndex *tui.State[int]
	draft         *tui.State[string]
	onSend        func()
	msgRef        *tui.Ref
	app           *tui.App
}

func ChatPane(c client.Client, selectedIndex *tui.State[int], draft *tui.State[string], onSend func()) *chatPane {
	return &chatPane{
		client:        c,
		selectedIndex: selectedIndex,
		draft:         draft,
		onSend:        onSend,
		msgRef:        tui.NewRef(),
	}
}

func (c *chatPane) BindApp(app *tui.App) {
	c.app = app
}

func (c *chatPane) currentChat() client.Chat {
	chats := c.client.Chats()
	idx := c.selectedIndex.Get()
	if idx >= 0 && idx < len(chats) {
		return chats[idx]
	}
	if len(chats) > 0 {
		return chats[0]
	}
	return client.Chat{Name: "General", Username: "chat"}
}

func (c *chatPane) messages() []client.Message {
	chat := c.currentChat()
	if chat.ID > 0 {
		return c.client.Messages(chat.ID)
	}
	return nil
}

func (c *chatPane) onSubmit(text string) {
	if text == "" {
		return
	}
	chat := c.currentChat()
	if chat.ID <= 0 {
		return
	}
	c.client.Send(chat.ID, text)
	c.draft.Set("")
	c.onSend()
	c.app.QueueUpdate(func() {
		if el := c.msgRef.El(); el != nil {
			el.ScrollToBottom()
		}
	})
}

templ (c *chatPane) Render() {
	<div class="flex-col grow border-rounded px-2 py-0 gap-1">
		<div class="flex justify-between items-center shrink-0 pt-0">
			<div class="flex items-center gap-1">
				if c.currentChat().Online {
					<span class="text-green font-bold">{"●"}</span>
					<span class="text-green font-bold">Online</span>
				} else {
					<span class="font-dim">{"○"}</span>
					<span class="font-dim">Offline</span>
				}
				<span class="font-dim">•</span>
				<span class="font-bold text-white">{c.currentChat().Name}</span>
				<span class="font-bold text-magenta">{"(@" + c.currentChat().Username + ")"}</span>
			</div>
			<div class="flex items-center gap-1">
				<span class="font-dim">{fmt.Sprintf("%d messages", len(c.messages()))}</span>
			</div>
		</div>
		<hr />
		<div ref={c.msgRef} class="flex-col grow overflow-y-scroll scrollbar-magenta gap-1 px-1">
			for _, msg := range c.messages() {
				if msg.Sender == "system" {
					<div class="flex-col pb-1">
						<div class="flex items-center gap-1">
							<span class="text-yellow font-bold">{"◆ SYSTEM"}</span>
							<span class="font-dim">{msg.Timestamp}</span>
						</div>
						<div class="flex px-1">
							<span class="text-yellow font-dim">{msg.Text}</span>
						</div>
					</div>
				} else if msg.Self {
					<div class="flex-col pb-1">
						<div class="flex justify-between items-center">
							<div class="flex items-center gap-1">
								<span class="text-cyan font-bold">{"● @" + msg.Sender}</span>
								<span class="text-cyan font-dim">(You)</span>
							</div>
							<span class="font-dim">{msg.Timestamp}</span>
						</div>
						<div class="flex px-1">
							<span class="text-white font-bold">{msg.Text}</span>
						</div>
					</div>
				} else {
					<div class="flex-col pb-1">
						<div class="flex justify-between items-center">
							<div class="flex items-center gap-1">
								<span class="text-magenta font-bold">{"● @" + msg.Sender}</span>
							</div>
							<span class="font-dim">{msg.Timestamp}</span>
						</div>
						<div class="flex px-1">
							<span class="text-white">{msg.Text}</span>
						</div>
					</div>
				}
			}
			if len(c.messages()) == 0 {
				<div class="flex-col items-center justify-center grow gap-1">
					<span class="text-magenta font-bold">{"✨ Conversation Thread Opened"}</span>
					<span class="font-dim">{fmt.Sprintf("Type a message below to chat with %s (@%s)", c.currentChat().Name, c.currentChat().Username)}</span>
				</div>
			}
		</div>
		<div class="shrink-0 pb-1">
			<input value={c.draft} onSubmit={c.onSubmit} placeholder={"Message @" + c.currentChat().Username + "... (Press Tab to type, Enter to send)"} border={tui.BorderRounded} width={100} focusColor={tui.Magenta} />
		</div>
	</div>
}

