import stores from "$/stores";
import { type ParsedLocation, redirect } from "@tanstack/react-router";
import type { User } from "./server/types";

async function ensureStoreIsLoaded(): Promise<User | null> {
	let user = stores.auth.user;
	if (!user?.id) {
		const data = await stores.load();
		user = data?.user || null;
	}

	return user;
}

export async function requireAuth(location: ParsedLocation) {
	const user = await ensureStoreIsLoaded();

	// If the user is still not set, redirect to login with the current page attached
	if (!user) {
		throw redirect({
			to: "/auth/sign-in",
			search: { redirect: location?.pathname ?? "/" },
			replace: true,
		});
	}
}

export async function requireNoAuth() {
	const user = await ensureStoreIsLoaded();

	// If the user is set, redirect to the dashboard
	if (user) {
		throw redirect({ to: "/", replace: true });
	}
}
