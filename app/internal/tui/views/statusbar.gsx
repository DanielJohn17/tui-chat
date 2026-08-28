package views

import tui "github.com/grindlemire/go-tui"

templ StatusBar(v viewMode, editing bool) {
	if v == viewChats {
		<div class="flex shrink-0 bg-blue text-white w-full px-1">
			<span class="font-bold">↑/↓ j/k</span>
			<span class="font-dim">nav</span>
			<span class="font-dim">|</span>
			<span class="font-bold">Tab</span>
			<span class="font-dim"></span>
			<span class="font-dim">|</span>
			<span class="font-bold">p</span>
			<span class="font-dim">profile</span>
			<span class="font-dim">|</span>
			<span class="font-bold">Enter</span>
			<span class="font-dim">send</span>
			<span class="font-dim">|</span>
			<span class="font-bold">q</span>
			<span class="font-dim">quit</span>
			<span class="grow"></span>
			<span class="font-dim">● connected</span>
			<span class="font-dim">|</span>
			<span class="font-bold">tui-chat</span>
		</div>
	} else if editing {
		<div class="flex shrink-0 bg-green text-black w-full px-1">
			<span class="font-bold">EDITING PROFILE</span>
			<span class="font-dim px-1">|</span>
			<span class="font-bold">Tab</span>
			<span>next field</span>
			<span class="font-dim">|</span>
			<span class="font-bold">Enter</span>
			<span>save</span>
			<span class="font-dim">|</span>
			<span class="font-bold">Esc</span>
			<span>cancel</span>
			<span class="grow"></span>
			<span class="font-bold">tui-chat</span>
		</div>
	} else {
		<div class="flex shrink-0 bg-blue text-white w-full px-1">
			<span class="font-bold">c</span>
			<span class="font-dim">chats</span>
			<span class="font-dim">|</span>
			<span class="font-bold">e</span>
			<span class="font-dim">edit profile</span>
			<span class="font-dim">|</span>
			<span class="font-bold">Esc</span>
			<span class="font-dim">back</span>
			<span class="font-dim">|</span>
			<span class="font-bold">q</span>
			<span class="font-dim">quit</span>
			<span class="grow"></span>
			<span class="font-bold">tui-chat</span>
		</div>
	}
}
