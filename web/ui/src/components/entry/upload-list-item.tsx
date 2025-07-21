import { toHumanReadableSize } from "$/lib/utils";
import stores from "$/stores";
import type { FileEntry, LinkEntry, NewEntry } from "$/stores/uploads";
import { X } from "@phosphor-icons/react";
import {
	Card,
	Flex,
	Grid,
	IconButton,
	Link,
	Skeleton,
	Text,
} from "@radix-ui/themes";
import { useMemo } from "react";
import { useSnapshot } from "valtio";
import FileIcon from "./file-icon";
import { inferType } from "$/lib/file";
import Favicon from "./favicon";

type Props = {
	id: string;
	entry: NewEntry;
};
export default function UploadListItem(props: Props) {
	return (
		<Card variant="surface">
			{props.entry.type === "link" ? (
				<LinkItem id={props.id} entry={props.entry} />
			) : (
				<FileItem id={props.id} entry={props.entry} />
			)}
		</Card>
	);
}

type LinkItemProps = {
	id: string;
	entry: LinkEntry;
};

function LinkItem({ id, entry }: LinkItemProps) {
	const app = useSnapshot(stores.app);
	const uploads = useSnapshot(stores.uploads);

	const isLoading = useMemo(() => entry.status === "loading", [entry.status]);

	return (
		<Flex gap="3" align="start">
			<Skeleton loading={isLoading}>
				<Favicon url={entry.favicon} />
			</Skeleton>

			<Flex direction="column" flexGrow="1">
				<Grid width="100%" gap="1">
					<Skeleton loading={isLoading}>
						<Text weight="medium" className="truncate" align="left">
							{entry.title || entry.link}
						</Text>
					</Skeleton>
					<Skeleton loading={isLoading}>
						<Text size="1" color="gray" className="truncate" align="left">
							<Link
								href={entry.link}
								color={app.accentColor}
								underline="always"
								target="_blank"
							>
								{entry.domain}
							</Link>
							{entry.site_type && entry.domain ? " - " : ""}
							{entry.site_type ? (
								<span className="capitalize">{entry.site_type}</span>
							) : null}
						</Text>
					</Skeleton>
				</Grid>
			</Flex>

			<IconButton
				variant="ghost"
				size="1"
				color="gray"
				onClick={() => uploads.remove(id)}
			>
				<X className="text-[var(--accent-indicator)]" size={18} />
			</IconButton>
		</Flex>
	);
}

type FileItemProps = {
	id: string;
	entry: FileEntry;
};
function FileItem({ id, entry }: FileItemProps) {
	const uploads = useSnapshot(stores.uploads);
	const file = useMemo(
		() => uploads.files.find((f) => f.name === entry.name),
		[uploads.files, entry.name],
	);
	const fileType = useMemo(() => inferType(file?.type ?? ""), [file?.type]);
	return (
		<Flex gap="3" align="start">
			<Flex
				align="center"
				justify="center"
				className="aspect-square size-6 rounded-md"
			>
				<FileIcon type={file?.type ?? ""} />
			</Flex>

			<Flex direction="column" flexGrow="1">
				<Grid width="100%" gap="1">
					<Text weight="medium" className="truncate" align="left">
						{entry.name}
					</Text>
					<Text size="1" color="gray" className="truncate" align="left">
						{toHumanReadableSize(file?.size ?? 0)}
						{" - "}
						<Text as="span" weight="medium">
							{fileType}
						</Text>
					</Text>
				</Grid>
			</Flex>

			<IconButton
				variant="ghost"
				size="1"
				color="gray"
				onClick={() => uploads.remove(id)}
			>
				<X className="text-[var(--accent-indicator)]" size={18} />
			</IconButton>
		</Flex>
	);
}
