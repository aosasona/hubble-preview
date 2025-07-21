import {
	type HotkeyCallback,
	type Options as HotKeyOptions,
	useHotkeys,
} from "react-hotkeys-hook";

type ModifierKeys = "ctrl" | "mod" | "shift" | "alt" | "meta";
type Key =
	| "a"
	| "b"
	| "c"
	| "d"
	| "e"
	| "f"
	| "g"
	| "h"
	| "i"
	| "j"
	| "k"
	| "l"
	| "m"
	| "n"
	| "o"
	| "p"
	| "q"
	| "r"
	| "s"
	| "t"
	| "u"
	| "v"
	| "w"
	| "x"
	| "y"
	| "z"
	| "0"
	| "1"
	| "2"
	| "3"
	| "4"
	| "5"
	| "6"
	| "7"
	| "8"
	| "9"
	| "/"
	| "."
	| ","
	| " "
	| "enter"
	| "space"
	| "backspace"
	| "up"
	| "down"
	| "left"
	| "right"
	| "tab"
	| "del"
	| "esc";

type HotKey =
	| `${ModifierKeys}+${Key}`
	| `${ModifierKeys}+${ModifierKeys}+${Key}`
	| `${ModifierKeys}+${Key}+${Key}`
	| Key;

type Scope = "global" | "entries" | "import-entry" | "sources" | "plugins";

export default function useShortcut(
	keys: HotKey[],
	action: HotkeyCallback,
	options: HotKeyOptions & { scopes?: Scope[] } = {},
) {
	useHotkeys(keys, action, options);
}
