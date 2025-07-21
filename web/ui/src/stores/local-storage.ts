class LocalStorage<Key extends string> {
	constructor(private prefix: string) {
		if (!prefix || prefix?.toString() === "") {
			throw new Error("Prefix cannot be an empty string");
		}

		this.prefix = prefix?.toLowerCase()?.trim();
	}

	private getKey(key: Key) {
		return `${this.prefix}:${key}`;
	}

	get<T>(key: Key, defaultValue?: T): T | null {
		if (typeof window === "undefined") return defaultValue || null;

		const value = localStorage.getItem(this.getKey(key));

		if (!value) return defaultValue || null;

		return JSON.parse(value);
	}

	set<T>(key: Key, value: T) {
		if (typeof window === "undefined") return;

		localStorage.setItem(this.getKey(key), JSON.stringify(value));
	}

	remove(key: Key) {
		if (typeof window === "undefined") return;

		localStorage.removeItem(this.getKey(key));
	}
}

export default LocalStorage;
