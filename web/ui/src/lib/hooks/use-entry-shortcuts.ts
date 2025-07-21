import {
	copySelectedEntriesLinks,
	deleteSelectedEntries,
	moveFocusDown,
	moveFocusUp,
	selectRow,
	togglePreview,
	toggleSelectAll,
	type GetEntryFn,
} from "$/lib/entry-shortcuts";
import type { Entry } from "../server/types";
import useShortcut from "./use-shortcut";

export default function useEntriesListShortcuts(
	fn: GetEntryFn,
	entries: Entry[],
) {
	// Movements
	useShortcut(["j", "down"], () => moveFocusDown(fn), {
		scopes: ["entries"],
	});
	useShortcut(["k", "up"], () => moveFocusUp(fn), {
		scopes: ["entries"],
	});

	// Handle selections
	useShortcut(["x"], () => selectRow(fn), { scopes: ["entries"] });
	useShortcut(["shift+x"], () => toggleSelectAll(entries), {
		scopes: ["entries"],
	});
	useShortcut(["space"], (e) => togglePreview(e, fn), {
		scopes: ["entries"],
	});
	useShortcut(["d"], () => deleteSelectedEntries(fn), {
		scopes: ["entries"],
	});
	useShortcut(["y"], () => copySelectedEntriesLinks(fn), {
		scopes: ["entries"],
	});
}
