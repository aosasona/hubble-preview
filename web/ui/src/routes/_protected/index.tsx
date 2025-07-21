import { requireAuth } from "$/lib/auth";
import { beforeWorkspaceLoad } from "$/lib/workspace";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_protected/")({
	component: RouteComponent,
	beforeLoad: async ({ location }) => {
		await requireAuth(location);
		beforeWorkspaceLoad();
	},
});

function RouteComponent() {
	return <Outlet />;
}
