import {
	Books,
	CheckCircle,
	Clock,
	ClockClockwise,
	ClockCounterClockwise,
	Code,
	FileArchive,
	FileCsv,
	FileDoc,
	FileHtml,
	FilePdf,
	HourglassHigh,
	Image,
	Link,
	MarkdownLogo,
	MusicNoteSimple,
	Pause,
	Presentation,
	Question,
	SortAscending,
	SortDescending,
	TextAa,
	TreeStructure,
	Video,
	WarningCircle,
	XCircle,
	type Icon as PhosphorIcon,
} from "@phosphor-icons/react";
import type { Entry } from "$/lib/server/types";

export type Mapping = { label: string; icon: PhosphorIcon | React.FC<unknown> };

export const ENTRY_TYPES = [
	"link",
	"audio",
	"video",
	"image",
	"pdf",
	"interchange",
	"epub",
	"word_document",
	"presentation",
	"spreadsheet",
	"html",
	"markdown",
	"plain_text",
	"archive",
	"code",
	"other",
] as const;

export const ENTRY_TYPE_MAPPING: Record<EntryType, Mapping> = {
	link: { label: "Link", icon: Link },
	audio: { label: "Audio", icon: MusicNoteSimple },
	video: { label: "Video", icon: Video },
	image: { label: "Image", icon: Image },
	pdf: { label: "PDF", icon: FilePdf },
	interchange: { label: "Interchange", icon: TreeStructure },
	epub: { label: "EPUB", icon: Books },
	word_document: { label: "Word Document", icon: FileDoc },
	presentation: { label: "Presentation", icon: Presentation },
	spreadsheet: { label: "Spreadsheet", icon: FileCsv },
	html: { label: "Markup", icon: FileHtml },
	markdown: { label: "Markdown", icon: MarkdownLogo },
	plain_text: { label: "Plain Text", icon: TextAa },
	archive: { label: "Archive", icon: FileArchive },
	code: { label: "Code", icon: Code },
	other: { label: "Others", icon: Question },
};

export const STATUS_MAPPING: Record<Entry["status"], Mapping> = {
	queued: { label: "Queued", icon: HourglassHigh },
	processing: { label: "Processing", icon: Clock },
	completed: { label: "Completed", icon: CheckCircle },
	failed: { label: "Failed", icon: WarningCircle },
	canceled: { label: "Canceled", icon: XCircle },
	paused: { label: "Paused", icon: Pause },
};

export type EntryType = (typeof ENTRY_TYPES)[number];

export type Filter =
	| { type: "type"; value: EntryType }
	| { type: "created_at"; value: string }
	| { type: "last_updated_at"; value: string }
	| { type: "status"; value: Entry["status"] }
	| { type: "collection"; value: string };

export const CATEGORY_MAPPING: Record<Filter["type"], string> = {
	type: "Type",
	created_at: "Created At",
	last_updated_at: "Last Updated At",
	status: "Status",
	collection: "Collection",
};

export const SORT_BY = [
	"default",
	"newest",
	"oldest",
	"name_asc",
	"name_desc",
	"collection_asc",
	"collection_desc",
] as const;

export type SortBy = (typeof SORT_BY)[number];

export const SORT_BY_MAPPING: Record<SortBy, Mapping> = {
	default: { label: "Default", icon: Clock },
	newest: { label: "Newest", icon: ClockClockwise },
	oldest: { label: "Oldest", icon: ClockCounterClockwise },
	name_asc: { label: "Name (A-Z)", icon: SortAscending },
	name_desc: { label: "Name (Z-A)", icon: SortDescending },
	collection_asc: { label: "Collection (A-Z)", icon: SortAscending },
	collection_desc: { label: "Collection (Z-A)", icon: SortDescending },
};
