import stores from "$/stores";
import { House, MagnifyingGlass } from "@phosphor-icons/react";
import { Badge, Box, Button, Flex, Grid, Text } from "@radix-ui/themes";
import { Link, useParams } from "@tanstack/react-router";
import { useSnapshot } from "valtio";

export default function SidebarItems() {
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const app = useSnapshot(stores.app);
	const workspaces = useSnapshot(stores.workspace);

	return (
		<Box mt="1">
			<Flex gap="4" direction="column">
				<Flex gap="1" direction="column">
					<Link
						to="/$workspaceSlug"
						params={{ workspaceSlug: params.workspaceSlug }}
						className="sidebar-link"
					>
						<House size={18} /> Home
					</Link>
					<Link
						to="/$workspaceSlug/search"
						params={{ workspaceSlug: params.workspaceSlug }}
						className="sidebar-link"
					>
						<MagnifyingGlass size={18} /> Search
					</Link>
				</Flex>

				<Box>
					<Flex gap="1" align="center" justify="between">
						<Text size="2" color="gray">
							Collections
						</Text>
					</Flex>

					{workspaces.activeWorkspaceCollections.length === 0 ? (
						<Flex align="center" justify="center" mt="4">
							<Button
								onClick={() => app.openDialog("createCollection")}
								style={{ flexGrow: 1 }}
							>
								Create a collection
							</Button>
						</Flex>
					) : (
						<Flex direction="column" gap="1" mt="1">
							{workspaces.activeWorkspaceCollections.map((collection) => (
								<Link
									key={collection.id}
									to="/$workspaceSlug/c/$collectionSlug"
									params={{
										workspaceSlug: params.workspaceSlug ?? "",
										collectionSlug: collection.slug,
									}}
									activeOptions={{
										exact: true,
										includeSearch: false,
										includeHash: false,
									}}
									className="collection-sidebar-link"
								>
									{({ isActive }) => (
										<Flex gap="2" align="center" justify="between">
											<Grid>
												<Text size="2" truncate>
													{collection.name}
												</Text>
											</Grid>

											<Badge size="1" variant={isActive ? "surface" : "soft"}>
												{collection.entries_count}
											</Badge>
										</Flex>
									)}
								</Link>
							))}
						</Flex>
					)}
				</Box>
			</Flex>
		</Box>
	);
}
