export type FileType =
	| "Image"
	| "Video"
	| "Audio"
	| "PDF"
	| "Document" // doc, docx
	| "Text"
	| "Presentation" // ppt, pptx
	| "Spreadsheet" // xls, xlsx, csv
	| "Markdown"
	| "HTML"
	| "JSON"
	| "ZIP"
	| "JavaScript"
	| "Unknown";
export function inferType(fileType: string): FileType {
	if (fileType.startsWith("image/")) {
		return "Image";
	}

	if (fileType.startsWith("video/")) {
		return "Video";
	}

	if (fileType.startsWith("audio/")) {
		return "Audio";
	}

	if (fileType === "application/pdf") {
		return "PDF";
	}

	if (fileType === "text/plain") {
		return "Text";
	}

	if (fileType === "text/markdown") {
		return "Markdown";
	}

	if (fileType === "text/html") {
		return "HTML";
	}

	if (
		fileType === "application/javascript" ||
		fileType === "application/x-javascript"
	) {
		return "JavaScript";
	}

	if (fileType === "application/json") {
		return "JSON";
	}

	if (fileType === "application/zip") {
		return "ZIP";
	}

	if (
		fileType ===
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	) {
		return "Document";
	}

	if (fileType === "application/msword") {
		return "Document";
	}

	if (
		fileType ===
		"application/vnd.openxmlformats-officedocument.presentationml.presentation"
	) {
		return "Presentation";
	}

	if (
		fileType ===
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	) {
		return "Spreadsheet";
	}

	if (fileType === "application/vnd.ms-excel") {
		return "Spreadsheet";
	}

	if (fileType === "text/csv") {
		return "Spreadsheet";
	}

	return "Unknown";
}
