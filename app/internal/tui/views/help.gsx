package views

import tui "github.com/grindlemire/go-tui"

type helpModal struct {
	open *tui.State[bool]
}

func HelpModal(open *tui.State[bool]) *helpModal {
	return &helpModal{open: open}
}

templ (h *helpModal) Render() {
	<modal open={h.open} class="justify-center items-center" backdrop="dim">
		<div class="flex-col border-rounded p-2 gap-1 bg-black" width={52}>
			<div class="flex justify-between items-center">
				<span class="font-bold text-magenta">Commands & Shortcuts</span>
				<span class="font-dim text-yellow">esc</span>
			</div>
			<hr />
			<span class="font-bold text-cyan">Suggested</span>
			<div class="flex justify-between items-center">
				<span class="text-white">Switch conversation</span>
				<span class="text-magenta font-bold">j / k or ↑ / ↓</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">New conversation</span>
				<span class="text-magenta font-bold">n</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">User profile & identity</span>
				<span class="text-magenta font-bold">p</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">Focus message input</span>
				<span class="text-cyan font-bold">Tab</span>
			</div>
			<hr />
			<span class="font-bold text-cyan">Actions & Navigation</span>
			<div class="flex justify-between items-center">
				<span class="text-white">Send message</span>
				<span class="text-cyan font-bold">Enter</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">Return to chats</span>
				<span class="text-magenta font-bold">c / Esc</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">Edit user profile</span>
				<span class="text-magenta font-bold">e</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">Toggle shortcuts help</span>
				<span class="text-yellow font-bold">h / ?</span>
			</div>
			<div class="flex justify-between items-center">
				<span class="text-white">Quit messenger</span>
				<span class="text-magenta font-bold">q</span>
			</div>
		</div>
	</modal>
}

