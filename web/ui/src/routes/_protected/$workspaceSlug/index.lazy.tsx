import PageLayout from "$/components/layout/page-layout";
import client from "$/lib/server";
import stores from "$/stores";
import { useInfiniteQuery } from "@tanstack/react-query";
import { createLazyFileRoute } from "@tanstack/react-router";
import { useSnapshot } from "valtio";
import PageSpinner from "$/components/page-spinner";
import ErrorPage from "$/components/error-page";
import QueryKeys from "$/lib/keys";
import EntriesList, {
	type FetcherFnResult,
} from "$/components/entry/entries-list";
import { HotkeysProvider } from "react-hotkeys-hook";

export const Route = createLazyFileRoute("/_protected/$workspaceSlug/")({
	component: RouteComponent,
});

const DEFAULT_LIMIT = 50;

function RouteComponent() {
	const params = Route.useParams();
	const workspace = useSnapshot(stores.workspace);

	const query = useInfiniteQuery({
		queryKey: QueryKeys.FindAllWorkspaceEntries(params.workspaceSlug),
		queryFn: (ctx) => fetchPage(params.workspaceSlug, ctx.pageParam ?? 1),
		getNextPageParam: (lastGroup) => lastGroup.nextPage,
		initialPageParam: 1,
	});

	async function fetchPage(
		slug: string,
		page: number,
	): Promise<FetcherFnResult> {
		const data = await client.queries.workspaceEntriesAll({
			pagination: { page: page, per_page: DEFAULT_LIMIT },
			workspace_slug: slug,
		});

		return {
			entries: data.entries,
			nextPage: data.pagination.next_page,
			prevPage: data.pagination.previous_page,
			totalPages: data.pagination.total_pages,
			totalCount: data.pagination.total_count,
		};
	}

	if (query.status === "pending") {
		return <PageSpinner />;
	}

	if (query.status === "error") {
		return <ErrorPage error={query.error} />;
	}

	return (
		<HotkeysProvider initiallyActiveScopes={["entries"]}>
			<PageLayout
				heading={workspace.activeWorkspace?.name ?? "Workspace"}
				fullScreen
			>
				<title>{workspace.activeWorkspace?.name ?? "All entries"}</title>
				<EntriesList
					title="Recently Added"
					query={query}
					limit={DEFAULT_LIMIT}
				/>
			</PageLayout>
		</HotkeysProvider>
	);
}
