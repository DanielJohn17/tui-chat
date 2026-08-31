package views

import tui "github.com/grindlemire/go-tui"

templ StatusBar(v viewMode, editing bool) {
	if v == viewChats {
		<div class="flex shrink-0 bg-black text-white w-full px-1 items-center gap-1">
			<span class="text-cyan font-bold">{"\u2191/\u2193 j/k"}</span>
			<span class="text-white">nav</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">Tab</span>
			<span class="text-white">type</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">p</span>
			<span class="text-white">profile</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">Enter</span>
			<span class="text-white">send</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">q</span>
			<span class="text-white">quit</span>
			<span class="grow"></span>
			<span class="text-green font-bold">{"\u25cf"}</span>
			<span class="font-dim">connected</span>
			<span class="font-dim">|</span>
			<span class="font-bold text-cyan">tui-chat</span>
		</div>
	} else if editing {
		<div class="flex shrink-0 bg-black text-white w-full px-1 items-center gap-1">
			<span class="text-yellow font-bold">[EDITING PROFILE]</span>
			<span class="font-dim">|</span>
			<span class="text-yellow font-bold">Tab</span>
			<span class="text-white">next field</span>
			<span class="font-dim">|</span>
			<span class="text-green font-bold">Enter</span>
			<span class="text-white">save</span>
			<span class="font-dim">|</span>
			<span class="text-red font-bold">Esc</span>
			<span class="text-white">cancel</span>
			<span class="grow"></span>
			<span class="font-bold text-yellow">tui-chat</span>
		</div>
	} else {
		<div class="flex shrink-0 bg-black text-white w-full px-1 items-center gap-1">
			<span class="text-cyan font-bold">c</span>
			<span class="text-white">chats</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">e</span>
			<span class="text-white">edit profile</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">Esc</span>
			<span class="text-white">back</span>
			<span class="font-dim">|</span>
			<span class="text-cyan font-bold">q</span>
			<span class="text-white">quit</span>
			<span class="grow"></span>
			<span class="font-bold text-cyan">tui-chat</span>
		</div>
	}
}
