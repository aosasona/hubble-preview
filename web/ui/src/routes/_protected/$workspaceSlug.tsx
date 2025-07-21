import { useEffect } from "react";
import { useSnapshot } from "valtio";
import { useRobinQuery } from "$/lib/hooks";
import queryClient from "$/lib/query-client";
import QueryKeys from "$/lib/keys";
import SplashScreen from "$/components/splash-screen";
import stores from "$/stores";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { HotkeysProvider } from "react-hotkeys-hook";
import ImportEntryDialog from "$/components/entry/import-entry";
import ErrorPage from "$/components/error-page";
import client from "$/lib/server";
import DeleteEntriesDialog from "$/components/entry/entries-list/delete-dialog";

export const Route = createFileRoute("/_protected/$workspaceSlug")({
	component: RouteComponent,
	beforeLoad: ({ params }) => {
		stores.workspace.removeActiveCollection();
		stores.entriesList.clear();
		return params;
	},
	loader: async ({ params }) => {
		// Prefetch the current workspace
		return await queryClient.prefetchQuery({
			queryKey: QueryKeys.FindWorkspace(params.workspaceSlug),
			queryFn: () => {
				return client.queries.workspaceFind({ slug: params.workspaceSlug });
			},
		});
	},
	staleTime: 10_000, // 10 seconds
	errorComponent: ErrorPage,
	pendingComponent: SplashScreen,
});

function RouteComponent() {
	const { workspaceSlug: slug } = Route.useParams();

	const workspace = useSnapshot(stores.workspace);
	const app = useSnapshot(stores.app);

	const query = useRobinQuery(
		"workspace.find",
		{ slug },
		{
			queryKey: QueryKeys.FindWorkspace(slug),
			retry: false,
		},
	);

	useEffect(() => {
		if (query.isFetched && query.data) {
			workspace.setActiveWorkspace(
				query.data.workspace,
				query.data.collections,
			);
		}
	}, [query.data, query.isFetched, workspace]);

	if (query.isError && query.isFetched) {
		return <ErrorPage error={query.error as Error} />;
	}

	return (
		<>
			<Outlet />

			<HotkeysProvider initiallyActiveScopes={["import-entry"]}>
				<ImportEntryDialog
					open={app.dialogs.importItem}
					onOpen={() => app.openDialog("importItem")}
					onClose={() => app.closeDialog("importItem")}
				/>

				{/* MARK: Delete dialog */}
				<DeleteEntriesDialog />
			</HotkeysProvider>
		</>
	);
}
