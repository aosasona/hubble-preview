import PageLayout from "$/components/layout/page-layout";
import stores from "$/stores";
import { Check } from "@phosphor-icons/react";
import {
	Text,
	Checkbox,
	Flex,
	Heading,
	Box,
	DropdownMenu,
	Button,
} from "@radix-ui/themes";
import { createLazyFileRoute } from "@tanstack/react-router";
import { useSnapshot } from "valtio";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/",
)({
	component: RouteComponent,
});

function RouteComponent() {
	const workspaces = useSnapshot(stores.workspace);

	return (
		<PageLayout heading="General" header={{ parent: "settings" }} showHeader>
			<Flex direction="column" mt="6" gap="6">
				<Box>
					<Text as="label" size="2">
						<Flex gap="2">
							<Checkbox
								defaultChecked={workspaces.shouldSaveActiveWorkspace}
								onCheckedChange={workspaces.toggleSaveActiveWorkspace}
							/>
							Save active workspace
						</Flex>
					</Text>

					<Text size="1" color="gray">
						Automatically save the active workspace when you close the tab or
						refresh the page.
					</Text>
				</Box>

				<Box>
					<Heading size="2" color="gray" mb="2">
						Default workspace
					</Heading>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<Button variant="surface">
								{workspaces.findWorkspaceById(
									workspaces?.defaultWorkspaceId ?? "",
								)?.name ?? "Select a workspace"}
								<DropdownMenu.TriggerIcon />
							</Button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content>
							{workspaces.all.map((workspace) => (
								<DropdownMenu.Item
									key={workspace.id}
									onSelect={() => workspaces.setDefaultWorkspace(workspace.id)}
								>
									<Flex align="center" gap="2">
										{workspace.id === workspaces.defaultWorkspaceId ? (
											<Check />
										) : null}
										<Text>{workspace.name}</Text>
									</Flex>
								</DropdownMenu.Item>
							))}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Box>
			</Flex>
		</PageLayout>
	);
}
