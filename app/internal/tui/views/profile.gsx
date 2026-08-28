package views

import (
	tui "github.com/grindlemire/go-tui"
)

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
		<span class="font-bold text-cyan">Profile</span>
		if p.editMode.Get() {
			<span class="font-dim">Username</span>
			<input value={p.username} placeholder="Username" border={tui.BorderRounded} />
			<span class="font-dim">Account Code</span>
			<input value={p.accountCode} placeholder="Account Code" border={tui.BorderRounded} />
			<span class="font-dim">Password</span>
			<input value={p.password} placeholder="Password" border={tui.BorderRounded} />
			<span class="text-green font-dim px-1">Tab between fields | Enter save | Esc cancel</span>
		} else {
			<span class="font-dim">Username</span>
			<span class="px-1">{p.username.Get()}</span>
			<span class="font-dim">Account Code</span>
			<span class="px-1">{p.accountCode.Get()}</span>
			<span class="font-dim">Password</span>
			<span class="px-1">********</span>
			<span class="font-dim px-1">Press e to edit</span>
		}
	</div>
}