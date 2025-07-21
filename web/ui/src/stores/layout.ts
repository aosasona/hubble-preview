import type { AccentColor } from "$/stores/app";
import { Gear, type Icon } from "@phosphor-icons/react";
import type { JSX } from "react";
import { proxy } from "valtio";

export type HeaderItem =
	| {
			// The icon associated with the header item
			icon?: Icon;
			// The color of the icon
			color?: AccentColor;
			// The title of the header item
			title: string;
			// The URL to navigate to when the header item is clicked
			url?: string;
	  }
	| { title: string; component: () => JSX.Element };

export type Parent = "settings";

interface LayoutStore {
	// The header items to display
	headerItems: HeaderItem[];

	// Whether to show the header in the body
	fullScreen?: boolean;

	// Add a header item
	appendHeaderItem(item: HeaderItem): void;

	appendItems(items: HeaderItem[]): void;

	// Remove a header item
	setHeaderItems(items: HeaderItem[]): void;

	// Set the current header item
	setCurrentHeaderItem(item: HeaderItem): void;

	// Clear all header items
	clearHeaderItems(): void;

	// Clear all the items except the first one (i.e. the parent)
	clearChildrenItems(): void;

	// Check if there are any header items
	hasHeaderItems(): boolean;

	setParent(parent: Parent): void;
}

const layoutStore: LayoutStore = proxy<LayoutStore>({
	headerItems: [],

	appendHeaderItem(item) {
		layoutStore.headerItems.push(item);
	},

	appendItems(items) {
		layoutStore.headerItems.push(...items);
	},

	setHeaderItems(items) {
		layoutStore.headerItems = items;
	},

	setCurrentHeaderItem(item) {
		layoutStore.headerItems = [item];
	},

	clearHeaderItems() {
		layoutStore.headerItems = [];
	},

	clearChildrenItems() {
		layoutStore.headerItems = layoutStore.headerItems.slice(0, 1);
	},

	hasHeaderItems() {
		return layoutStore.headerItems.length > 0;
	},

	setParent(parent: Parent) {
		switch (parent) {
			case "settings":
				layoutStore.setHeaderItems([
					{ title: "Settings", icon: Gear, color: "gray" },
				]);
				break;
		}
	},
});

export default layoutStore;
