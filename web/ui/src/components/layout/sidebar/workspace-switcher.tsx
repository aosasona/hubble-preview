import { CaretDown, Gear, Plus } from "@phosphor-icons/react";
import {
	Button,
	DropdownMenu,
	Flex,
	Grid,
	IconButton,
	Text,
} from "@radix-ui/themes";
import Show from "$/components/show";
import stores from "$/stores";
import { Link, useParams } from "@tanstack/react-router";
import { useMemo } from "react";
import { useSnapshot } from "valtio";
import WorkspaceIndicator from "$/components/workspace/workspace-indicator";

export default function WorkspaceSwitcher() {
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const app = useSnapshot(stores.app);
	const workspaces = useSnapshot(stores.workspace);

	const otherWorkspaces = useMemo(() => {
		return workspaces.all.filter((w) => w.slug !== params.workspaceSlug);
	}, [params.workspaceSlug, workspaces.all]);

	return (
		<>
			<Show when={workspaces.all.length === 0}>
				<Link
					to="/workspace/new"
					className="!no-underline rounded-[var(--radius-2)] border border-[var(--accent-6)] bg-[var(--accent-surface)] px-[var(--space-3)] py-[var(--space-1)] transition-all hover:bg-[var(--accent-4)]"
				>
					<Flex align="center" justify="center" gap="2">
						<Plus color="var(--accent-11)" />
						<Text size="2" color={app.accentColor}>
							Create a workspace
						</Text>
					</Flex>
				</Link>
			</Show>

			<Show when={workspaces.all.length > 0}>
				<Flex
					width="100%"
					direction="row"
					align="center"
					justify="between"
					gap="4"
					p="1"
				>
					<Grid width="100%">
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button variant="ghost" color="gray" size="2">
									<Flex width="100%" align="center" justify="start" gap="2">
										<WorkspaceIndicator />
										<CaretDown size={12} className="text-[var(--accent-10)]" />
									</Flex>
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content size="2" align="center">
								<Show when={otherWorkspaces.length > 0}>
									<DropdownMenu.Group>
										<DropdownMenu.Label>Switch Workspace</DropdownMenu.Label>

										{otherWorkspaces.map((workspace, idx) => (
											<DropdownMenu.Item
												key={workspace.id}
												shortcut={idx < 9 ? `⌥ ${idx + 1}` : undefined}
												asChild
											>
												<Link
													to="/$workspaceSlug"
													params={{ workspaceSlug: workspace.slug }}
													className="dropdown"
												>
													<Text>{workspace.name}</Text>
												</Link>
											</DropdownMenu.Item>
										))}
									</DropdownMenu.Group>

									<DropdownMenu.Separator />
								</Show>

								{workspaces.activeWorkspace ? (
									<DropdownMenu.Item asChild>
										<Link
											to="/$workspaceSlug/settings/workspace"
											params={{
												workspaceSlug: workspaces.activeWorkspace.slug,
											}}
											className="dropdown"
										>
											<Flex align="center" gap="1">
												<Gear size={18} />
												<Text>Manage workspace</Text>
											</Flex>
										</Link>
									</DropdownMenu.Item>
								) : null}

								<DropdownMenu.Item asChild>
									<Link to="/workspace/new" className="dropdown">
										<Text>Create or join a workspace</Text>
									</Link>
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</Grid>

					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<IconButton variant="solid" size="1">
								<Plus />
							</IconButton>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content size="2" align="start">
							<DropdownMenu.Item
								onSelect={() => app.openDialog("importItem")}
								shortcut="c"
								className="dropdown"
							>
								Import a link/document
							</DropdownMenu.Item>
							<DropdownMenu.Item
								onSelect={() => app.openDialog("createCollection")}
								shortcut="⌥ c"
								className="dropdown"
							>
								Create collection
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Flex>
			</Show>
		</>
	);
}
