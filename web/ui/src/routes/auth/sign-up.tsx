import { createFileRoute } from "@tanstack/react-router";
import { requireNoAuth } from "$/lib/auth";

export const Route = createFileRoute("/auth/sign-up")({
	beforeLoad: async () => {
		await requireNoAuth();
	},
});
