import { Box, DropdownMenu, Flex, Grid, Text } from "@radix-ui/themes";
import { toTitleCase, redactEmail } from "$/lib/utils";
import {
	CaretUpDown,
	Gear,
	Moon,
	Palette,
	SignOut,
	Sun,
} from "@phosphor-icons/react";
import { useSnapshot } from "valtio";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import stores from "$/stores";
import { useRobinMutation } from "$/lib/hooks";
import type { ColorScheme } from "$/stores/app";

export default function SidebarUser() {
	const navigate = useNavigate();
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const app = useSnapshot(stores.app);
	const auth = useSnapshot(stores.auth);

	const logoutMutation = useRobinMutation("auth.sign-out", {
		onSuccess: () => {
			auth.clear();
			return navigate({ to: "/auth/sign-in" });
		},
	});

	if (!auth.user) return null;

	return (
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				<button
					type="button"
					className="w-full select-none rounded-md p-2 hover:bg-[var(--accent-9)]/10 focus:outline-none focus:ring-2 focus:ring-[var(--accent-9)]"
				>
					<Flex width="100%" align="center" gap="3">
						<Box className="rounded-full bg-[var(--accent-9)]/10" p="2">
							<Text size="2" weight="bold" color={app.accentColor}>
								{/* Render image if avatar is available */}
								{auth.user?.first_name?.charAt(0)?.toUpperCase()}
								{auth.user?.last_name?.charAt(0)?.toUpperCase()}
							</Text>
						</Box>

						<Flex width="100%" align="center" justify="between" gap="1">
							<Grid>
								<Text weight="medium" className="truncate" align="left">
									{toTitleCase(
										`${auth.user?.first_name} ${auth.user?.last_name}`,
									)}
								</Text>

								<Text
									size="1"
									weight="regular"
									color="gray"
									align="left"
									className="truncate"
								>
									{redactEmail(auth.user?.email, 1)}
								</Text>
							</Grid>

							<CaretUpDown size={16} className="text-[var(--accent-9)]" />
						</Flex>
					</Flex>
				</button>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content>
				<DropdownMenu.Label>Settings</DropdownMenu.Label>

				<DropdownMenu.Sub>
					<DropdownMenu.SubTrigger>
						<Palette />
						<Text>Theme</Text>
					</DropdownMenu.SubTrigger>
					<DropdownMenu.SubContent>
						<DropdownMenu.RadioGroup
							value={app.colorScheme}
							onValueChange={(value) =>
								stores.app.setColorScheme(value as ColorScheme)
							}
						>
							<DropdownMenu.RadioItem value="inherit">
								Automatic
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="light">
								<Sun />
								Light
							</DropdownMenu.RadioItem>
							<DropdownMenu.RadioItem value="dark">
								<Moon />
								Dark
							</DropdownMenu.RadioItem>
						</DropdownMenu.RadioGroup>
					</DropdownMenu.SubContent>
				</DropdownMenu.Sub>

				<DropdownMenu.Item shortcut="⌥ ." asChild>
					<Link
						to="/$workspaceSlug/settings"
						params={{ workspaceSlug: params.workspaceSlug }}
						className="dropdown"
					>
						<Gear /> Settings
					</Link>
				</DropdownMenu.Item>

				<DropdownMenu.Separator />

				<DropdownMenu.Item
					color="red"
					onClick={() => logoutMutation.call()}
					disabled={logoutMutation.isMutating}
				>
					<SignOut />
					Sign out
				</DropdownMenu.Item>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	);
}
