import type { User } from "$/lib/server/types";
import { proxyWithPersist } from "./persist";

type AuthStore = {
	user: User | null;
	state: "loading" | "loaded" | "error";
	setUser: (data: User) => void;
	isLogggedIn: () => boolean;
	clear: () => void;
};

const authStore = proxyWithPersist<AuthStore>("user", {
	defaultValue: {
		state: "loading",
		setUser(data: User) {
			authStore.user = data;
		},
		isLogggedIn() {
			return !!authStore.user?.id;
		},
		clear() {
			authStore.user = null;
		},
	} as AuthStore,
	excludeKeys: ["state"],
});

export default authStore;
