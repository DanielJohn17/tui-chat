package views

import (
	"fmt"
	"math/rand/v2"
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
	client       client.Client
	view         *tui.State[viewMode]
	selectedChat *tui.State[int]
	draft        *tui.State[string]
	username     *tui.State[string]
	accountCode  *tui.State[string]
	password     *tui.State[string]
	profileEdit  *tui.State[bool]
	replyPending *tui.State[int]
}

func App(c client.Client) *app {
	p := c.Profile()
	return &app{
		client:       c,
		view:         tui.NewState(viewChats),
		selectedChat: tui.NewState(0),
		draft:        tui.NewState(""),
		username:     tui.NewState(p.Username),
		accountCode:  tui.NewState(p.AccountCode),
		password:     tui.NewState(p.Password),
		profileEdit:  tui.NewState(false),
		replyPending: tui.NewState(-1),
	}
}

func (a *app) KeyMap() tui.KeyMap {
	km := tui.KeyMap{
		tui.OnStop(tui.KeyCtrlC, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.KeyTab, func(ke tui.KeyEvent) { ke.App().FocusNext() }),
		tui.OnStop(tui.KeyTab.Shift(), func(ke tui.KeyEvent) { ke.App().FocusPrev() }),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('p'), func(ke tui.KeyEvent) { a.view.Set(viewProfile); a.profileEdit.Set(false) }),
		tui.OnStop(tui.Rune('c'), func(ke tui.KeyEvent) { a.saveProfile(); a.view.Set(viewChats); a.profileEdit.Set(false) }),
	}
	if a.view.Get() == viewChats {
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
		km = append(km,
			tui.On(tui.KeyUp, moveUp),
			tui.On(tui.Rune('k'), moveUp),
			tui.On(tui.KeyDown, moveDown),
			tui.On(tui.Rune('j'), moveDown),
		)
	}
	if a.view.Get() == viewProfile && a.profileEdit.Get() {
	} else {
		km = append(km, tui.On(tui.KeyEscape, func(ke tui.KeyEvent) { a.view.Set(viewChats) }))
	}
	return km
}

func (a *app) Watchers() []tui.Watcher {
	return []tui.Watcher{
		tui.OnTimer(3*time.Second, func() {
			if a.replyPending.Get() >= 0 {
				chatID := a.replyPending.Get()
				a.client.Send(chatID, mockAutoReply())
				a.replyPending.Set(-1)
			}
		}),
	}
}

templ (a *app) Render() {
	<div class="flex-col h-full">
		<div class="flex justify-between items-center px-1 shrink-0">
			<div class="flex items-center gap-1">
				<span class="font-bold text-gradient-cyan-magenta">⚡ TUI CHAT</span>
				<span class="font-dim text-cyan">|</span>
				<span class="font-dim">terminal messenger</span>
			</div>
			<div class="flex items-center gap-1">
				<span class="text-green">● Online</span>
				<span class="font-dim">|</span>
				if a.view.Get() == viewChats {
					<span class="font-bold text-cyan">[ Chat ]</span>
				} else {
					<span class="font-bold text-magenta">[ Profile ]</span>
				}
			</div>
		</div>
		<hr />
		<div class="flex grow min-h-0">
			@Sidebar(a.client, a.selectedChat)
			if a.view.Get() == viewChats {
				@ChatPane(a.client, a.selectedChat, a.draft, a.onSend)
			} else {
				@Profile(a.username, a.accountCode, a.password, a.profileEdit, a.saveProfile, a.cancelProfile)
			}
		</div>
		<hr />
		@StatusBar(a.view.Get(), a.profileEdit.Get())
	</div>
}

func (a *app) onSend() {
	a.replyPending.Set(a.selectedChat.Get())
}

func (a *app) saveProfile() {
	a.client.UpdateProfile(client.Profile{
		Username:    a.username.Get(),
		AccountCode: a.accountCode.Get(),
		Password:    a.password.Get(),
	})
	a.profileEdit.Set(false)
}

func (a *app) cancelProfile() {
	p := a.client.Profile()
	a.username.Set(p.Username)
	a.accountCode.Set(p.AccountCode)
	a.password.Set(p.Password)
	a.profileEdit.Set(false)
}

var mockReplies = []string{
	"That's a good point!",
	"Let me think about that...",
	"I'll look into it.",
	"Sure, sounds good!",
	"Haha, nice one!",
	"Thanks for sharing!",
	"Interesting, I hadn't considered that.",
	"👍",
	"lol",
	"Agreed!",
}

var mockResponders = []string{"bob", "charlie", "system"}

func mockAutoReply() string {
	reply := mockReplies[rand.IntN(len(mockReplies))]
	sender := mockResponders[rand.IntN(len(mockResponders))]
	return fmt.Sprintf("%s: %s", sender, reply)
}

func viewName(v viewMode) string {
	if v == viewProfile {
		return "profile"
	}
	return "chat"
}
