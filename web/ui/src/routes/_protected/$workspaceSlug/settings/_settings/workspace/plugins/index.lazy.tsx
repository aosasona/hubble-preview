import PageLayout from "$/components/layout/page-layout";
import { useRobinMutation, useRobinQuery } from "$/lib/hooks";
import {
	ArchiveBox,
	ArrowClockwise,
	ArrowDown,
	ArrowSquareOut,
	ArrowUpRight,
	Info,
	MagnifyingGlass,
	MinusCircle,
} from "@phosphor-icons/react";
import {
	Text,
	Flex,
	TextField,
	Skeleton,
	DropdownMenu,
	Button,
	Grid,
	Card,
	Badge,
	Tooltip,
	AlertDialog,
	Link as ThemedLink,
	Box,
	DataList,
	Heading,
	Code,
	IconButton,
} from "@radix-ui/themes";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { toast } from "sonner";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/plugins/",
)({
	component: RouteComponent,
});

enum View {
	All = "all",
	Installed = "installed",
	Available = "available",
}

type TargetPlugin = {
	source_id: string;
	name: string;
	view_only: boolean;
};

function RouteComponent() {
	const data = Route.useLoaderData();

	const [view, setView] = useState<View>(View.All);
	const [searchQuery, setSearchQuery] = useState("");
	const [current, setCurrent] = useState<TargetPlugin | null>(null);

	const query = useRobinQuery("plugin.list", {
		workspace_id: data.workspace.id,
	});

	const plugins = useMemo(() => {
		if (!query.data) return [];

		let plugins = query.data.plugins;
		const search = searchQuery.trim().toLowerCase();

		// Filter by active view
		if (view === View.Installed) {
			plugins = plugins.filter((plugin) => plugin.installed);
		} else if (view === View.Available) {
			plugins = plugins.filter((plugin) => !plugin.installed);
		}

		// Search
		if (search !== "") {
			plugins = plugins.filter((plugin) => {
				return (
					plugin.name.toLowerCase().includes(search) ||
					plugin.description.toLowerCase().includes(search)
				);
			});
		}

		return plugins;
	}, [query.data, searchQuery, view]);

	const target = useMemo(() => {
		if (!current) return null;
		return plugins.find((plugin) => {
			return (
				plugin.source.id === current.source_id && plugin.name === current.name
			);
		});
	}, [current, plugins]);

	const installMutation = useRobinMutation("plugin.install", {
		onSuccess: (data) => {
			toast.success(`Plugin "${data.plugin_name}" installed`);
		},
		invalidates: ["plugin.list"],
		retry: false,
	});

	const updateMutation = useRobinMutation("plugin.update", {
		onSuccess: (data) => {
			toast.success(`Plugin "${data.plugin_name}" updated`);
		},
		invalidates: ["plugin.list"],
		retry: false,
	});

	const removeMutation = useRobinMutation("plugin.remove", {
		onSuccess: (data) => {
			toast.success(`Plugin "${data.plugin_name}" removed`);
		},
		invalidates: ["plugin.list"],
		retry: false,
	});

	async function installOrUpdatePlugin(
		name: string,
		action: "install" | "update",
	) {
		if (!current) return;

		if (action === "install") {
			await installMutation.call({
				workspace_id: data.workspace.id,
				source_id: current.source_id,
				name: name,
			});
		} else if (action === "update") {
			await updateMutation.call({
				workspace_id: data.workspace.id,
				source_id: current.source_id,
				name: name,
			});
		}

		setCurrent(null);
	}

	return (
		<PageLayout heading="Plugins" header={{ parent: "settings" }} showHeader>
			<Flex width="100%" direction="column" gap="2" mt="4">
				<Flex
					width="100%"
					direction="row"
					align="center"
					justify="between"
					gap="3"
					wrap="wrap"
				>
					<Flex gap="2" align="center">
						<TextField.Root
							name="search"
							type="search"
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
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button color="gray" variant="surface">
									{view === View.All
										? "All"
										: view === View.Installed
											? "Installed"
											: "Available"}
									<DropdownMenu.TriggerIcon />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content>
								<DropdownMenu.RadioGroup value={view}>
									<DropdownMenu.RadioItem
										value={View.All}
										onClick={() => setView(View.All)}
									>
										All
									</DropdownMenu.RadioItem>
									<DropdownMenu.RadioItem
										value={View.Available}
										onClick={() => setView(View.Available)}
									>
										Available
									</DropdownMenu.RadioItem>
									<DropdownMenu.RadioItem
										value={View.Installed}
										onClick={() => setView(View.Installed)}
									>
										Installed
									</DropdownMenu.RadioItem>
								</DropdownMenu.RadioGroup>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</Flex>
					<Link
						to="/$workspaceSlug/settings/workspace/plugins/sources"
						params={{
							workspaceSlug: data.workspace.slug,
						}}
						className="!no-underline"
					>
						<Flex align="center" gap="1">
							<Text>Manage sources</Text>
							<ArrowUpRight />
						</Flex>
					</Link>
				</Flex>

				{/* MARK: all plugins */}
				<Skeleton loading={query.isLoading}>
					<Flex direction="column" mt="3">
						{plugins.length === 0 ? (
							<Flex
								direction="column"
								height="200px"
								align="center"
								justify="center"
							>
								<Text size="2" color="gray">
									No plugins found
								</Text>
							</Flex>
						) : (
							<Grid columns={{ initial: "1", sm: "2" }} gap="2">
								{plugins.map((plugin) => (
									<Card key={plugin.name}>
										<Flex
											direction="column"
											gap={{ initial: "1", md: "2" }}
											style={{ height: "100%" }}
										>
											<Flex direction="row" justify="between">
												<Text color="gray" size="3" weight="bold" highContrast>
													{plugin.name}
												</Text>

												<Flex align="center" gap="2">
													{plugin.installed ? (
														<Badge size="1" color="gray" variant="surface">
															Installed
														</Badge>
													) : (
														<div />
													)}
													<IconButton
														variant="ghost"
														onClick={() => {
															setCurrent({
																source_id: plugin.source.id,
																name: plugin.name,
																view_only: true,
															});
														}}
													>
														<Info size={18} />
													</IconButton>
												</Flex>
											</Flex>

											<Text color="gray" size={{ initial: "2", md: "1" }}>
												{plugin.description}
											</Text>

											<Flex direction="row" justify="between" mt="auto">
												<Tooltip content={plugin.source.url}>
													<Flex align="center" gap="1">
														<ArchiveBox size={14} color="var(--gray-11)" />
														<Text size="1" color="gray">
															{plugin.source.name}
														</Text>
													</Flex>
												</Tooltip>

												<Flex align="center" gap="2">
													{plugin.installed && plugin.updatable ? (
														<Button
															size="1"
															variant="soft"
															color="orange"
															loading={updateMutation.isMutating}
															disabled={
																installMutation.isMutating ||
																removeMutation.isMutating
															}
															onClick={() => {
																setCurrent({
																	source_id: plugin.source.id,
																	name: plugin.name,
																	view_only: false,
																});
															}}
														>
															<ArrowClockwise /> Update
														</Button>
													) : null}

													{plugin.installed ? (
														<Button
															size="1"
															variant="soft"
															color="red"
															loading={
																removeMutation.isMutating &&
																target?.name === plugin.name
															}
															disabled={
																installMutation.isMutating ||
																updateMutation.isMutating
															}
															onClick={() => {
																removeMutation.call({
																	workspace_id: data.workspace.id,
																	plugin_id: plugin.identifier,
																});
															}}
														>
															<MinusCircle /> Remove
														</Button>
													) : (
														<Button
															size="1"
															variant="soft"
															color="grass"
															loading={
																installMutation.isMutating &&
																target?.name === plugin.name
															}
															disabled={
																updateMutation.isMutating ||
																removeMutation.isMutating
															}
															onClick={() => {
																setCurrent({
																	source_id: plugin.source.id,
																	name: plugin.name,
																	view_only: false,
																});
															}}
														>
															<ArrowDown /> Install
														</Button>
													)}
												</Flex>
											</Flex>
										</Flex>
									</Card>
								))}
							</Grid>
						)}
					</Flex>
				</Skeleton>
			</Flex>

			<AlertDialog.Root
				open={!!current}
				onOpenChange={(open) => !open && setCurrent(null)}
			>
				<AlertDialog.Content maxWidth="475px">
					<AlertDialog.Title mb="1">{target?.name}</AlertDialog.Title>
					<AlertDialog.Description size="2" color="gray">
						{current?.view_only
							? "This plugin is already installed"
							: "Before you proceed, review the following details (& privileges requested by this plugin; if any) for security reasons."}
					</AlertDialog.Description>

					<Flex direction="column" mt="3" gap="5">
						<Box>
							<Heading size="2" color="gray" weight="medium" mb="1">
								Details
							</Heading>
							<Card>
								<DataList.Root size="2">
									<DataList.Item>
										<DataList.Label>Name</DataList.Label>
										<DataList.Value>{target?.name}</DataList.Value>
									</DataList.Item>
									<DataList.Item>
										<DataList.Label>Description</DataList.Label>
										<DataList.Value>{target?.description}</DataList.Value>
									</DataList.Item>
									<DataList.Item>
										<DataList.Label>Source</DataList.Label>
										<DataList.Value>
											<ThemedLink href={target?.source.url} underline="always">
												<Flex align="center" gap="1">
													<Text>{target?.source.name}</Text>
													<ArrowSquareOut size={16} />
												</Flex>
											</ThemedLink>
										</DataList.Value>
									</DataList.Item>
									<DataList.Item>
										<DataList.Label>Author</DataList.Label>
										<DataList.Value>{target?.author}</DataList.Value>
									</DataList.Item>
								</DataList.Root>
							</Card>
						</Box>

						<Box>
							<Heading size="2" color="gray" weight="medium" mb="1">
								Privileges
							</Heading>
							<Card>
								<DataList.Root size="2">
									{target?.privileges.map((privilege) => (
										<DataList.Item key={privilege.identifier}>
											<DataList.Label>
												<Code>{privilege.identifier}</Code>
											</DataList.Label>
											<DataList.Value>{privilege.description}</DataList.Value>
										</DataList.Item>
									))}
								</DataList.Root>
							</Card>
						</Box>

						<Flex justify="end" gap="3">
							<AlertDialog.Cancel>
								<Button variant="soft" color="gray">
									Cancel
								</Button>
							</AlertDialog.Cancel>
							{!current?.view_only ? (
								<Button
									variant="solid"
									loading={
										installMutation.isMutating || updateMutation.isMutating
									}
									onClick={() => {
										installOrUpdatePlugin(
											target?.name || "",
											target?.installed ? "update" : "install",
										);
									}}
								>
									{target?.updatable ? <ArrowClockwise /> : <ArrowDown />}
									{target?.installed ? "Update" : "Install"}
								</Button>
							) : null}
						</Flex>
					</Flex>
				</AlertDialog.Content>
			</AlertDialog.Root>
		</PageLayout>
	);
}
