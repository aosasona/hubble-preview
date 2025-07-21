import type { FileMetadata, Metadata } from "$/lib/server/types";
import { toHumanReadableSize } from "$/lib/utils";
import {
	Badge,
	Box,
	DataList,
	Flex,
	Heading,
	Separator,
	Skeleton,
	Text,
	Link as ThemedLink,
} from "@radix-ui/themes";
import { Link, useParams } from "@tanstack/react-router";
import { AnimatePresence, motion } from "motion/react";
import { QueueStatus } from "./item";
import { calculateColorHash } from "$/stores/workspace";
import { Archive, ImageBroken } from "@phosphor-icons/react";
import { formatDateString } from "$/lib/date";
import { useEffect, useState } from "react";
import { useSnapshot } from "valtio";
import stores from "$/stores";

type ImageState = "loading" | "error" | "success";

export default function PreviewPane() {
	const entriesList = useSnapshot(stores.entriesList);

	const params = useParams({ from: "/_protected/$workspaceSlug" });
	const [imageState, setImageState] = useState<ImageState>("loading");

	useEffect(() => {
		setImageState("loading");

		if (
			entriesList?.preview?.type === "link" &&
			!(entriesList?.preview?.metadata as Metadata)?.thumbnail.trim()
		) {
			setImageState("error");
		}
	}, [entriesList?.preview?.metadata, entriesList?.preview?.type]);

	return (
		<AnimatePresence>
			{!entriesList.preview ? null : (
				<motion.div
					initial={{ width: 0 }}
					animate={{
						width: "400px",
					}}
					exit={{ width: 0 }}
					transition={{ duration: 0.125, type: "tween" }}
					className="flex h-full min-h-0 flex-shrink-0"
				>
					<Flex
						direction="column"
						className="border-[var(--gray-4)] border-l"
						gap="3"
						px="4"
						py="4"
						minHeight="0"
						flexGrow="1"
						overflowY="scroll"
						width="100%"
						minWidth="375px"
						style={{ whiteSpace: "normal" }}
					>
						{!entriesList.preview.archived_at?.startsWith("0001") ? (
							<Box mb="2">
								<Badge color="gray" variant="soft">
									<Flex align="center" gap="1">
										<Archive color="var(--gray-9)" size={14} />
										<Text>Archived</Text>
									</Flex>
								</Badge>
							</Box>
						) : null}

						<Flex align="center" gap="2" justify="between">
							<Link
								to="/$workspaceSlug/c/$collectionSlug"
								params={{
									workspaceSlug: params.workspaceSlug,
									collectionSlug: entriesList.preview.collection.slug,
								}}
							>
								<Badge
									variant="surface"
									color={calculateColorHash(
										entriesList.preview.collection.name,
									)}
								>
									{entriesList.preview.collection.name}
								</Badge>
							</Link>

							{entriesList.preview.status ? (
								<QueueStatus status={entriesList.preview.status} />
							) : null}
						</Flex>

						{entriesList.preview.type === "link" ? (
							<Skeleton loading={imageState === "loading"}>
								<Flex>
									{imageState === "error" ||
									!(
										entriesList.preview.metadata as Metadata
									).thumbnail?.trim() ? (
										<Flex
											width="100%"
											justify="center"
											align="center"
											className="aspect-video rounded-[var(--radius-2)] bg-[var(--gray-2)]"
										>
											<ImageBroken color="var(--gray-9)" size={24} />
										</Flex>
									) : (
										<img
											src={(entriesList.preview.metadata as Metadata).thumbnail}
											alt={entriesList.preview.name}
											onError={() => setImageState("error")}
											onLoad={() => setImageState("success")}
											className="aspect-video w-full rounded-[var(--radius-2)] bg-[var(--gray-2)] object-cover"
										/>
									)}
								</Flex>
							</Skeleton>
						) : null}

						<Heading size="6" weight="bold">
							{entriesList.preview.name}
						</Heading>

						<Separator
							orientation="horizontal"
							my="1"
							style={{ width: "100%" }}
						/>

						{entriesList.preview.type === "link" ? (
							<LinkProperties
								metadata={entriesList.preview.metadata as Metadata}
							/>
						) : (
							<FileProperties
								metadata={entriesList.preview.metadata as FileMetadata}
							/>
						)}
					</Flex>
				</motion.div>
			)}
		</AnimatePresence>
	);
}

function LinkProperties({ metadata }: { metadata: Metadata }) {
	const { preview: entry } = useSnapshot(stores.entriesList);
	if (!entry) return null;

	return (
		<Flex direction="column" gap="3">
			<ThemedLink
				size="2"
				underline="always"
				target="_blank"
				href={metadata.link}
			>
				<Text>{metadata.link}</Text>
			</ThemedLink>

			<DataList.Root>
				<DataList.Item>
					<DataList.Label>Title</DataList.Label>
					<DataList.Value>{metadata.title || entry.name}</DataList.Value>
				</DataList.Item>
				<DataList.Item>
					<DataList.Label>Author</DataList.Label>
					<DataList.Value>
						{metadata.author.trim() ? (
							<Text>{metadata.author}</Text>
						) : (
							<Text color="gray">Unknown</Text>
						)}
					</DataList.Value>
				</DataList.Item>
				{metadata.domain.trim() ? (
					<DataList.Item>
						<DataList.Label>Domain</DataList.Label>
						<DataList.Value>
							<ThemedLink href={metadata.domain}>{metadata.domain}</ThemedLink>
						</DataList.Value>
					</DataList.Item>
				) : null}
				<DataList.Item>
					<DataList.Label>Type</DataList.Label>
					<DataList.Value>
						<Badge color="gray" variant="surface">
							{metadata.site_type}
						</Badge>
					</DataList.Value>
				</DataList.Item>
				<DataList.Item>
					<DataList.Label>Description</DataList.Label>
					<DataList.Value>
						{metadata.description.trim() ? (
							<Text>{metadata.description}</Text>
						) : (
							<Text color="gray">No description</Text>
						)}
					</DataList.Value>
				</DataList.Item>
				{entry.added_by ? (
					<DataList.Item>
						<DataList.Label>Added By</DataList.Label>
						<DataList.Value>
							{entry.added_by?.first_name} {entry.added_by?.last_name}
						</DataList.Value>
					</DataList.Item>
				) : null}
				<DataList.Item>
					<DataList.Label>Created</DataList.Label>
					<DataList.Value>{formatDateString(entry.created_at)}</DataList.Value>
				</DataList.Item>
			</DataList.Root>
		</Flex>
	);
}

function FileProperties({ metadata }: { metadata: FileMetadata }) {
	const { preview: entry } = useSnapshot(stores.entriesList);
	if (!entry) return null;

	return (
		<Flex direction="column" gap="2">
			<DataList.Root size="2">
				<DataList.Item>
					<DataList.Label>Filename</DataList.Label>
					<DataList.Value>{metadata.original_filename}</DataList.Value>
				</DataList.Item>
				<DataList.Item>
					<DataList.Label>Size</DataList.Label>
					<DataList.Value>
						{toHumanReadableSize(entry.filesize_bytes)}
					</DataList.Value>
				</DataList.Item>
				<DataList.Item>
					<DataList.Label>MIME Type</DataList.Label>
					<DataList.Value>{metadata.mime_type}</DataList.Value>
				</DataList.Item>
				<DataList.Item>
					<DataList.Label>Category</DataList.Label>
					<DataList.Value>
						<Badge color="gray" variant="surface">
							{entry.type?.split("_")?.join(" ") ?? "unknown"}
						</Badge>
					</DataList.Value>
				</DataList.Item>
				{entry.version ? (
					<DataList.Item>
						<DataList.Label>Version</DataList.Label>
						<DataList.Value>{entry.version}</DataList.Value>
					</DataList.Item>
				) : null}
				{entry.added_by ? (
					<DataList.Item>
						<DataList.Label>Added By</DataList.Label>
						<DataList.Value>
							{entry.added_by?.first_name} {entry.added_by?.last_name}
						</DataList.Value>
					</DataList.Item>
				) : null}
				<DataList.Item>
					<DataList.Label>Created</DataList.Label>
					<DataList.Value>{formatDateString(entry.created_at)}</DataList.Value>
				</DataList.Item>
			</DataList.Root>
		</Flex>
	);
}
