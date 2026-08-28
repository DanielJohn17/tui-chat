package views

import (
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tui "github.com/grindlemire/go-tui"
)

templ Sidebar(c client.Client, selectedChat *tui.State[int]) {
	<div class="flex-col border-single shrink-0 px-1" width={24}>
		<div class="flex justify-between items-center shrink-0">
			<span class="text-cyan font-bold">CHANNELS</span>
		</div>
		<hr />
		<div class="flex-col gap-1">
			for i, ch := range c.Chats() {
				if i == selectedChat.Get() {
					<span class="text-cyan font-bold">{"\u276f #" + ch.Name}</span>
				} else {
					<span class="font-dim">{"  #" + ch.Name}</span>
				}
			}
		</div>
		<div class="grow"></div>
		<hr />
		<div class="flex items-center gap-1 shrink-0">
			<span class="text-green font-bold">{"\u25cf"}</span>
			<span class="font-bold text-white">{c.Profile().Username}</span>
		</div>
	</div>
}
