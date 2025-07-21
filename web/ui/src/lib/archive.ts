import JSZip from "jszip";
import type { Entry } from "./server/types";
import { slugify } from "./utils";

type ExportedEntryJSON = {
	id: string;
	name: string;
	version: number;
	type: string;
	status: string;
	created_at: string;
	added_by: string;
	collection_id: string;
	workspace_id: string;
	metadata: Record<string, unknown>;
	filesize_bytes?: number;
	file_id?: string | null;
};

export function exportEntry(entry: Entry) {
	const plaintext = entry.text_content;
	const markdown = entry.content;

	const metadata: ExportedEntryJSON = {
		id: entry.id,
		name: entry.name,
		version: entry.version,
		type: entry.type,
		status: entry.status,
		created_at: entry.created_at,
		added_by: entry.added_by.username,
		collection_id: entry.collection.id,
		workspace_id: entry.workspace.id,
		metadata: entry.metadata,
	};

	if (entry.type !== "link") {
		metadata.filesize_bytes = entry.filesize_bytes;
		metadata.file_id = entry.file_id ?? null;
	}

	const zip = new JSZip();

	zip.file("metadata.json", JSON.stringify(metadata, null, 2));
	zip.file("content.txt", plaintext);
	zip.file("content.md", markdown);

	zip.generateAsync({ type: "blob" }).then((content) => {
		const url = URL.createObjectURL(content);
		const a = document.createElement("a");
		a.href = url;
		a.download = `${slugify(entry.name)}-${entry.id}.zip`;
		a.click();
		URL.revokeObjectURL(url);
	});
}
