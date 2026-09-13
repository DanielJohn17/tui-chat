package views

import (
	"strings"
	tui "github.com/grindlemire/go-tui"
)

type newDMModal struct {
	open     *tui.State[bool]
	username *tui.State[string]
	onStart  func(username string)
	onCancel func()
}

func NewDMModal(open *tui.State[bool], username *tui.State[string], onStart func(username string), onCancel func()) *newDMModal {
	return &newDMModal{
		open:     open,
		username: username,
		onStart:  onStart,
		onCancel: onCancel,
	}
}

func (n *newDMModal) submit() {
	u := strings.TrimSpace(n.username.Get())
	u = strings.TrimPrefix(u, "@")
	if u == "" {
		return
	}
	n.onStart(u)
}

templ (n *newDMModal) Render() {
	<modal open={n.open} class="justify-center items-center" backdrop="dim">
		<div class="flex-col border-rounded p-2 gap-1 bg-black" width={52}>
			<div class="flex justify-between items-center">
				<span class="font-bold text-magenta">Start New Conversation</span>
				<span class="font-dim text-yellow">esc</span>
			</div>
			<hr />
			<span class="font-bold text-cyan">Recipient Username</span>
			<input value={n.username} onSubmit={func(string) { n.submit() }} placeholder="e.g. sarah or @sarah..." border={tui.BorderRounded} width={45} focusColor={tui.Magenta} autoFocus={true} />
			<div class="flex justify-between items-center pt-1">
				<div class="flex gap-1">
					<span class="text-green font-bold">[Enter]</span>
					<span class="text-white">Start</span>
					<span class="font-dim">•</span>
					<span class="text-yellow font-bold">[Esc]</span>
					<span class="text-white">Cancel</span>
				</div>
				<span class="font-dim text-magenta">[q] Quit</span>
			</div>
		</div>
	</modal>
}
