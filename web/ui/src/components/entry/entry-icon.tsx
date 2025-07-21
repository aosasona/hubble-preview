import {
	Chat,
	File,
	FileAudio,
	FileCode,
	FileCsv,
	FileDoc,
	FileHtml,
	FileImage,
	FileMd,
	FilePdf,
	FilePpt,
	Files,
	FileText,
	FileVideo,
	Link,
	type Icon as PhosphorIcon,
} from "@phosphor-icons/react";
import { useMemo } from "react";
import { Flex } from "@radix-ui/themes";
import type { AccentColor } from "$/stores/app";
import type { Entry } from "$/lib/server/types";

type EntryType = Entry["type"];
type Props = {
	type: EntryType;
};

const ICON_MAP: Record<EntryType, [AccentColor, PhosphorIcon]> = {
	pdf: ["tomato", FilePdf],
	image: ["teal", FileImage],
	video: ["amber", FileVideo],
	audio: ["plum", FileAudio],
	markdown: ["gold", FileMd],
	presentation: ["red", FilePpt],
	code: ["yellow", FileCode],
	word_document: ["blue", FileDoc],
	archive: ["red", Files],
	other: ["gray", File],
	link: ["grass", Link],
	html: ["indigo", FileHtml],
	interchange: ["brown", FileCode],
	epub: ["purple", File],
	spreadsheet: ["cyan", FileCsv],
	plain_text: ["gray", FileText],
	comment: ["sky", Chat],
} as const;

export default function EntryIcon(props: Props) {
	const color = useMemo(
		() => ICON_MAP[props.type]?.[0] ?? "gray",
		[props.type],
	);
	const Icon = useMemo(() => ICON_MAP[props.type]?.[1] ?? Files, [props.type]);

	return (
		<Flex align="center" justify="center" className="aspect-square size-6">
			<Icon size={20} style={{ color: `var(--${color}-10)` }} />
		</Flex>
	);
}
