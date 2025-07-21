import { createFileRoute } from "@tanstack/react-router";
import { requireNoAuth } from "$/lib/auth";

export const Route = createFileRoute("/auth/forgot-password")({
	beforeLoad: async () => {
		await requireNoAuth();
	},
});
