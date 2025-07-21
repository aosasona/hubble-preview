import client from "$/lib/server";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/",
)({
	beforeLoad: ({ params }) => params,
	loader: ({ params }) => {
		return client.queries.workspaceMemberStatus({
			workspace_slug: params.workspaceSlug,
		});
	},
});
