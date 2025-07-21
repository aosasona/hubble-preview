import stores from "$/stores";
import { createFileRoute, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/c/$collectionSlug",
)({
	component: RouteComponent,
	beforeLoad: ({ params }) => {
		// Set the active collection
		stores.workspace.setActiveCollection({ slug: params.collectionSlug });
	},
});

function RouteComponent() {
	return <Outlet />;
}
