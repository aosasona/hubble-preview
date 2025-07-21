import { createFileRoute } from "@tanstack/react-router";
import * as v from "valibot";

const searchSchema = v.object({
	email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
	scope: v.optional(v.union([v.literal("reset"), v.literal("change")])),
	token: v.optional(
		v.pipe(v.string(), v.regex(/^[a-zA-Z0-9_]{8}$/, "Invalid token.")),
	),
});
export const Route = createFileRoute("/auth/change-password")({
	validateSearch: searchSchema,
});
