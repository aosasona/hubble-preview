import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import {
	Button,
	DropdownMenu,
	Flex,
	IconButton,
	Text,
	TextField,
	Tooltip,
} from "@radix-ui/themes";
import PageLayout from "$/components/layout/page-layout";
import AddSourceDialog from "$/components/plugins/add-source-dialog";
import { HotkeysProvider } from "react-hotkeys-hook";
import useShortcut from "$/lib/hooks/use-shortcut";
import {
	ArrowLeft,
	ArrowRight,
	Copy,
	DotsThree,
	MagnifyingGlass,
	Plus,
	Trash,
} from "@phosphor-icons/react";
import { useRobinQuery } from "$/lib/hooks";
import PageSpinner from "$/components/page-spinner";
import { toast } from "sonner";
import RemoveSourceDialog from "$/components/plugins/remove-source-dialog";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/plugins/sources",
)({
	component: () => (
		<HotkeysProvider initiallyActiveScopes={["sources"]}>
			<RouteComponent />
		</HotkeysProvider>
	),
});

const PAGE_SIZE = 50;
function RouteComponent() {
	const data = Route.useLoaderData();
	const search = Route.useSearch();

	const [showSourcesDialog, setShowSourcesDialog] = useState(false);
	const [deleteSourceId, setDeleteSourceId] = useState<string | null>(null);
	const [searchQuery, setSearchQuery] = useState("");

	const query = useRobinQuery("plugin.source.list", {
		workspace_id: data.workspace.id,
		pagination: {
			page: search.page,
			per_page: PAGE_SIZE,
		},
	});

	useShortcut(
		["n"],
		(e) => {
			e.preventDefault();
			e.stopPropagation();

			setShowSourcesDialog(true);
		},
		{
			scopes: ["sources"],
		},
	);

	const sources = useMemo(() => {
		if (!query.data) return [];

		const sources = query.data.sources || [];
		if (!searchQuery) return sources;

		const lowerSearchQuery = searchQuery.toLowerCase();
		return sources.filter((source) => {
			return (
				source.name.toLowerCase().includes(lowerSearchQuery) ||
				source.author.toLowerCase().includes(lowerSearchQuery) ||
				source.description.toLowerCase().includes(lowerSearchQuery)
			);
		});
	}, [query.data, searchQuery]);

	function formatDate(dateString: string) {
		const date = new Date(dateString);
		return date.toLocaleDateString("en-US", {
			year: "numeric",
			month: "short",
			day: "numeric",
		});
	}

	function copySourceUrl(sourceUrl: string) {
		navigator.clipboard.writeText(sourceUrl);
		toast.info("Source copied to clipboard");
	}

	if (query.isLoading) {
		return <PageSpinner text="Loading sources..." />;
	}

	return (
		<PageLayout
			heading="Sources"
			header={{
				parent: "settings",
				items: [
					{
						title: "Plugins",
						url: `/${data.workspace.slug}/settings/workspace/plugins`,
					},
					{ title: "Sources" },
				],
			}}
			showHeader
		>
			<Flex width="100%" direction="column" gap="4" mt="4">
				<Flex
					width="100%"
					direction="row"
					justify="between"
					gap="2"
					wrap="wrap"
				>
					<TextField.Root
						name="search"
						variant="surface"
						color="gray"
						placeholder="Search by name, author..."
						size="2"
						value={searchQuery}
						onChange={(e) => setSearchQuery(e.target.value)}
						style={{ width: "min(300px, 75vw)" }}
					>
						<TextField.Slot side="left">
							<MagnifyingGlass />
						</TextField.Slot>
					</TextField.Root>
					<Tooltip content="Add a new source">
						<IconButton onClick={() => setShowSourcesDialog(true)}>
							<Plus size={16} />
						</IconButton>
					</Tooltip>
				</Flex>

				<Flex direction="column">
					{sources.length === 0 ? (
						<Flex
							align="center"
							justify="center"
							direction="column"
							gap="2"
							width="100%"
							height="100%"
							py="6"
						>
							<Text size="2" color="gray">
								No sources found
							</Text>
						</Flex>
					) : (
						<Flex direction="column" gap="2">
							{sources.map((source) => (
								<Flex
									key={source.name}
									direction="column"
									gap="1"
									py="3"
									pl="2"
									className="border-b border-b-[var(--gray-4)]"
								>
									<Flex align="center" justify="between">
										<Text size="3" weight="bold">
											{source.name}
										</Text>
										<DropdownMenu.Root>
											<DropdownMenu.Trigger>
												<IconButton color="gray" variant="ghost" size="1">
													<DotsThree />
												</IconButton>
											</DropdownMenu.Trigger>
											<DropdownMenu.Content size="1">
												<DropdownMenu.Item
													onSelect={() => copySourceUrl(source.source_url)}
												>
													<Copy /> Copy source URL
												</DropdownMenu.Item>
												{data.status.role === "owner" ||
												data.status.role === "admin" ? (
													<DropdownMenu.Item
														color="red"
														onSelect={() => setDeleteSourceId(source.id)}
													>
														<Trash /> Remove source
													</DropdownMenu.Item>
												) : null}
											</DropdownMenu.Content>
										</DropdownMenu.Root>
									</Flex>

									<Text size="1" color="gray">
										{source.description}
									</Text>

									<Flex gap="3" align="center" justify="between" wrap="wrap">
										<Flex gap="2" align="center" justify="between">
											<Text size="1" color="gray" highContrast>
												Publisher:{" "}
												<Tooltip content={source.source_url}>
													<Text as="span" size="1" color="gray">
														{source.author}
													</Text>
												</Tooltip>
											</Text>
											<Text size="1" color="gray" highContrast>
												Versioned by:{" "}
												<Text as="span" size="1" color="gray">
													{source.versioning_strategy}
												</Text>
											</Text>
										</Flex>
										<Text size="1" color="gray" highContrast>
											Added on{" "}
											<Text as="span" size="1" color="gray">
												{formatDate(source.added_at)}
											</Text>
										</Text>
									</Flex>
								</Flex>
							))}
						</Flex>
					)}
				</Flex>

				<Flex align="center" justify="end" gap="3" width="100%" mt="4">
					<Button
						size="1"
						color="gray"
						variant="soft"
						disabled={!query.data?.pagination.previous_page}
						asChild
					>
						<Link
							to="/$workspaceSlug/settings/workspace/plugins/sources"
							params={{ workspaceSlug: data.workspace.slug }}
							search={{ page: query.data?.pagination.previous_page ?? 0 }}
						>
							<ArrowLeft size={12} />
							Previous
						</Link>
					</Button>

					<Button
						variant="solid"
						size="1"
						disabled={!query.data?.pagination.next_page}
						asChild
					>
						<Link
							to="/$workspaceSlug/settings/workspace/plugins/sources"
							params={{ workspaceSlug: data.workspace.slug }}
							search={{ page: query.data?.pagination.next_page ?? 1 }}
							className="text-[var(--accent-foreground)] hover:text-inherit"
						>
							Next
							<ArrowRight size={12} />
						</Link>
					</Button>
				</Flex>
			</Flex>

			<AddSourceDialog
				open={showSourcesDialog}
				onOpenChange={setShowSourcesDialog}
				data={data}
			/>
			<RemoveSourceDialog
				source={sources.find((s) => s.id === deleteSourceId) ?? null}
				workspace={data.workspace}
				onClose={() => setDeleteSourceId(null)}
			/>
		</PageLayout>
	);
}
