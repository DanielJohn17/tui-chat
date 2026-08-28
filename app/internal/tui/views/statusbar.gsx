package views

import tui "github.com/grindlemire/go-tui"

templ StatusBar(v viewMode, editing bool) {
	if v == viewChats {
		<div class="shrink-0 bg-blue text-white px-1">
			<span class="font-bold px-1">⌨  ↑/↓ j/k nav</span>
			<span class="px-1">•</span>
			<span class="font-bold">Tab type</span>
			<span class="px-1">•</span>
			<span class="font-bold">p profile</span>
			<span class="px-1">•</span>
			<span class="font-bold">Enter send</span>
			<span class="px-1">•</span>
			<span class="font-bold">q quit</span>
		</div>
	} else if editing {
		<div class="shrink-0 bg-green text-black px-1">
			<span class="font-bold px-1">✎ Editing Profile</span>
			<span class="px-1">•</span>
			<span>Tab field</span>
			<span class="px-1">•</span>
			<span>Enter save</span>
			<span class="px-1">•</span>
			<span>Esc cancel</span>
		</div>
	} else {
		<div class="shrink-0 bg-blue text-white px-1">
			<span class="font-bold px-1">⌨  c chats</span>
			<span class="px-1">•</span>
			<span class="font-bold">e edit</span>
			<span class="px-1">•</span>
			<span class="font-bold">esc back</span>
			<span class="px-1">•</span>
			<span class="font-bold">q quit</span>
		</div>
	}
}