export const MixedNameRegex = /^[a-zA-Z0-9]([a-zA-Z0-9-_ ]+)?$/;

export function slugify(text: string) {
	return text
		.toLowerCase()
		.replace(/\s+/g, "-")
		.replace(/[^\w-]+/g, "")
		.replace(/--+/g, "-")
		.replace(/^-+/, "")
		.replace(/-+$/, "");
}

export function toTitleCase(text: string) {
	return text
		.trim()
		?.split(" ")
		.map((word) => word.charAt(0).toUpperCase() + word.slice(1)?.toLowerCase())
		.join(" ");
}

export function pluralize(text: string, count: number) {
	if (count === 1) return text;

	// Account for the special cases
	if (text.endsWith("child")) {
		return count === 1 ? text : "children";
	}
	if (text.endsWith("person")) {
		return count === 1 ? text : "people";
	}
	if (text.endsWith("tooth")) {
		return count === 1 ? text : "teeth";
	}
	if (text.endsWith("foot")) {
		return count === 1 ? text : "feet";
	}
	if (text.endsWith("mouse")) {
		return count === 1 ? text : "mice";
	}

	if (text.endsWith("s")) {
		return `${text}es`;
	}

	if (text.endsWith("y")) {
		return `${text.slice(0, -1)}ies`;
	}

	if (text.endsWith("o")) {
		return `${text}es`;
	}

	if (text.endsWith("ch") || text.endsWith("sh")) {
		return `${text}es`;
	}

	if (text.endsWith("f")) {
		return `${text.slice(0, -1)}ves`;
	}

	if (text.endsWith("fe")) {
		return `${text.slice(0, -2)}ves`;
	}

	return `${text}s`;
}

/**
 * @description Redact parts of an email address, leaving only the first `minLength` characters visible.
 */
export function redactEmail(email?: string, minLength = 3) {
	if (!email) return "";

	const [local, domain] = email.split("@");
	const redactedLocal = local
		.split("")
		.map((char, index) => (index < minLength ? char : "*"))
		.join("");

	return `${redactedLocal}@${domain}`;
}

export function toHumanReadableSize(byteSize: number) {
	let size = byteSize;
	const units = ["B", "KB", "MB", "GB", "TB"];
	let unitIndex = 0;

	while (size >= 1024) {
		size /= 1024;
		unitIndex++;
	}

	return `${size.toFixed(2)} ${units[unitIndex]}`;
}
