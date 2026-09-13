package views

import (
	"fmt"
	tui "github.com/grindlemire/go-tui"
)

type profile struct {
	name        *tui.State[string]
	username    *tui.State[string]
	password    *tui.State[string]
	userID      int64
	token       string
	createdAt   string
	editMode    *tui.State[bool]
	onSave      func()
	onCancel    func()
}

func Profile(
	name *tui.State[string],
	username *tui.State[string],
	password *tui.State[string],
	userID int64,
	token string,
	createdAt string,
	editMode *tui.State[bool],
	onSave func(),
	onCancel func(),
) *profile {
	return &profile{
		name:      name,
		username:  username,
		password:  password,
		userID:    userID,
		token:     token,
		createdAt: createdAt,
		editMode:  editMode,
		onSave:    onSave,
		onCancel:  onCancel,
	}
}

templ (p *profile) Render() {
	<div class="flex-col grow border-rounded px-2 py-0 gap-1">
		<div class="flex justify-between items-center shrink-0 pt-0">
			<span class="font-bold text-magenta">{"◈ USER IDENTITY & API AUTH"}</span>
			if p.editMode.Get() {
				<span class="text-yellow font-bold">[ Editing Mode ]</span>
			} else {
				<span class="text-green font-bold">[ Press 'e' to Edit ]</span>
			}
		</div>
		<hr />
		if p.editMode.Get() {
			<div class="flex-col gap-1 px-1">
				<span class="font-bold text-magenta">Display Name</span>
				<input value={p.name} placeholder="Enter full name..." border={tui.BorderRounded} width={50} focusColor={tui.Magenta} autoFocus={true} />
				<span class="font-bold text-magenta">Username</span>
				<input value={p.username} placeholder="Enter username..." border={tui.BorderRounded} width={50} focusColor={tui.Magenta} />
				<span class="font-bold text-magenta">Password</span>
				<input value={p.password} placeholder="Enter password..." border={tui.BorderRounded} width={50} focusColor={tui.Magenta} />
				<div class="flex gap-2 pt-1">
					<span class="text-green font-bold">[Enter] Save Changes</span>
					<span class="font-dim">•</span>
					<span class="text-magenta font-bold">[Esc] Cancel</span>
				</div>
			</div>
		} else {
			<div class="flex-col gap-1 px-1 py-1" width={60}>
				<div class="flex justify-between items-center">
					<span class="font-bold text-white">Active Session Details</span>
					<span class="text-green font-bold">● Authenticated</span>
				</div>
				<hr />
				<div class="flex justify-between">
					<span class="font-dim">User ID:</span>
					<span class="font-bold text-yellow">{fmt.Sprintf("#%d", p.userID)}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">Display Name:</span>
					<span class="font-bold text-cyan">{p.name.Get()}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">Username:</span>
					<span class="font-bold text-magenta">{"@" + p.username.Get()}</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">Password:</span>
					<span class="font-dim">••••••••••••</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">API Auth Token:</span>
					<span class="text-cyan">JWT Signed (Bearer)</span>
				</div>
				<div class="flex justify-between">
					<span class="font-dim">Registered:</span>
					<span class="font-dim">{p.createdAt}</span>
				</div>
				<span class="font-dim pt-1">Press [e] to modify user details or [c / Esc] to return to chats</span>
			</div>
		}
	</div>
}

