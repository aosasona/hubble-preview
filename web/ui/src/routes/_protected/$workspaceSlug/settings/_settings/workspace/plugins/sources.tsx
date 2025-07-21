import ErrorComponent from "$/components/route/error-component";
import client from "$/lib/server";
import { createFileRoute } from "@tanstack/react-router";
import * as v from "valibot";

const paramsSchema = v.object({
	page: v.optional(v.number(), 1),
});

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/plugins/sources",
)({
	validateSearch: paramsSchema,
	beforeLoad: ({ params }) => params,
	loader: async ({ params }) => {
		const workspaceWithMembershipStatus =
			await client.queries.workspaceMemberStatus({
				workspace_slug: params.workspaceSlug,
			});

		return {
			workspace: workspaceWithMembershipStatus.workspace,
			status: workspaceWithMembershipStatus.status,
		};
	},
	errorComponent: ErrorComponent,
});
