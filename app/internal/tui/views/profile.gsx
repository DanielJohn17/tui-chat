package views

import tui "github.com/grindlemire/go-tui"

type profile struct {
	username    *tui.State[string]
	accountCode *tui.State[string]
	password    *tui.State[string]
	editMode    *tui.State[bool]
	onSave      func()
	onCancel    func()
}

func Profile(username *tui.State[string], accountCode *tui.State[string], password *tui.State[string], editMode *tui.State[bool], onSave func(), onCancel func()) *profile {
	return &profile{
		username:    username,
		accountCode: accountCode,
		password:    password,
		editMode:    editMode,
		onSave:      onSave,
		onCancel:    onCancel,
	}
}

func (p *profile) KeyMap() tui.KeyMap {
	if p.editMode.Get() {
		return tui.KeyMap{
			tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { p.onCancel() }),
			tui.OnStop(tui.KeyEnter, func(ke tui.KeyEvent) { p.onSave() }),
		}
	}
	return tui.KeyMap{
		tui.On(tui.Rune('e'), func(ke tui.KeyEvent) { p.editMode.Set(true) }),
	}
}

templ (p *profile) Render() {
	<div class="flex-col grow px-2 gap-1">
		<div class="flex justify-between items-center shrink-0">
			<span class="font-bold text-cyan">User Profile</span>
			if p.editMode.Get() {
				<span class="text-yellow font-bold">[Editing Mode]</span>
			} else {
				<span class="font-dim">Press [e] to edit</span>
			}
		</div>
		<hr />
		if p.editMode.Get() {
			<div class="flex-col gap-1">
				<span class="font-bold text-white">Username</span>
				<input value={p.username} placeholder="Enter username..." border={tui.BorderRounded} width={45} focusColor={tui.Cyan} autoFocus={true} />
				<span class="font-bold text-white">Account Code</span>
				<input value={p.accountCode} placeholder="Enter account code..." border={tui.BorderRounded} width={45} focusColor={tui.Cyan} />
				<span class="font-bold text-white">Password</span>
				<input value={p.password} placeholder="Enter password..." border={tui.BorderRounded} width={45} focusColor={tui.Cyan} />
				<div class="flex gap-2 pt-1">
					<span class="text-green font-bold">[Enter] Save Changes</span>
					<span class="font-dim">|</span>
					<span class="text-red font-bold">[Esc] Cancel</span>
				</div>
			</div>
		} else {
			<div class="flex-col gap-1">
				<div class="flex-col border-single px-2" width={48}>
					<div class="flex justify-between">
						<span class="font-bold text-white">Account Information</span>
						<span class="text-green font-bold">{"\u25cf Active"}</span>
					</div>
					<hr />
					<div class="flex justify-between">
						<span class="font-dim">Username:</span>
						<span class="font-bold text-cyan">{p.username.Get()}</span>
					</div>
					<div class="flex justify-between">
						<span class="font-dim">Account Code:</span>
						<span class="font-bold text-white">{p.accountCode.Get()}</span>
					</div>
					<div class="flex justify-between">
						<span class="font-dim">Password:</span>
						<span class="font-dim">{"\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022"}</span>
					</div>
				</div>
				<span class="font-dim pt-1">Press [e] to edit profile details</span>
			</div>
		}
	</div>
}
