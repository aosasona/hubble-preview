import { createLazyFileRoute } from "@tanstack/react-router";
import QueryKeys from "$/lib/keys";
import EntriesList, {
	type FetcherFnResult,
} from "$/components/entry/entries-list";
import client from "$/lib/server";
import { useInfiniteQuery } from "@tanstack/react-query";
import CollectionIndicator from "$/components/collection-indicator";
import PageLayout from "$/components/layout/page-layout";
import PageSpinner from "$/components/page-spinner";
import ErrorPage from "$/components/error-page";
import { HotkeysProvider } from "react-hotkeys-hook";
import { useSnapshot } from "valtio";
import stores from "$/stores";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/c/$collectionSlug/",
)({
	component: RouteComponent,
});

const DEFAULT_LIMIT = 30;

function RouteComponent() {
	const params = Route.useParams();
	const workspace = useSnapshot(stores.workspace);

	const query = useInfiniteQuery({
		queryKey: QueryKeys.FindAllCollectionEntries(
			params.workspaceSlug,
			params.collectionSlug,
		),
		queryFn: (ctx) => fetchPage(ctx.pageParam ?? 1),
		getNextPageParam: (lastGroup) => lastGroup.nextPage,
		getPreviousPageParam: (firstGroup) => firstGroup.prevPage,
		initialPageParam: 1,
	});

	async function fetchPage(page: number): Promise<FetcherFnResult> {
		const data = await client.queries.collectionEntriesAll({
			pagination: { page: page, per_page: DEFAULT_LIMIT },
			workspace_slug: params.workspaceSlug,
			collection_slug: params.collectionSlug,
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
		<PageLayout
			heading={{
				title: workspace.activeCollection?.name ?? "Collection",
				component: () => (
					<CollectionIndicator name={workspace.activeCollection?.name ?? ""} />
				),
			}}
			fullScreen
		>
			<HotkeysProvider initiallyActiveScopes={["entries"]}>
				<EntriesList
					title={workspace.activeCollection?.name ?? "All Entries"}
					query={query}
					limit={DEFAULT_LIMIT}
					activeCollectionSlug={params.collectionSlug}
				/>
			</HotkeysProvider>
		</PageLayout>
	);
}
