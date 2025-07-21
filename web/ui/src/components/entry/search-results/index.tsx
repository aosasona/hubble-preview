import type {
	CollapsedSearchResult,
	Entry,
	HybridSearchResults,
} from "$/lib/server/types";
import { Box, Flex, Heading } from "@radix-ui/themes";
import TopBar from "../entries-list/top-bar";
import Show from "$/components/show";
import { FileMagnifyingGlass, MagnifyingGlass } from "@phosphor-icons/react";
import stores from "$/stores";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useRef, useCallback, useMemo } from "react";
import { useSnapshot } from "valtio";
import useEntriesListShortcuts from "$/lib/hooks/use-entry-shortcuts";
import Item from "./item";
import PreviewPane from "../entries-list/preview-pane";

type Props = {
	data: HybridSearchResults | null;
	timeTaken: number;
	currentQuery: string;
	serverQuery: string;
	isFetching: boolean;
};

export default function SearchResults({
	data,
	timeTaken,
	currentQuery: query,
	isFetching,
	serverQuery,
}: Props) {
	const app = useSnapshot(stores.app);
	const entriesList = useSnapshot(stores.entriesList);

	const parentRef = useRef<HTMLDivElement>(null);

	const applyFilters = useCallback(
		(entry: Entry) => {
			return entriesList.applyFilters(entry);
		},
		[entriesList],
	);

	const sortByFn = useCallback(
		(a: Entry, b: Entry) => entriesList.applySortingOrder(a, b),
		[entriesList],
	);

	const entries = useMemo(() => {
		const e = data?.results ?? [];
		const filtered = (e as unknown as Entry[]).filter(applyFilters);
		return filtered.sort(sortByFn);
	}, [applyFilters, data?.results, sortByFn]);

	const virtualizer = useVirtualizer({
		count: entries.length,
		getScrollElement: () => parentRef.current,
		estimateSize: () => (app.uiScale === "110%" ? 150 : 120),
		overscan: 5,
	});

	// Shortcuts
	useEntriesListShortcuts(getEntry, entries);

	const items = virtualizer.getVirtualItems();

	function getEntry(index: string) {
		if (index.includes("-")) {
			// We most likely have an entry ID
			return entries.find((e) => e.id === index);
		}
		return entries[Number.parseInt(index)];
	}

	return (
		<Flex
			flexGrow="1"
			minHeight="0"
			height={{ initial: "calc(100vh - 40px)", md: "100%" }}
			direction="column"
			width="100%"
		>
			<Flex flexGrow="1" minHeight={{ md: "0" }} height="100%" width="100%">
				<Flex
					flexGrow="1"
					direction="column"
					minHeight={{ md: "0" }}
					height="100%"
					width="100%"
					maxWidth="100%"
					position="relative"
				>
					{/* MARK: Filter bar */}
					<TopBar
						title={`Showing results for "${query}"${timeTaken > 0 ? ` (completed in ${timeTaken}ms)` : ""}`}
						isSearch
					/>

					{/* MARK: Loading state */}
					<Show when={isFetching}>
						<Flex
							height="100%"
							align="center"
							justify="center"
							gap="2"
							className="animate-pulse"
						>
							<MagnifyingGlass size={24} color="var(--gray-10)" />
							<Heading size="1" weight="medium" color="gray">
								Searching...
							</Heading>
						</Flex>
					</Show>

					{/* MARK: empty state */}
					<Show when={entries.length === 0 && !isFetching}>
						<Flex
							direction="column"
							height="100%"
							align="center"
							justify="center"
							gap="2"
						>
							<FileMagnifyingGlass size={48} color="var(--gray-9)" />
							<Heading size="1" weight="medium" color="gray">
								Oops! No results found, try a different query...
							</Heading>
						</Flex>
					</Show>

					{/* MARK: results */}
					<Show when={entries.length > 0}>
						<Box
							ref={parentRef}
							overflow="auto"
							flexGrow="1"
							style={{ contain: "strict" }}
						>
							<Box
								style={{
									height: virtualizer.getTotalSize(),
									width: "100%",
									position: "relative",
								}}
							>
								<Box
									width="100%"
									position="absolute"
									top="0"
									left="0"
									right="0"
									style={{
										transform: `translateY(${items[0]?.start ?? 0}px)`,
									}}
									data-focusable-list
									data-last-focused={entriesList.preview?.id}
								>
									{items.map((virtualItem) => {
										const result = entries[
											virtualItem.index
										] as unknown as CollapsedSearchResult;

										return (
											<Item
												key={virtualItem.index}
												index={virtualItem.index}
												result={result}
												currentQuery={query}
												serverQuery={serverQuery}
												measureElement={virtualizer.measureElement}
												scores={{
													maximum: data?.max_hybrid_score ?? 0,
													minimum: data?.min_hybrid_score ?? 0,
												}}
											/>
										);
									})}
								</Box>
							</Box>
						</Box>
					</Show>
				</Flex>

				{/* MARK: Preview sidebar */}
				<PreviewPane />
			</Flex>
		</Flex>
	);
}
