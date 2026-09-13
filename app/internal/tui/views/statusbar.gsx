package views

import tui "github.com/grindlemire/go-tui"

templ StatusBar(v viewMode, editing bool) {
	if v == viewChats {
		<div class="flex shrink-0 bg-black text-white w-full px-1 py-0 items-center gap-1">
			<span class="text-magenta font-bold">q</span>
			<span class="text-white">Quit</span>
			<span class="font-dim">•</span>
			<span class="text-magenta font-bold">n</span>
			<span class="text-white">New DM</span>
			<span class="font-dim">•</span>
			<span class="text-magenta font-bold">p</span>
			<span class="text-white">Profile</span>
			<span class="font-dim">•</span>
			<span class="text-cyan font-bold">Tab</span>
			<span class="text-white">Type</span>
			<span class="font-dim">•</span>
			<span class="text-cyan font-bold">Enter</span>
			<span class="text-white">Send</span>
			<span class="font-dim">•</span>
			<span class="text-yellow font-bold">j/k</span>
			<span class="text-white">Navigate</span>
			<span class="grow"></span>
			<span class="text-magenta font-bold">tui-chat</span>
		</div>
	} else if v == viewNewDM {
		<div class="flex shrink-0 bg-black text-white w-full px-1 py-0 items-center gap-1">
			<span class="text-magenta font-bold">[NEW CONVERSATION]</span>
			<span class="font-dim">•</span>
			<span class="text-cyan font-bold">Tab</span>
			<span class="text-white">Next Field</span>
			<span class="font-dim">•</span>
			<span class="text-green font-bold">Enter</span>
			<span class="text-white">Start</span>
			<span class="font-dim">•</span>
			<span class="text-yellow font-bold">Esc</span>
			<span class="text-white">Cancel</span>
			<span class="grow"></span>
			<span class="text-magenta font-bold">tui-chat</span>
		</div>
	} else if editing {
		<div class="flex shrink-0 bg-black text-white w-full px-1 py-0 items-center gap-1">
			<span class="text-yellow font-bold">[EDIT PROFILE]</span>
			<span class="font-dim">•</span>
			<span class="text-cyan font-bold">Tab</span>
			<span class="text-white">Next Field</span>
			<span class="font-dim">•</span>
			<span class="text-green font-bold">Enter</span>
			<span class="text-white">Save</span>
			<span class="font-dim">•</span>
			<span class="text-magenta font-bold">Esc</span>
			<span class="text-white">Cancel</span>
			<span class="grow"></span>
			<span class="text-yellow font-bold">tui-chat</span>
		</div>
	} else {
		<div class="flex shrink-0 bg-black text-white w-full px-1 py-0 items-center gap-1">
			<span class="text-cyan font-bold">c / Esc</span>
			<span class="text-white">Chats</span>
			<span class="font-dim">•</span>
			<span class="text-magenta font-bold">e</span>
			<span class="text-white">Edit Profile</span>
			<span class="font-dim">•</span>
			<span class="text-yellow font-bold">q</span>
			<span class="text-white">Quit</span>
			<span class="grow"></span>
			<span class="text-magenta font-bold">tui-chat</span>
		</div>
	}
}


