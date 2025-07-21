import PageLayout from "$/components/layout/page-layout";
import { toTitleCase } from "$/lib/utils";
import stores from "$/stores";
import {
	ACCENT_COLORS,
	type AccentColor,
	type ColorScheme,
	UI_SCALES,
	type UIScale,
} from "$/stores/app";
import { Moon, Sun } from "@phosphor-icons/react";
import {
	Box,
	Text,
	SegmentedControl,
	Flex,
	Button,
	DropdownMenu,
	Heading,
	Theme,
	RadioCards,
} from "@radix-ui/themes";
import { createLazyFileRoute } from "@tanstack/react-router";
import { memo } from "react";
import { useSnapshot } from "valtio";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/appearance",
)({
	component: RouteComponent,
});

function RouteComponent() {
	const app = useSnapshot(stores.app);

	return (
		<PageLayout heading="Appearance" header={{ parent: "settings" }} showHeader>
			<Flex direction="column" gap="5" mt="4">
				<Box>
					<Heading color="gray" size="2" weight="medium">
						Color scheme
					</Heading>
					<Box mt="2">
						<SegmentedControl.Root
							value={app.colorScheme}
							onValueChange={(v) => app.set("colorScheme", v as ColorScheme)}
							size="2"
						>
							<SegmentedControl.Item value="light">
								<Flex align="center" gap="2">
									<Sun size={14} />
									<Text size="1" color="gray" highContrast>
										Light
									</Text>
								</Flex>
							</SegmentedControl.Item>
							<SegmentedControl.Item value="dark">
								<Flex align="center" gap="2">
									<Moon size={14} />
									<Text size="1" color="gray" highContrast>
										Dark
									</Text>
								</Flex>
							</SegmentedControl.Item>
							<SegmentedControl.Item value="inherit">
								Auto
							</SegmentedControl.Item>
						</SegmentedControl.Root>
					</Box>
				</Box>

				<Flex width="fit-content" direction="column" gap="2">
					<Heading color="gray" size="2" weight="medium">
						Theme
					</Heading>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<Button color="gray" variant="surface">
								<Flex align="center" justify="between" gap="6">
									<ColorListItem color={app.accentColor} />
									<DropdownMenu.TriggerIcon />
								</Flex>
							</Button>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content variant="soft" color="gray" size="2">
							{ACCENT_COLORS.map((color) => (
								<DropdownMenu.Item
									key={color}
									onSelect={() => app.set("accentColor", color)}
								>
									<ColorListItem
										color={color}
										selected={app.accentColor === color}
									/>
								</DropdownMenu.Item>
							))}
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Flex>

				<Flex direction="column" gap="2">
					<Heading color="gray" size="2" weight="medium">
						Scaling
					</Heading>
					<RadioCards.Root
						defaultValue={app.uiScale}
						onValueChange={(value) => app.set("uiScale", value as UIScale)}
						columns={{ initial: "2", xs: "3", lg: "4", xl: "5" }}
					>
						{UI_SCALES.map((scale) => (
							<ScaleListItem key={scale} scale={scale} />
						))}
					</RadioCards.Root>
				</Flex>
			</Flex>
		</PageLayout>
	);
}

type ScaleListItemProps = {
	scale: UIScale;
};
function ScaleListItem(props: ScaleListItemProps) {
	const appState = useSnapshot(stores.app);
	return (
		<RadioCards.Item value={props.scale}>
			<Theme
				accentColor={appState.accentColor}
				appearance={appState.colorScheme}
				panelBackground="translucent"
				radius="large"
				scaling={props.scale}
				grayColor={appState.grayColor}
				hasBackground={false}
			>
				<Flex width="100%" direction="column" gap="2">
					<Heading size="2">Heading ({props.scale})</Heading>
					<Text size="1" color="gray">
						This is a sample text.
					</Text>
				</Flex>
			</Theme>
		</RadioCards.Item>
	);
}

type ColorListItemProps = {
	color: AccentColor;
	selected?: boolean;
};
function RawColorListItem(props: ColorListItemProps) {
	return (
		<Flex align="center" gap="2">
			<Box
				className="aspect-square w-4 rounded-[var(--radius-1)]"
				style={{
					backgroundColor: `var(--${props.color}-9)`,
				}}
			/>
			<Text
				weight={props.selected ? "bold" : "regular"}
				color={props.selected ? props.color : undefined}
			>
				{toTitleCase(props.color)}
			</Text>
		</Flex>
	);
}

const ColorListItem = memo(RawColorListItem);
