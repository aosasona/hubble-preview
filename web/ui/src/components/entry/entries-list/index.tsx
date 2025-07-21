import { useCallback, useEffect, useMemo, useRef } from "react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { useSnapshot } from "valtio";
import { Box, Flex, Heading, Kbd, Link, Spinner, Text } from "@radix-ui/themes";
import type { Entry } from "$/lib/server/types";
import type {
	InfiniteData,
	UseInfiniteQueryResult,
} from "@tanstack/react-query";
import Show from "../../show";
import stores from "$/stores";
import Item from "$/components/entry/entries-list/item";
import { FilePlus } from "@phosphor-icons/react";
import TopBar from "./top-bar";
import PreviewPane from "./preview-pane";
import useEntriesListShortcuts from "$/lib/hooks/use-entry-shortcuts";

export type FetcherFnResult = {
	entries: Entry[];
	nextPage: number | null;
	prevPage: number | null;
	totalPages: number;
	totalCount: number;
};

type Props = {
	query: UseInfiniteQueryResult<InfiniteData<FetcherFnResult, unknown>, Error>;
	limit: number;
	title: string;
	activeCollectionSlug?: string;
};

export default function EntriesList(props: Props) {
	const app = useSnapshot(stores.app);
	const entriesList = useSnapshot(stores.entriesList);

	const parentRef = useRef<HTMLDivElement>(null);

	const { data, isFetching, isFetchingNextPage, fetchNextPage, hasNextPage } =
		props.query;

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
		const e = data?.pages.flatMap((page) => page.entries) ?? [];
		const filtered = e.filter(applyFilters);
		return filtered.sort(sortByFn);
	}, [applyFilters, data, sortByFn]);

	const getItemKey = useCallback(
		(index: number) => {
			if (index >= entries.length) {
				return `loader${index}`;
			}

			return entries[index].id;
		},
		[entries],
	);

	const virtualizer = useVirtualizer({
		count: hasNextPage ? entries.length + 1 : entries.length,
		getScrollElement: () => parentRef.current,
		overscan: 5,
		estimateSize: () => 100,
		getItemKey: getItemKey,
	});

	// biome-ignore lint/correctness/useExhaustiveDependencies: we definitelty need it
	useEffect(() => {
		const virtualItems = virtualizer.getVirtualItems();
		if (!virtualItems.length) {
			return;
		}

		const lastItem = virtualItems[virtualItems.length - 1];

		if (
			lastItem.index >= entries.length - 1 &&
			hasNextPage &&
			!isFetchingNextPage
		) {
			fetchNextPage();
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [
		fetchNextPage,
		hasNextPage,
		isFetchingNextPage,
		entries.length,
		// eslint-disable-next-line react-hooks/exhaustive-deps
		virtualizer.getVirtualItems(),
	]);

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
					<TopBar title={props.title} />

					{/* MARK: entries */}
					<Show when={entries.length === 0}>
						<Flex
							direction="column"
							height="100%"
							align="center"
							justify="center"
							gap="2"
						>
							<FilePlus size={48} color="var(--gray-8)" />
							<Heading size="3" weight="medium">
								Oops! No entries yet...
							</Heading>
							<Text color="gray">
								Press <Kbd>C</Kbd> or click{" "}
								<Link
									color={app.accentColor}
									onClick={() => app.openDialog("importItem")}
									underline="auto"
									className="!cursor-pointer"
								>
									here
								</Link>{" "}
								to create a new entry
							</Text>
						</Flex>
					</Show>

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
										const isLoaderRow = virtualItem.index > entries.length - 1;
										const entry = entries[virtualItem.index];

										if (isLoaderRow) {
											if (!hasNextPage) {
												return null;
											}

											return (
												<Flex
													width="100%"
													align="center"
													justify="center"
													py="5"
													gap="1"
													key={virtualItem.index}
												>
													<Show when={!hasNextPage}>
														<Text size="1" color="gray" align="center">
															You've reached the end of the list.
														</Text>
													</Show>
													<Show when={isFetchingNextPage}>
														<Spinner size="2" />
														<Text size="1" color="gray" align="center">
															Loading more entries...
														</Text>
													</Show>
												</Flex>
											);
										}

										return (
											<Item
												key={virtualItem.index}
												index={virtualItem.index}
												entry={entry}
												measureElement={virtualizer.measureElement}
												activeCollectionSlug={props.activeCollectionSlug}
											/>
										);
									})}
									{isFetching && !isFetchingNextPage && (
										<Flex width="100%" align="center" justify="center" py="4">
											<Text size="1" color="gray" align="center">
												Loading more data in the background...
											</Text>
										</Flex>
									)}
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
