import { proxy, subscribe } from "valtio";
import LocalStorage from "./local-storage";

type StoreName = "app" | "user" | "workspace";

type Options<T extends object> = {
	defaultValue?: T;
	excludeKeys?: (keyof T)[];
};

const localStore = new LocalStorage<StoreName>("hubble.store");

export function proxyWithPersist<T extends object>(
	name: StoreName,
	options: Options<T> = {
		defaultValue: {} as T,
	},
): T {
	const store = proxy<T>(options.defaultValue);

	const data = localStore.get<T>(name);
	for (const key in data) {
		const value = data[key as keyof T];
		if (typeof value === "undefined") {
			// Skip undefined values
			continue;
		}

		store[key as keyof T] = value;
	}

	saveOnChange(name, store, options);

	return store;
}

export function saveOnChange<T extends object>(
	name: StoreName,
	store: T,
	options?: Options<T>,
) {
	const unsubscribe = subscribe(store, (ops) => {
		// Check if the op includes our ignored keys or functions
		const isIgnored = ops.some((op) => {
			const key = (op[1]?.toString() ?? "") as keyof T;
			const value = op[2];

			return (
				options?.excludeKeys?.includes(key as keyof T) ||
				typeof value === "function"
			);
		});

		if (isIgnored) {
			return;
		}

		const persistedObject: T = {} as T;

		// Only persist the keys that are not functions
		for (const key in store) {
			if (
				options?.excludeKeys?.includes(key as keyof T) ||
				typeof store[key as keyof T] === "function"
			) {
				continue;
			}

			persistedObject[key as keyof T] = store[key as keyof T];
		}

		localStore.set(name, persistedObject);
	});

	window.addEventListener("beforeunload", () => {
		unsubscribe();
	});
}
