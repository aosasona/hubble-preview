import {
	File,
	FileAudio,
	FileCode,
	FileDoc,
	FileHtml,
	FileImage,
	FileJs,
	FileMd,
	FilePdf,
	FilePpt,
	Files,
	FileText,
	FileVideo,
	FileZip,
	type Icon as PhosphorIcon,
} from "@phosphor-icons/react";
import { useMemo } from "react";
import { type FileType, inferType } from "$/lib/file";
import { Flex } from "@radix-ui/themes";
import type { AccentColor } from "$/stores/app";

type Props = {
	type: string;
};
const ICON_MAP: Partial<Record<FileType, [AccentColor, PhosphorIcon]>> = {
	PDF: ["tomato", FilePdf],
	Image: ["teal", FileImage],
	Video: ["yellow", FileVideo],
	Audio: ["plum", FileAudio],
	Markdown: ["teal", FileMd],
	Presentation: ["red", FilePpt],
	JSON: ["yellow", FileCode],
	HTML: ["blue", FileHtml],
	JavaScript: ["yellow", FileJs],
	Text: ["gray", FileText],
	Document: ["gold", FileDoc],
	ZIP: ["red", FileZip],
	Unknown: ["gray", File],
} as const;

export default function FileIcon(props: Props) {
	const fileType = useMemo(() => inferType(props.type), [props.type]);
	const color = useMemo(() => ICON_MAP[fileType]?.[0] ?? "gray", [fileType]);
	const Icon = useMemo(() => ICON_MAP[fileType]?.[1] ?? Files, [fileType]);

	return (
		<Flex align="center" justify="center" className="aspect-square size-6">
			<Icon size={20} style={{ color: `var(--${color}-10)` }} />
		</Flex>
	);
}
