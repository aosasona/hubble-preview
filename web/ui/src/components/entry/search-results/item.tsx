import type React from "react";
import { useMemo, useState } from "react";
import { timeAgo } from "$/lib/date";
import type { CollapsedSearchResult, Metadata } from "$/lib/server/types";
import stores from "$/stores";
import { Clock } from "@phosphor-icons/react";
import {
	Badge,
	Box,
	Checkbox,
	Flex,
	Grid,
	Heading,
	Popover,
	Progress,
	Text,
	Tooltip,
} from "@radix-ui/themes";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import type { Virtualizer } from "@tanstack/react-virtual";
import { AnimatePresence, motion } from "motion/react";
import { useSnapshot } from "valtio";
import EntryIcon from "../entry-icon";
import Favicon from "../favicon";
import { deselectEntry, selectEntry } from "$/lib/entry-shortcuts";
import { pluralize } from "$/lib/utils";
import { calculateColorHash } from "$/stores/workspace";
import type { AccentColor } from "$/stores/app";

type Props = {
	result: CollapsedSearchResult;
	currentQuery: string;
	serverQuery: string;
	index: number;
	scores: {
		minimum: number;
		maximum: number;
	};
	measureElement: Virtualizer<HTMLDivElement, Element>["measureElement"];
};

export default function Item({
	result,
	index,
	serverQuery,
	scores,
	measureElement,
}: Props) {
	const navigate = useNavigate();
	const selections = useSnapshot(stores.entriesList.selections);

	const [hovered, setHovered] = useState(false);
	const [popoverOpen, setPopoverOpen] = useState(false);
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const selected = selections.has(result.id);

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
			selectEntry(result.id);
		} else {
			deselectEntry(result.id);
		}
	}

	// Find the perfect window to show
	// Ideally, this is either the first 100-ish charcaters or 40 words to and from the first exact match
	const preview = useMemo(() => {
		const text = result.matches?.[0]?.text || "";

		if (!text) return "…";

		const windowSize = 240;
		const lowerText = text.toLowerCase();
		const queryIndex = lowerText.indexOf(serverQuery);

		let snippet: string;

		if (queryIndex !== -1) {
			const start = Math.max(0, queryIndex - windowSize / 2);
			const end = Math.min(text.length, start + windowSize);
			snippet = text.slice(start, end).trim();
		} else {
			// fallback to middle slice
			const middleStart = Math.max(
				0,
				Math.floor((text.length - windowSize) / 2),
			);
			snippet = text.slice(middleStart, middleStart + windowSize).trim();
		}

		return `...${snippet}...`;
	}, [result, serverQuery]);

	const collectionAccentColor = useMemo(() => {
		return calculateColorHash(result.collection.slug);
	}, [result.collection.slug]);

	const [relevanceColor, relevanceCount]: [AccentColor, number] =
		useMemo(() => {
			if (result.relevance_percent <= 20) {
				return ["red", 1];
			}

			if (result.relevance_percent <= 40) {
				return ["amber", 2];
			}

			if (result.relevance_percent <= 60) {
				return ["amber", 3];
			}

			return ["green", 5];
		}, [result.relevance_percent]);

	// Percentage of the score
	function calculateRelevance(score: number) {
		if (score === 0) return 0;
		if (score === 1) return 100;

		return (score / scores.maximum) * 100;
	}

	return (
		<Link
			to="/$workspaceSlug/c/$collectionSlug/$entryId"
			params={{
				workspaceSlug: params.workspaceSlug,
				collectionSlug: result.collection.slug,
				entryId: result.id,
			}}
			className="unstyled entry-list-item flex h-full w-full flex-grow flex-col"
			ref={measureElement}
			data-index={index}
			data-list-item={result.id}
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
								) : result.type === "link" ? (
									<AnimatedDiv>
										<Favicon
											url={(result.metadata as Metadata)?.favicon ?? ""}
										/>
									</AnimatedDiv>
								) : (
									<AnimatedDiv>
										<EntryIcon type={result.type} />
									</AnimatedDiv>
								)}
							</Flex>
						</AnimatePresence>
						<Grid m="0" p="0">
							<Text size="2" weight="medium" className="truncate">
								{result.name}
							</Text>
						</Grid>
					</Flex>

					<Tooltip content="Click to view breakdown">
						<Popover.Root open={popoverOpen} onOpenChange={setPopoverOpen}>
							<Popover.Trigger
								onClick={(e) => {
									e.stopPropagation();
									e.preventDefault();
									setPopoverOpen((prev) => !prev);
								}}
							>
								<Badge size="1" variant="soft" radius="full">
									{result.matches.length}{" "}
									{pluralize("match", result.matches.length)}
								</Badge>
							</Popover.Trigger>
							<Popover.Content
								width={{ initial: "350px", sm: "420px", xl: "500px" }}
								maxHeight={{ initial: "480px", sm: "620px", xl: "700px" }}
							>
								<Flex direction="column" gap="2" py="3" px="3">
									<Heading size="5" weight="medium" mb="2">
										Matching chunks
									</Heading>
									{result.matches.map((match, idx) => (
										<Flex
											key={match.id}
											width="100%"
											direction="column"
											align="start"
											gap="3"
										>
											<Flex gap="2" m="0" p="0" wrap="wrap">
												<Badge color="iris" variant="surface">
													Index: #{match.index}
												</Badge>
												<Badge color="crimson" variant="surface">
													Rank: #{match.rank}
												</Badge>
												{match.hybrid_score ? (
													<Badge color="mint" variant="surface">
														Score: {match.hybrid_score.toFixed(4)}
													</Badge>
												) : null}
											</Flex>
											<RelevanceBar
												percentage={calculateRelevance(match.hybrid_score)}
											/>

											<Grid>
												<Text
													as="p"
													size="1"
													color="gray"
													className="!leading-normal italic"
													trim="both"
												>
													"{match.text}"
												</Text>
											</Grid>

											<Flex
												width="100%"
												align="center"
												justify="between"
												gap="1"
											>
												<Flex direction="row" align="center" gap="1">
													{match.text_score > 0 ? (
														<Badge color="gray">Full-text match</Badge>
													) : null}
													{match.semantic_score > 0 ? (
														<Badge color="gray">Semantic match</Badge>
													) : null}
												</Flex>
											</Flex>

											{idx !== result.matches.length - 1 ? (
												<Box
													className="w-full border-b border-b-[var(--gray-5)]"
													my="3"
												/>
											) : null}
										</Flex>
									))}
								</Flex>
							</Popover.Content>
						</Popover.Root>
					</Tooltip>
				</Flex>

				<Text color="gray" size="1" className="italic">
					{preview}
				</Text>

				<Flex px="1" align="center" justify="between" gap="2">
					<Flex gap="3" align="center">
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
										workspaceSlug: result.workspace.slug,
										collectionSlug: result.collection.slug,
									},
								});
							}}
						>
							<Grid>
								<Text className="truncate" size="1">
									{result.collection.name}
								</Text>
							</Grid>
						</Badge>
						<Flex gap="1">
							<Clock size={15} color="var(--gray-11)" />
							<Text size="1" color="gray" m="0">
								{timeAgo(result.created_at)}
							</Text>
						</Flex>
					</Flex>

					<Tooltip
						content={`${result.relevance_percent?.toFixed(2)}% relevance`}
					>
						<Flex gap="1" align="end">
							{Array.from({ length: 5 }, (_, i) => (
								<Box
									// biome-ignore lint/suspicious/noArrayIndexKey: <explanation>
									key={`relevance-${i}`}
									style={{
										height: "5px",
										width: "10px",
										backgroundColor:
											i < relevanceCount
												? `var(--${relevanceColor}-9)`
												: "var(--gray-5)",
										borderRadius: "2px",
									}}
								/>
							))}
						</Flex>
					</Tooltip>
				</Flex>
			</Flex>

			{/* MARK: border */}
			<Box ml="3" mt="auto">
				<Box className="item-divider" />
			</Box>
		</Link>
	);
}

function RelevanceBar({ percentage }: { percentage: number }) {
	return (
		<Tooltip content={`Relevance: ${percentage.toFixed(2)}%`}>
			<Flex direction="column" gap="1" style={{ width: "100%" }}>
				<Progress value={percentage} />
			</Flex>
		</Tooltip>
	);
}
