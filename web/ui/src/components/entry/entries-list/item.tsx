import type React from "react";
import { useMemo, useState } from "react";
import { timeAgo } from "$/lib/date";
import type { Entry, Metadata } from "$/lib/server/types";
import stores from "$/stores";
import type { AccentColor } from "$/stores/app";
import { Clock } from "@phosphor-icons/react";
import { Badge, Box, Checkbox, Flex, Grid, Text } from "@radix-ui/themes";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import type { Virtualizer } from "@tanstack/react-virtual";
import { AnimatePresence, motion } from "motion/react";
import { useSnapshot } from "valtio";
import EntryIcon from "../entry-icon";
import Favicon from "../favicon";
import { deselectEntry, selectEntry } from "$/lib/entry-shortcuts";
import { calculateColorHash } from "$/stores/workspace";

type Props = {
	entry: Entry;
	index: number;
	activeCollectionSlug?: string;
	measureElement: Virtualizer<HTMLDivElement, Element>["measureElement"];
};

export default function Item({
	entry,
	index,
	activeCollectionSlug,
	measureElement,
}: Props) {
	const navigate = useNavigate();
	const selections = useSnapshot(stores.entriesList.selections);

	const [hovered, setHovered] = useState(false);
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const selected = selections.has(entry.id);

	const AnimatedDiv = ({
		children,
	}: {
		children: React.ReactNode;
	}) => {
		return (
			<motion.div
				key={selected || hovered ? "checkable" : "normal"}
				initial={{ opacity: 0 }}
				animate={{ opacity: 1 }}
				exit={{ opacity: 0 }}
				transition={{ duration: 0.1 }}
				className="flex items-center justify-center"
			>
				{children}
			</motion.div>
		);
	};

	function onSelect(e: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
		e.preventDefault();
		e.stopPropagation();
		const checked = (e.currentTarget.dataset?.state ?? "") !== "checked";

		if (checked) {
			selectEntry(entry.id);
		} else {
			deselectEntry(entry.id);
		}
	}

	const collectionAccentColor = useMemo(() => {
		return calculateColorHash(entry.collection.slug);
	}, [entry.collection.slug]);

	// TODO: add context menu for desktop
	// TODO: add menu for mobile
	return (
		<Link
			to="/$workspaceSlug/c/$collectionSlug/$entryId"
			params={{
				workspaceSlug: params.workspaceSlug,
				collectionSlug: entry.collection.slug,
				entryId: entry.id,
			}}
			className="unstyled entry-list-item flex h-full w-full flex-grow flex-col"
			ref={measureElement}
			data-index={index}
			data-list-item={entry.id}
			data-selected={selected}
		>
			<Flex
				direction="column"
				flexGrow="1"
				height="100%"
				gap="2"
				py="3"
				px="3"
				onMouseEnter={() => setHovered(true)}
				onMouseLeave={() => setHovered(false)}
			>
				<Flex justify="between" align="center">
					<Flex align="center" flexGrow="1" gap="1">
						<AnimatePresence>
							<Flex align="center" justify="center" className="size-6">
								{hovered || selected ? (
									<AnimatedDiv>
										<Checkbox
											data-checkbox
											size="2"
											checked={selected}
											onClick={(e) => onSelect(e)}
										/>
									</AnimatedDiv>
								) : entry.type === "link" ? (
									<AnimatedDiv>
										<Favicon
											url={(entry.metadata as Metadata)?.favicon ?? ""}
										/>
									</AnimatedDiv>
								) : (
									<AnimatedDiv>
										<EntryIcon type={entry.type} />
									</AnimatedDiv>
								)}
							</Flex>
						</AnimatePresence>
						<Grid m="0" p="0">
							<Text size="2" weight="medium" className="truncate">
								{entry.name}
							</Text>
						</Grid>
					</Flex>

					{entry.status !== "completed" ? (
						<QueueStatus status={entry.status} />
					) : null}
				</Flex>

				<Flex px="1" gap="2" align="center">
					{!activeCollectionSlug ? (
						<Badge
							size="1"
							color={collectionAccentColor}
							radius="large"
							variant="surface"
							onClick={(e) => {
								e.preventDefault();
								e.stopPropagation();
								navigate({
									to: "/$workspaceSlug/c/$collectionSlug",
									params: {
										workspaceSlug: params.workspaceSlug,
										collectionSlug: entry.collection.slug,
									},
								});
							}}
						>
							<Grid>
								<Text className="truncate" size="1">
									{entry.collection.name}
								</Text>
							</Grid>
						</Badge>
					) : null}

					<Flex gap="1">
						<Clock size={15} color="var(--gray-11)" />
						<Text size="1" color="gray" m="0">
							{timeAgo(entry.created_at)}
						</Text>
					</Flex>
				</Flex>
			</Flex>

			{/* MARK: border */}
			<Box ml="3" mt="auto">
				<Box className="item-divider" />
			</Box>
		</Link>
	);
}

export function QueueStatus({ status }: { status: Entry["status"] }) {
	let color: AccentColor = "gray";

	switch (status) {
		case "queued":
			color = "brown";
			break;
		case "processing":
			color = "amber";
			break;
		case "completed":
			color = "green";
			break;
		case "failed":
			color = "red";
			break;
		case "paused":
			color = "bronze";
			break;
		case "canceled":
			color = "gray";
			break;
	}

	return (
		<Badge
			className="capitalize"
			size="1"
			variant="outline"
			radius="full"
			color={color}
		>
			<Text size="1" color={color}>
				{status}
			</Text>
		</Badge>
	);
}
