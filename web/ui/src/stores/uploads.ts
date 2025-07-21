import { proxy, ref } from "valtio";
import { slugify } from "$/lib/utils";
import { isUrl } from "$/lib/url";
import client from "$/lib/server";
import { extractError } from "$/lib/error";
import type { Metadata } from "$/lib/server/types";

type Status = {
	/** The status of the meta lookup or any other op */
	status: "loading" | "success" | "error";
	error?: string;
};

export type FileEntry = {
	type: "file";
	name: string;
	tags: string[];
} & Status;

export type LinkEntry = {
	type: "link";
} & Metadata &
	Status;

export type NewEntry = FileEntry | LinkEntry;

type UploadStore = {
	entries: Record<string, NewEntry>;
	files: File[]; // NOTE: this is necessary because we don't want to pass around proxies to Blob types, so we use a valtio `ref` instead
	state: "idle" | "loading" | "success" | "error";

	generateId(entry: NewEntry): string;
	add: (entry: NewEntry) => string | null;
	addFile: (file: File) => void;
	update: (id: string, entry: Partial<NewEntry>) => void;
	updateProperty: <K extends keyof NewEntry>(
		id: string,
		key: K,
		value: NewEntry[K],
	) => void;
	remove: (id: string) => void;
	addLink: (url: string) => void;
	reset: () => void;
};

const uploadsStore = proxy<UploadStore>({
	entries: {},
	files: ref([]),
	state: "idle",
	generateId(entry): string {
		switch (entry.type) {
			case "file":
				return slugify(entry.name);
			case "link":
				return slugify(entry.link);
		}
	},
	addFile(file) {
		if (uploadsStore.state === "loading") {
			return;
		}

		// Ensure the file is not already in the list
		if (uploadsStore.files.some((f) => f.name === file.name)) {
			return;
		}

		uploadsStore.files.push(file);
	},
	add(entry): string | null {
		if (uploadsStore.state === "loading") {
			return null;
		}

		const id = uploadsStore.generateId(entry);
		uploadsStore.entries[id] = entry;
		return id;
	},
	update(id, entry) {
		uploadsStore.entries[id] = {
			...uploadsStore.entries[id],
			...entry,
		} as NewEntry;
	},
	updateProperty(id, key, value) {
		uploadsStore.update(id, { [key]: value });
	},
	remove(id) {
		delete uploadsStore.entries[id];
	},
	addLink(link) {
		if (!isUrl(link)) return;

		const id = uploadsStore.add({
			type: "link",
			title: "",
			description: "",
			favicon: "",
			author: "",
			thumbnail: "",
			site_type: "",
			domain: "",
			link,
			status: "loading",
		});

		if (!id) return;

		// Load the metadata for the link
		client.queries
			.getLinkMetadata(link)
			.then((data) => {
				uploadsStore.update(id, { status: "success", type: "link", ...data });
			})
			.catch((error) => {
				const err = extractError(error);
				uploadsStore.update(id, { status: "error", error: err?.message ?? "" });
			});
	},
	reset() {
		if (uploadsStore.state === "loading") {
			return;
		}

		uploadsStore.entries = {};
		uploadsStore.files.splice(0, uploadsStore.files.length);
		uploadsStore.state = "idle";
	},
});

export default uploadsStore;
