import type { AccentColor } from "$/stores/app";
import {
	FileAudio,
	FileJpg,
	FilePdf,
	Files,
	FileVideo,
	type Icon,
	Link,
} from "@phosphor-icons/react";
import { Flex, Grid } from "@radix-ui/themes";

export default function DropIcon() {
	return (
		<Grid columns="3" gap="3" position="relative">
			<EntryIcon color="blue" icon={FilePdf} />
			<EntryIcon color="grass" icon={Link} />
			<EntryIcon color="red" icon={FileJpg} />
			<EntryIcon color="teal" icon={FileVideo} />
			<EntryIcon color="yellow" icon={Files} />
			<EntryIcon color="plum" icon={FileAudio} />
		</Grid>
	);
}

type EntryProps = {
	color: AccentColor;
	icon: Icon;
};

function EntryIcon(props: EntryProps) {
	return (
		<Flex
			align="center"
			justify="center"
			className="aspect-square size-8 rounded-full border"
			style={{
				background: `var(--${props.color}-2)`,
				borderColor: `var(--${props.color}-10)`,
			}}
		>
			<props.icon size={16} style={{ color: `var(--${props.color}-10)` }} />
		</Flex>
	);
}
