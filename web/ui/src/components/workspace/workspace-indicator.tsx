import stores from "$/stores";
import { Flex, Grid, Text, Theme } from "@radix-ui/themes";
import { useSnapshot } from "valtio";

export default function WorkspaceIndicator() {
	const workspace = useSnapshot(stores.workspace);

	return (
		<Theme accentColor={workspace.accentColor} className="w-full">
			<Flex align="center" gap="2">
				<Flex
					align="center"
					justify="center"
					className="aspect-square size-[22px] rounded-[var(--radius-2)] border border-[var(--accent-8)]/25 bg-[var(--accent-9)]/10"
				>
					<Text
						size="1"
						weight="bold"
						color={workspace.accentColor}
						className="truncate"
					>
						{workspace.activeWorkspace?.name?.charAt(0)?.toUpperCase()}
					</Text>
				</Flex>

				<Grid className="w-full text-left">
					<Text
						size="2"
						color="gray"
						weight="medium"
						className="truncate"
						highContrast
					>
						{workspace.activeWorkspace?.name}
					</Text>
				</Grid>
			</Flex>
		</Theme>
	);
}
