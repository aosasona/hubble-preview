export function isUrl(text: string) {
	try {
		new URL(text);
		return true;
	} catch {
		return false;
	}
}
