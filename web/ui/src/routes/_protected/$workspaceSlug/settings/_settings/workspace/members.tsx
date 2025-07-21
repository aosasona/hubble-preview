import client from "$/lib/server";
import { createFileRoute } from "@tanstack/react-router";
import * as v from "valibot";

const paramsSchema = v.object({
	page: v.optional(v.number(), 1),
	search: v.optional(v.string()),
});

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/members",
)({
	validateSearch: paramsSchema,
	beforeLoad: ({ params }) => params,
	loader: ({ params }) => {
		return client.queries.workspaceMemberStatus({
			workspace_slug: params.workspaceSlug,
		});
	},
});
