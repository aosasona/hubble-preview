import stores from "$/stores";
import {
	CaretRight,
	Gear,
	type Icon,
	Palette,
	PuzzlePiece,
	Shield,
	User,
	Users,
} from "@phosphor-icons/react";
import {
	Box,
	Button,
	DropdownMenu,
	Flex,
	Heading,
	Text,
} from "@radix-ui/themes";
import {
	createFileRoute,
	Link,
	Outlet,
	useRouterState,
} from "@tanstack/react-router";
import { useMemo } from "react";
import { useSnapshot } from "valtio";

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings",
)({
	component: RouteComponent,
});

type Section = {
	icon?: Icon;
	label: string;
	link: string;
};

type Settings = Record<string, Section[]>;

function RouteComponent() {
	const state = useRouterState();
	const { workspaceSlug } = Route.useParams();

	const workspaces = useSnapshot(stores.workspace);

	const settings: Settings = useMemo(() => {
		const defaultSettings = {
			Application: [
				{
					icon: Gear,
					label: "General",
					link: `/${workspaceSlug}/settings`,
				},
				{
					icon: Palette,
					label: "Appearance",
					link: `/${workspaceSlug}/settings/appearance`,
				},
			],
			Account: [
				{
					icon: User,
					label: "Account",
					link: `/${workspaceSlug}/settings/account`,
				},
				{
					icon: Shield,
					label: "Security",
					link: `/${workspaceSlug}/settings/security`,
				},
			],
			Workspace: [
				{
					icon: Gear,
					label: "General",
					link: `/${workspaceSlug}/settings/workspace`,
				},
				{
					icon: Users,
					label: "Members",
					link: `/${workspaceSlug}/settings/workspace/members`,
				},
				{
					icon: PuzzlePiece,
					label: "Plugins",
					link: `/${workspaceSlug}/settings/workspace/plugins`,
				},
			],
		};

		const collections: Section[] = [];
		for (const collection of workspaces.activeWorkspaceCollections) {
			collections.push({
				label: collection.name,
				link: `/${workspaceSlug}/settings/workspace/c/${collection.slug}`,
			});
		}

		return {
			...defaultSettings,
			Collections: collections,
		};
	}, [workspaceSlug, workspaces.activeWorkspaceCollections]);

	const currentPath = useMemo(() => {
		let current: Section | undefined;

		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		for (const [_, items] of Object.entries(settings)) {
			for (const item of items) {
				if (item.link === state.location.pathname) {
					current = item;
					break;
				}
			}
		}

		return current;
	}, [settings, state.location.pathname]);

	return (
		<Flex
			position={{ initial: "relative", md: "static" }}
			direction={{ initial: "column", md: "row" }}
			width="100%"
			height={{ initial: "auto", md: "100%" }}
			minHeight="0"
			flexGrow="1"
		>
			{/* Desktop */}
			<Box
				display={{ initial: "none", md: "block" }}
				minWidth={{ md: "200px", lg: "250px" }}
				width={{ md: "200px", lg: "250px" }}
				p="3"
				className="border-r border-r-[var(--gray-3)]"
			>
				<Flex direction="column" gap="4" mt="2">
					{Object.entries(settings).map(([section, links]) => (
						<Flex direction="column" gap="2" key={section}>
							<Heading size="2" weight="regular" color="gray">
								{section}
							</Heading>

							<Flex direction="column" gap="1">
								{links.map(({ icon: Icon, label, link }) => (
									<Link
										key={link}
										to={link}
										className="inner-sidebar-link"
										activeOptions={{
											exact: true,
											includeSearch: false,
											includeHash: false,
										}}
									>
										<Flex align="center" gap="2">
											{Icon ? <Icon size={18} /> : null}
											{label}
										</Flex>
									</Link>
								))}
							</Flex>
						</Flex>
					))}
				</Flex>
			</Box>

			{/* Mobile */}
			<Box
				display={{ initial: "block", md: "none" }}
				position="sticky"
				left="0"
				right="0"
				py="2"
				px="3"
				className="top-[37px] z-[9998] border-b border-b-[var(--gray-3)] bg-background"
			>
				<Flex gap="2" align="center">
					<Text size="2" color="gray">
						Settings
					</Text>

					<Box>
						<CaretRight size={12} className="text-[var(--gray-10)]" />
					</Box>

					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<Button
								color="gray"
								variant="surface"
								style={{ flexGrow: 1, maxWidth: "250px" }}
							>
								<Flex align="center" justify="between" width="100%" gap="4">
									<Flex width="160px" align="center" gap="2">
										{currentPath?.icon && <currentPath.icon size={16} />}
										<Text size="2" weight="medium" className="truncate">
											{currentPath?.label ?? "Settings"}
										</Text>
									</Flex>

									<Box>
										<DropdownMenu.TriggerIcon />
									</Box>
								</Flex>
							</Button>
						</DropdownMenu.Trigger>

						<DropdownMenu.Content>
							{Object.entries(settings).map(([section, links]) => (
								<DropdownMenu.Group key={section}>
									<DropdownMenu.Label>{section}</DropdownMenu.Label>
									{links.map(({ icon: Icon, label, link }) => (
										<DropdownMenu.Item key={link} asChild>
											<Link to={link} className="dropdown">
												{Icon ? <Icon size={18} /> : null}
												{label}
											</Link>
										</DropdownMenu.Item>
									))}
								</DropdownMenu.Group>
							))}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Flex>
			</Box>

			<Outlet />
		</Flex>
	);
}
