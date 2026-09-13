package views

import tui "github.com/grindlemire/go-tui"

type newDM struct {
	name     *tui.State[string]
	username *tui.State[string]
	onStart  func()
	onCancel func()
}

func NewDM(name *tui.State[string], username *tui.State[string], onStart func(), onCancel func()) *newDM {
	return &newDM{
		name:     name,
		username: username,
		onStart:  onStart,
		onCancel: onCancel,
	}
}

func (n *newDM) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { n.onCancel() }),
		tui.OnStop(tui.KeyEnter, func(ke tui.KeyEvent) { n.onStart() }),
	}
}

templ (n *newDM) Render() {
	<div class="flex-col grow border-rounded px-2 py-0 gap-1">
		<div class="flex justify-between items-center shrink-0 pt-0">
			<span class="font-bold text-magenta">{"◈ START NEW CONVERSATION"}</span>
			<span class="text-green font-bold">[ POST /api/v1/conversations ]</span>
		</div>
		<hr />
		<div class="flex-col gap-1 px-1 py-1" width={60}>
			<span class="font-bold text-magenta">Recipient Display Name</span>
			<input value={n.name} placeholder="e.g. David Kim" border={tui.BorderRounded} width={50} focusColor={tui.Magenta} autoFocus={true} />
			<span class="font-bold text-magenta">Recipient Username</span>
			<input value={n.username} placeholder="e.g. david" border={tui.BorderRounded} width={50} focusColor={tui.Magenta} />
			<div class="flex gap-2 pt-1">
				<span class="text-green font-bold">[Enter] Start Chat</span>
				<span class="font-dim">•</span>
				<span class="text-magenta font-bold">[Esc] Cancel</span>
			</div>
			<span class="font-dim pt-1">Creates or fetches direct message thread for peer user ID</span>
		</div>
	</div>
}
