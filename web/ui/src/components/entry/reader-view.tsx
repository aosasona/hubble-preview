import type { Entry } from "$/lib/server/types";
import Markdown from "react-markdown";
import {
	Text,
	Container,
	Flex,
	Heading,
	ScrollArea,
	Theme,
} from "@radix-ui/themes";
import { Clock } from "@phosphor-icons/react";
import { readingTime } from "reading-time-estimator";
import { CUSTOM_COMPONENTS } from "$/components/entry/custom-components";
import type { Metadata } from "$/lib/server/types";
import Favicon from "$/components/entry/favicon";
import { memo, useMemo } from "react";
import type { UIScale } from "$/stores/app";

type Props = {
	entry: Entry;
	scaling: UIScale;
};
function RawReaderView({ entry, scaling }: Props) {
	const readtime = useMemo(() => {
		const stat = readingTime(entry?.content ?? "");
		return stat.minutes;
	}, [entry?.content]);

	return (
		<ScrollArea
			style={{
				flexGrow: 1,
				width: "100%",
				maxWidth: "100vw",
				height: "100%",
				minHeight: 0,
			}}
		>
			<Container
				align="center"
				size={{ initial: "1", sm: "2", lg: "3" }}
				px={{ initial: "5", sm: "8", xl: "5" }}
				pt={{ initial: "4", xs: "6", md: "8" }}
				pb="8"
			>
				<Flex direction="column" mb="4" gap="3">
					<Heading size={{ initial: "8", sm: "9" }} style={{ lineHeight: 1.2 }}>
						{entry.name}
					</Heading>
					<Flex align="center" gap="4">
						{entry.type === "link"
							? (() => {
									const meta = entry.metadata as Metadata;
									return (
										<Flex gap="1" align="center">
											<Favicon url={meta.favicon} />
											<Text color="gray" weight="medium">
												{meta.author || "Unknown"}
											</Text>
										</Flex>
									);
								})()
							: null}
						<Flex align="center" gap="1">
							<Clock size={16} color="var(--yellow-10)" />
							<Text color="gray">
								{readtime} minute{readtime > 1 ? "s" : ""}
							</Text>
						</Flex>
					</Flex>
				</Flex>

				<Theme scaling={scaling}>
					<Markdown components={CUSTOM_COMPONENTS}>{entry.content}</Markdown>
				</Theme>
			</Container>
		</ScrollArea>
	);
}

export default memo(RawReaderView);
