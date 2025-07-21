import stores from "$/stores";
import { toast } from "sonner";
import type { Entry } from "./server/types";

export type GetEntryFn = (index: string) => Entry | undefined;

function updatePreview(element: HTMLElement, fn: GetEntryFn) {
	if (!stores.entriesList.preview || !element) return;

	const index = element.getAttribute("data-index");
	if (!index) return;

	const entry = fn(index);
	if (!entry) return;

	stores.entriesList.updatePreview(entry);
}

export function moveFocusDown(fn: GetEntryFn) {
	// Move focus to the next entry
	const next = document.activeElement?.nextElementSibling as HTMLElement;
	if (!next) return;

	next?.focus();

	// Update preview if it's open
	updatePreview(next, fn);
}

export function moveFocusUp(fn: GetEntryFn) {
	// Move focus to the previous entry
	const prev = document.activeElement?.previousElementSibling as HTMLElement;
	if (!prev) return;

	prev?.focus();

	// Update preview if it's open
	updatePreview(prev, fn);
}

export function selectRow(fn: GetEntryFn) {
	const index = document.activeElement?.getAttribute("data-index");
	if (!index) return;

	const entry = fn(index);
	if (!entry) return;

	stores.entriesList.toggleSelection(entry.id);
}

export function togglePreview(e: KeyboardEvent, fn: GetEntryFn) {
	// Prevent space from scrolling the page
	e.preventDefault();
	e.stopPropagation();

	// If preview is open, close it
	if (stores.entriesList.preview) {
		stores.entriesList.clearPreview();
		return;
	}

	// Otherwise, open the preview
	const entry = fn(document.activeElement?.getAttribute("data-index") ?? "");
	if (!entry) return;

	stores.entriesList.updatePreview(entry);
}

export function toggleSelectAll(entries: Entry[]) {
	stores.entriesList.toggleFullSelection(entries);
}

export function selectEntry(entryId: string) {
	stores.entriesList.select(entryId);
}

export function deselectEntry(entryId: string) {
	stores.entriesList.deselect(entryId);
}

export function deleteSelectedEntries(fn: GetEntryFn) {
	if (stores.entriesList.selections.size > 0) {
		stores.app.openDialog("deleteSelectedEntries");
	} else {
		// Select the current entry
		const index = document.activeElement?.getAttribute("data-index");
		if (!index) return;

		const entry = fn(index);
		if (!entry) return;

		stores.entriesList.select(entry.id);
		stores.app.openDialog("deleteSelectedEntries");
	}
}

export function copySelectedEntriesLinks(fn: GetEntryFn) {
	const links = [];

	if (stores.entriesList.selections.size > 0) {
		const selectedLinks = Array.from(stores.entriesList.selections.values())
			.map((id) => {
				const entry = fn(id);
				if (!entry) return "";

				return makeEntryUrl(entry);
			})
			.filter(Boolean);

		if (selectedLinks.length === 0) return;
		links.push(...selectedLinks);
	} else {
		// Select the current entry
		const index = document.activeElement?.getAttribute("data-index");
		if (!index) return;

		const entry = fn(index);
		if (!entry) return;

		links.push(makeEntryUrl(entry));
	}

	if (links.length > 0) {
		navigator.clipboard.writeText(links.join("\n")).then(() => {
			toast.success(
				`Copied ${links.length} link${links.length > 1 ? "s" : ""} to clipboard`,
			);
		});
	}
}

function makeEntryUrl(entry: Entry) {
	const scheme = window.location.protocol;
	const host = window.location.host;
	return `${entry.name}: ${scheme}//${host}/${entry.workspace.slug}/c/${entry.collection.slug}/${entry.id}`;
}
