import type { Filter, SortBy } from "$/components/entry/entries-list/filter";
import type { Entry } from "$/lib/server/types";
import { proxy } from "valtio";
import { proxySet } from "valtio/utils";

class EntriesListStore {
	selections: Set<string>;
	preview: Entry | null;
	filters: Array<Filter>;
	sortBy: SortBy;

	constructor() {
		this.selections = proxySet<string>([]);
		this.preview = null;
		this.filters = [];
		this.sortBy = "default";
	}

	isSelected(entryId: string): boolean {
		return this.selections.has(entryId);
	}

	select(entryId: string): void {
		this.selections.add(entryId);
	}
	deselect(entryId: string): void {
		this.selections.delete(entryId);
	}

	toggleSelection(entryId: string): void {
		if (this.isSelected(entryId)) {
			this.deselect(entryId);
			return;
		}
		this.select(entryId);
	}

	toggleFullSelection(entries: Array<Entry>): void {
		// If they are all selected, deselect all
		if (this.selections.size === entries.length) {
			this.selections.clear();
			return;
		}

		// Otherwise, select all
		for (const entry of entries) {
			if (!this.isSelected(entry.id)) {
				this.selections.add(entry.id);
			}
		}
	}

	clearSelections(): void {
		this.selections.clear();
	}

	updatePreview(entry: Entry): void {
		this.preview = entry;
	}
	clearPreview(): void {
		this.preview = null;
	}

	addFilter(filter: Filter): void {
		this.filters.push(filter);
	}

	removeFilter(filter: Filter): void {
		const index = this.filters.findIndex(
			(f) => f.type === filter.type && f.value === filter.value,
		);
		if (index === -1) {
			return;
		}

		this.filters.splice(index, 1);
	}

	clearFilters(): void {
		this.filters.splice(0, this.filters.length);
	}

	toggleFilter(filter: Filter): void {
		if (this.hasFilter(filter)) {
			this.removeFilter(filter);
			return;
		}

		this.addFilter(filter);
	}

	hasFilter(filter: Filter): boolean {
		return this.filters.some(
			(f) => f.type === filter.type && f.value === filter.value,
		);
	}

	applyFilter(filter: Filter, entry: Entry): boolean {
		// Multiple filters of the same type are OR'd
		const types = this.filters.filter((f) => f.type === "type");
		const statuses = this.filters.filter((f) => f.type === "status");
		const collections = this.filters.filter((f) => f.type === "collection");

		switch (filter.type) {
			case "type":
				return types.some((f) => f.value === entry.type);
			case "status":
				return statuses.some((f) => f.value === entry.status);
			case "collection":
				return collections.some((f) => f.value === entry.collection.id);
		}

		return false;
	}

	applyFilters(entry: Entry): boolean {
		if (this.filters.length === 0) {
			return true;
		}

		for (const filter of this.filters) {
			if (!this.applyFilter(filter, entry)) {
				return false;
			}
		}

		return true;
	}

	setSortBy(sortBy: SortBy): void {
		this.sortBy = sortBy;
	}

	applySortingOrder(a: Entry, b: Entry): number {
		switch (this.sortBy) {
			case "newest":
				return (
					new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
				);
			case "oldest":
				return (
					new Date(a.created_at).getTime() - new Date(b.created_at).getTime()
				);
			case "name_asc":
				return a.name.localeCompare(b.name);
			case "name_desc":
				return b.name.localeCompare(a.name);
			case "collection_asc":
				return a.collection.slug.localeCompare(b.collection.slug);
			case "collection_desc":
				return b.collection.slug.localeCompare(a.collection.slug);
			case "default":
				return 0;
		}
	}

	clear(): void {
		this.clearSelections();
		this.clearFilters();
		this.clearPreview();
		this.setSortBy("default");
	}
}

const entriesListStore = proxy(new EntriesListStore());

export default entriesListStore;
