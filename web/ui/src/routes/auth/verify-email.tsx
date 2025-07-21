import { createFileRoute } from "@tanstack/react-router";
import * as v from "valibot";
import { requireNoAuth } from "$/lib/auth";

const verifyEmailSchema = v.object({
	email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
	token: v.optional(
		v.pipe(v.string(), v.regex(/^[a-zA-Z0-9_]{8}$/, "Invalid token.")),
	),
});

export const Route = createFileRoute("/auth/verify-email")({
	validateSearch: verifyEmailSchema,
	beforeLoad: async () => {
		await requireNoAuth();
	},
});
