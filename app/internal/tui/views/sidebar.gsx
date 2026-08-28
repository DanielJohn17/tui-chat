package views

import (
	tui "github.com/grindlemire/go-tui"

	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
)

templ Sidebar(c client.Client, selectedChat *tui.State[int]) {
	<div class="flex-col border-single shrink-0 px-1" width={22}>
		<span class="text-cyan font-bold">Chats</span>
		<hr />
		for i, ch := range c.Chats() {
			if i == selectedChat.Get() {
				<span class="text-cyan font-bold">{"> " + ch.Name}</span>
			} else {
				<span class="font-dim">{"  " + ch.Name}</span>
			}
		}
	</div>
}