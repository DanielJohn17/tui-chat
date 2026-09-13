package views

import (
	"fmt"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tui "github.com/grindlemire/go-tui"
)

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "…"
}

templ Sidebar(c client.Client, selectedIndex *tui.State[int]) {
	<div class="flex-col border-rounded shrink-0 px-1 py-0 gap-0" width={30}>
		<div class="flex justify-between items-center shrink-0 pt-0">
			<span class="text-magenta font-bold">{"◈ CONVERSATIONS"}</span>
			<span class="text-green font-bold">[+n]</span>
		</div>
		<hr />
		<div class="flex-col gap-1 grow overflow-y-scroll scrollbar-magenta">
			for i, ch := range c.Chats() {
				if i == selectedIndex.Get() {
					<div class="flex-col px-1">
						<div class="flex justify-between items-center">
							<div class="flex items-center gap-1">
								<span class="text-magenta font-bold">{"▶"}</span>
								<span class="text-cyan font-bold">{truncate(ch.Name, 15)}</span>
							</div>
							if ch.Online {
								<span class="text-green font-bold">{"●"}</span>
							} else {
								<span class="font-dim">{"○"}</span>
							}
						</div>
						<div class="flex justify-between items-center">
							<span class="text-magenta font-dim">{"@" + truncate(ch.Username, 12)}</span>
							<span class="font-dim">{ch.Time}</span>
						</div>
						if ch.LastMessage != "" {
							<span class="font-dim">{truncate(ch.LastMessage, 22)}</span>
						}
					</div>
				} else {
					<div class="flex-col px-1">
						<div class="flex justify-between items-center">
							<div class="flex items-center gap-1">
								if ch.Online {
									<span class="text-green font-dim">{"●"}</span>
								} else {
									<span class="font-dim">{"○"}</span>
								}
								<span class="text-white">{truncate(ch.Name, 14)}</span>
							</div>
							if ch.Unread > 0 {
								<span class="text-magenta font-bold">{fmt.Sprintf("(%d)", ch.Unread)}</span>
							} else {
								<span class="font-dim">{ch.Time}</span>
							}
						</div>
						<span class="font-dim">{"@" + truncate(ch.Username, 14)}</span>
					</div>
				}
			}
		</div>
		<hr />
		<div class="flex items-center justify-between shrink-0 px-1 pb-0">
			<div class="flex items-center gap-1">
				<span class="text-green font-bold">{"●"}</span>
				<span class="font-bold text-white">{truncate(c.Profile().Name, 12)}</span>
			</div>
			<span class="text-magenta font-dim">{"@" + truncate(c.Profile().Username, 10)}</span>
		</div>
	</div>
}


