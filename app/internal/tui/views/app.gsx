package views

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"github.com/DanielJohn17/tui-chat/app/internal/tui/client"
	tui "github.com/grindlemire/go-tui"
)

type viewMode int

const (
	viewChats viewMode = iota
	viewProfile
)

type app struct {
	client        client.Client
	view          *tui.State[viewMode]
	selectedChat  *tui.State[int]
	draft         *tui.State[string]
	name          *tui.State[string]
	username      *tui.State[string]
	password      *tui.State[string]
	newDMUsername *tui.State[string]
	showNewDM     *tui.State[bool]
	showHelp      *tui.State[bool]
	profileEdit   *tui.State[bool]
	replyPending  *tui.State[int]
}

func App(c client.Client) *app {
	p := c.Profile()
	return &app{
		client:        c,
		view:          tui.NewState(viewChats),
		selectedChat:  tui.NewState(0),
		draft:         tui.NewState(""),
		name:          tui.NewState(p.Name),
		username:      tui.NewState(p.Username),
		password:      tui.NewState(p.Password),
		newDMUsername: tui.NewState(""),
		showNewDM:     tui.NewState(false),
		showHelp:      tui.NewState(false),
		profileEdit:   tui.NewState(false),
		replyPending:  tui.NewState(-1),
	}
}

func (a *app) KeyMap() tui.KeyMap {
	// 1. Help Modal Active
	if a.showHelp.Get() {
		return tui.KeyMap{
			tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
			tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
			tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
			tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
			tui.OnStop(tui.Rune('?'), func(ke tui.KeyEvent) { a.showHelp.Set(false) }),
		}
	}

	// 2. New DM Modal Active
	if a.showNewDM.Get() {
		return tui.KeyMap{
			tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
			tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.cancelNewDM() }),
		}
	}

	// 3. Profile View Active
	if a.view.Get() == viewProfile {
		if a.profileEdit.Get() {
			return tui.KeyMap{
				tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
				tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.cancelProfile() }),
			}
		}
		return tui.KeyMap{
			tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
			tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
			tui.OnStop(tui.Rune('c'), func(ke tui.KeyEvent) { a.view.Set(viewChats) }),
			tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { a.view.Set(viewChats) }),
			tui.OnStop(tui.Rune('e'), func(ke tui.KeyEvent) { a.profileEdit.Set(true) }),
			tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
			tui.OnStop(tui.Rune('?'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
		}
	}

	// 4. Default Chat View
	chats := a.client.Chats()
	moveUp := func(ke tui.KeyEvent) {
		a.selectedChat.Update(func(v int) int {
			if v <= 0 {
				return len(chats) - 1
			}
			return v - 1
		})
	}
	moveDown := func(ke tui.KeyEvent) {
		a.selectedChat.Update(func(v int) int {
			if v >= len(chats)-1 {
				return 0
			}
			return v + 1
		})
	}

	return tui.KeyMap{
		tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
		tui.OnStop(tui.KeyTab.Shift(), func(ke tui.KeyEvent) { ke.App().FocusPrev() }),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) { a.view.Set(viewProfile); a.profileEdit.Set(false) }),
		tui.OnStop(tui.Rune('n'), func(ke tui.KeyEvent) { a.showNewDM.Set(true) }),
		tui.OnStop(tui.Rune('h'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
		tui.OnStop(tui.Rune('?'), func(ke tui.KeyEvent) { a.showHelp.Set(true) }),
		tui.On(tui.KeyUp, moveUp),
		tui.On(tui.Rune('k'), moveUp),
		tui.On(tui.KeyDown, moveDown),
		tui.On(tui.Rune('j'), moveDown),
	}
}

func (a *app) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(2*time.Second, func() {
			if a.replyPending.Get() >= 0 {
				chatIdx := a.replyPending.Get()
				chats := a.client.Chats()
				if chatIdx >= 0 && chatIdx < len(chats) {
					chat := chats[chatIdx]
					reply := mockReplies[rand.IntN(len(mockReplies))]
					a.client.Send(chat.ID, reply)
				}
				a.replyPending.Set(-1)
			}
		}),
	}
}

func (a *app) profile() client.Profile {
	return a.client.Profile()
}

templ (a *app) Render() {
	<div class="flex-col h-full bg-black">
		<div class="flex justify-between items-center px-1 py-0 shrink-0">
			<div class="flex items-center gap-1">
				<span class="font-bold text-magenta">{"⚡ TUI CHAT"}</span>
				<span class="font-dim text-cyan">v1.2.0</span>
				<span class="font-dim">•</span>
				<span class="font-dim">terminal direct messenger</span>
			</div>
			<div class="flex items-center gap-2">
				if a.view.Get() == viewChats {
					<span class="font-bold text-magenta">[ Direct Messages ]</span>
				} else {
					<span class="font-bold text-yellow">[ Profile & Auth ]</span>
				}
			</div>
		</div>
		<hr />
		<div class="flex grow min-h-0 gap-1 px-1">
			@Sidebar(a.client, a.selectedChat)
			if a.view.Get() == viewChats {
				@ChatPane(a.client, a.selectedChat, a.draft, a.onSend)
			} else {
				@Profile(a.name, a.username, a.password, a.profile().ID, a.profile().Token, a.profile().CreatedAt, a.profileEdit, a.saveProfile, a.cancelProfile)
			}
		</div>
		<hr />
		@StatusBar(a.view.Get(), a.profileEdit.Get())
		@HelpModal(a.showHelp)
		@NewDMModal(a.showNewDM, a.newDMUsername, a.startNewDM, a.cancelNewDM)
	</div>
}

func (a *app) onSend() {
	a.replyPending.Set(a.selectedChat.Get())
}

func (a *app) saveProfile() {
	p := a.client.Profile()
	p.Name = a.name.Get()
	p.Username = a.username.Get()
	p.Password = a.password.Get()
	a.client.UpdateProfile(p)
	a.profileEdit.Set(false)
}

func (a *app) cancelProfile() {
	p := a.client.Profile()
	a.name.Set(p.Name)
	a.username.Set(p.Username)
	a.password.Set(p.Password)
	a.profileEdit.Set(false)
}

func (a *app) startNewDM(username string) {
	username = strings.TrimSpace(strings.TrimPrefix(username, "@"))
	if username == "" {
		return
	}
	// Derive clean display name from username e.g. "alex" -> "Alex"
	displayName := strings.ToUpper(username[:1]) + username[1:]
	for _, ch := range a.client.Chats() {
		if strings.EqualFold(ch.Username, username) {
			// Select existing chat
			for i, c := range a.client.Chats() {
				if c.ID == ch.ID {
					a.selectedChat.Set(i)
					break
				}
			}
			a.newDMUsername.Set("")
			a.showNewDM.Set(false)
			a.view.Set(viewChats)
			return
		}
	}
	a.client.AddChat(fmt.Sprintf("%s", displayName), username)
	a.newDMUsername.Set("")
	a.showNewDM.Set(false)
	a.selectedChat.Set(len(a.client.Chats()) - 1)
	a.view.Set(viewChats)
}

func (a *app) cancelNewDM() {
	a.newDMUsername.Set("")
	a.showNewDM.Set(false)
}

var mockReplies = []string{
	"Sounds great! Looking into it right now.",
	"Checked the pull request, everything looks solid.",
	"Let me test the websocket endpoint locally.",
	"Agreed, nice improvement on the UI.",
	"Deploying the latest build to staging 🚀",
	"Thanks for the update!",
	"👍 Got it!",
	"Looks fantastic!",
}


