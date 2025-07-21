import { createFileRoute } from "@tanstack/react-router";
import { requireNoAuth } from "$/lib/auth";
import * as v from "valibot";

const SignInSchema = v.object({
	email: v.optional(v.pipe(v.string(), v.email())),
	redirect: v.optional(v.string()), // This is used to redirect the user to the page they were trying to access
});

export const Route = createFileRoute("/auth/sign-in")({
	validateSearch: SignInSchema,
	beforeLoad: async () => {
		await requireNoAuth();
	},
});
