import PageLayout from "$/components/layout/page-layout";
import { useRobinQuery } from "$/lib/hooks";
import {
	Button,
	Flex,
	Heading,
	IconButton,
	ScrollArea,
	Spinner,
	Text,
} from "@radix-ui/themes";
import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { MagnifyingGlass, X, XCircle } from "@phosphor-icons/react";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import Show from "$/components/show";
import { useDebouncedCallback } from "use-debounce";
import * as v from "valibot";
import { useEffect, useState } from "react";
import SearchResults from "$/components/entry/search-results";
import { HotkeysProvider } from "react-hotkeys-hook";

const searchSchema = v.object({
	q: v.optional(v.string()),
});

export const Route = createFileRoute("/_protected/$workspaceSlug/search")({
	validateSearch: searchSchema,
	component: RouteComponent,
});

function RouteComponent() {
	const search = Route.useSearch();
	const params = Route.useParams();
	const navigate = useNavigate({ from: Route.fullPath });

	const workspace = useSnapshot(stores.workspace);

	const [value, setValue] = useState(search.q ?? "");

	const query = useRobinQuery(
		"entry.search",
		{
			workspace_slug: params.workspaceSlug,
			query: search.q ?? "",
		},
		{
			queryKey: ["entry", "search", search.q ?? ""],
			enabled: !!(search.q?.length && search.q.length >= 2),
			staleTime: 1000 * 60 * 2, // 2 minute
		},
	);

	function sendQuery(q: string) {
		if (!q.length || q?.trim() === "" || q?.length < 2) {
			return;
		}

		navigate({ search: { q: q } });

		return query.refetch({ cancelRefetch: true }).catch();
	}

	const debouncedCall = useDebouncedCallback((q) => {
		sendQuery(q);
	}, 750);

	// biome-ignore lint/correctness/useExhaustiveDependencies: <explanation>
	useEffect(() => {
		if (search.q) {
			sendQuery(search.q);
			stores.workspace.addRecentSearchQuery(search.q?.trim() ?? "");
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [search.q]);

	return (
		<PageLayout heading="Search" fullScreen>
			<Flex
				direction="column"
				position="relative"
				width="100%"
				height="100%"
				minHeight="0"
				flexGrow="1"
			>
				<Flex
					width="100%"
					align="center"
					justify="between"
					px="2"
					py="2"
					gap="3"
					className="sticky top-0 right-0 left-0 z-20"
					style={{ borderBottom: "1px solid var(--gray-4)" }}
				>
					<Flex align="center" gap="3" width="100%" px="2">
						<Flex align="center" gap="2" width="100%" px="2">
							<MagnifyingGlass size={16} color="var(--gray-11)" />
							<input
								name="query"
								type="text"
								value={value}
								onChange={(e) => {
									setValue(e.target.value);
									if (e.target.value?.trim() === "") {
										navigate({ search: { q: "" } });
										return;
									}
									debouncedCall(e.target.value);
								}}
								placeholder="Search across your workspace..."
								id="search-entries"
								className="w-full border-none bg-transparent px-[var(--space-1)] py-[var(--space-1)] text-[var(--gray-12)] outline-none placeholder:text-[var(--gray-8)]"
								// biome-ignore lint/a11y/noAutofocus: <explanation>
								autoFocus
							/>
							<IconButton
								variant="ghost"
								color="gray"
								hidden={value?.trim() === ""}
								onClick={() => {
									setValue("");
									navigate({ search: { q: "" } });
								}}
							>
								<XCircle size={16} />
							</IconButton>
						</Flex>
						{query.isLoading && value?.length >= 2 ? (
							<Spinner size="2" />
						) : null}
					</Flex>
				</Flex>

				<Show when={value?.length < 2}>
					<ScrollArea
						style={{
							flexGrow: 1,
							width: "100%",
							maxWidth: "100vw",
							height: "100%",
							minHeight: 0,
						}}
					>
						<Flex direction="column">
							<Flex justify="between" align="center" px="5" py="4">
								<Heading size="2" weight="regular" color="gray">
									Recent searches
								</Heading>

								<Button
									size="1"
									color="gray"
									variant="ghost"
									disabled={workspace.recentSearchQueries.length === 0}
									onClick={() => stores.workspace.clearRecentSearchQueries()}
								>
									Clear all
								</Button>
							</Flex>
							{Array.from(workspace.recentSearchQueries)
								.reverse()
								.map((query) => (
									<Flex
										key={query}
										px="5"
										py="2"
										align="center"
										justify="between"
										className="cursor-pointer transition-all hover:bg-[var(--gray-2)]"
										onClick={() => {
											setValue(query);
											sendQuery(query);
										}}
									>
										<Flex align="center" gap="2">
											<MagnifyingGlass size={16} color="var(--gray-10)" />
											<Text size="2" color="gray" highContrast>
												{query}
											</Text>
										</Flex>

										<IconButton
											color="gray"
											variant="ghost"
											onClick={(e) => {
												e.preventDefault();
												e.stopPropagation();

												stores.workspace.removeRecentSearchQuery(query);
											}}
										>
											<X />
										</IconButton>
									</Flex>
								))}
						</Flex>
					</ScrollArea>
				</Show>

				<Show when={value?.length >= 2}>
					<HotkeysProvider initiallyActiveScopes={["entries"]}>
						<SearchResults
							data={query.data?.results ?? null}
							timeTaken={query.data?.time_taken_ms ?? 0}
							currentQuery={value ?? search.q ?? ""}
							serverQuery={query.data?.query ?? ""}
							isFetching={query.isLoading}
						/>
					</HotkeysProvider>
				</Show>
			</Flex>
		</PageLayout>
	);
}
