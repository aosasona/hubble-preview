import stores from "$/stores";
import { ArrowsDownUp, DotsThree, Funnel, Trash } from "@phosphor-icons/react";
import {
	Button,
	DropdownMenu,
	Flex,
	Heading,
	IconButton,
	Text,
} from "@radix-ui/themes";
import { useSnapshot } from "valtio";
import {
	ENTRY_TYPE_MAPPING,
	SORT_BY_MAPPING,
	STATUS_MAPPING,
	type Mapping,
	type SortBy,
} from "./filter";
import FilterDropdown from "./filter-dropdown";
import FilterBar from "./filter-bar";
import { AnimatePresence } from "motion/react";
import { useRobinMutation } from "$/lib/hooks";
import { toast } from "sonner";
import QueryKeys from "$/lib/keys";
import { pluralize } from "$/lib/utils";
import { CollectionIcon } from "$/components/collection-indicator";
import { useMemo } from "react";

type Props = {
	title: string;
	isSearch?: boolean;
};

export default function TopBar(props: Props) {
	const workspace = useSnapshot(stores.workspace);

	const entriesList = useSnapshot(stores.entriesList);
	const filters = useSnapshot(stores.entriesList.filters);
	const selections = useSnapshot(stores.entriesList.selections);

	const requeueMutation = useRobinMutation("entry.requeue", {
		onSuccess: (d) => {
			stores.entriesList.clearSelections();
			toast.success(`Re-queued ${d.count} ${pluralize("entry", d.count)}`);
		},
		invalidates: [
			QueryKeys.FindAllWorkspaceEntries(workspace.activeWorkspace?.slug ?? ""),
			QueryKeys.FindAllCollectionEntries(
				workspace.activeWorkspace?.slug ?? "",
				workspace.activeCollection?.slug ?? "",
			),
		],
		retry: false,
	});

	const FilterCollections: Record<string, Mapping> = useMemo(() => {
		// If we are in a collection, we don't need to show the collection filter
		if (workspace.activeCollection?.id) {
			return {};
		}

		const collections = workspace.activeWorkspaceCollections;
		if (!collections) return {};

		return Object.fromEntries(
			collections.map((collection) => [
				collection.id,
				{
					label: collection.name,
					icon: () => <CollectionIcon name={collection.name} size={5} />,
				},
			]),
		);
	}, [workspace.activeCollection?.id, workspace.activeWorkspaceCollections]);

	function onRequeue() {
		toast.promise(
			() => {
				return requeueMutation.call({
					workspace_id: workspace.activeWorkspace?.id ?? "",
					entry_ids: Array.from(selections),
				});
			},
			{ loading: "Requeuing entries..." },
		);
	}

	return (
		<Flex
			direction="column"
			width="100%"
			maxWidth="100%"
			className="sticky top-0 right-0 left-0 z-20"
		>
			<Flex
				width="100%"
				py="2"
				px="4"
				gap="2"
				align="center"
				justify="between"
				className="border-[var(--gray-4)] border-b bg-background"
				wrap="wrap"
			>
				<Heading size="3" weight="medium">
					{props.title}
				</Heading>

				<Flex align="center" gap="4">
					<Flex align="center" gap="2">
						{/* MARK: Sort By */}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button variant="soft" color="gray" size="1">
									<ArrowsDownUp size={16} />
									{entriesList.sortBy === "default" && props.isSearch
										? "Relevance"
										: SORT_BY_MAPPING[stores.entriesList.sortBy].label}
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content size="2" align="start">
								{Object.entries(SORT_BY_MAPPING).map(([key, order]) => (
									<DropdownMenu.Item
										key={key}
										onSelect={() =>
											stores.entriesList.setSortBy(key as unknown as SortBy)
										}
									>
										<Flex align="center" gap="2">
											<order.icon size={16} />
											<Text size="2">{order.label}</Text>
										</Flex>
									</DropdownMenu.Item>
								))}
							</DropdownMenu.Content>
						</DropdownMenu.Root>

						{/* MARK: Filter */}
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button size="1" variant="soft" color="gray">
									<Flex align="center" gap="2">
										{filters.length > 0 ? (
											<Text size="1" color="gray">
												{filters.length}
											</Text>
										) : null}
										<Flex align="center" gap="1">
											<Funnel size={14} color="var(--gray-11)" />
											<Text size="1" weight="medium">
												Filter
											</Text>
										</Flex>
									</Flex>
								</Button>
							</DropdownMenu.Trigger>

							<DropdownMenu.Content size="2" align="start">
								<FilterDropdown
									type="type"
									category="Type"
									items={ENTRY_TYPE_MAPPING}
								/>

								<FilterDropdown
									type="status"
									category="Status"
									items={STATUS_MAPPING}
								/>

								<FilterDropdown
									type="collection"
									category="Collection"
									items={FilterCollections}
								/>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</Flex>

					{selections.size > 0 ? (
						<Flex align="center" gap="4">
							<Text size="1" color="gray">
								{selections.size} selected
							</Text>

							<DropdownMenu.Root>
								<DropdownMenu.Trigger>
									<IconButton variant="ghost" color="gray" size="1">
										<DotsThree size={20} />
									</IconButton>
								</DropdownMenu.Trigger>
								<DropdownMenu.Content size="2" align="start">
									<DropdownMenu.Item
										onSelect={onRequeue}
										disabled={requeueMutation.isMutating}
									>
										Re-queue selected items
									</DropdownMenu.Item>
									<DropdownMenu.Item
										onSelect={() => stores.entriesList.clearSelections()}
										shortcut="shift+x"
									>
										Deselect all
									</DropdownMenu.Item>
									<DropdownMenu.Item
										onSelect={() => {
											stores.app.openDialog("deleteSelectedEntries");
										}}
										color="red"
										shortcut="d"
									>
										<Trash size={16} /> Delete
									</DropdownMenu.Item>
								</DropdownMenu.Content>
							</DropdownMenu.Root>
						</Flex>
					) : null}
				</Flex>
			</Flex>

			{/* MARK: Filter bar */}
			<Flex direction="column" width="100%" maxWidth="100%">
				<AnimatePresence>
					{filters.length > 0 ? <FilterBar /> : null}
				</AnimatePresence>
			</Flex>
		</Flex>
	);
}
