import { createLazyFileRoute, useRouter } from "@tanstack/react-router";
import PageLayout from "$/components/layout/page-layout";
import PageSpinner from "$/components/page-spinner";
import ErrorComponent from "$/components/route/error-component";
import QueryKeys from "$/lib/keys";
import client from "$/lib/server";
import { useQuery } from "@tanstack/react-query";

import {
	Badge,
	Button,
	DropdownMenu,
	Flex,
	Grid,
	Text,
	IconButton,
	Card,
	Heading,
} from "@radix-ui/themes";
import {
	CaretLeft,
	Copy,
	DotsThree,
	FileArchive,
	MagnifyingGlassPlus,
	Minus,
	Plus,
} from "@phosphor-icons/react";

import type { Metadata } from "$/lib/server/types";
import { useMemo, useState } from "react";
import ReaderView from "$/components/entry/reader-view";
import useShortcut from "$/lib/hooks/use-shortcut";
import { toast } from "sonner";
import { exportEntry } from "$/lib/archive";
import type { UIScale } from "$/stores/app";
import Show from "$/components/show";
import { toTitleCase } from "$/lib/utils";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/c/$collectionSlug/$entryId",
)({
	component: RouteComponent,
});

enum LinkView {
	Reader = "reader",
	Web = "live",
}

function RouteComponent() {
	const router = useRouter();
	const params = Route.useParams();
	const navigate = Route.useNavigate();

	const [view, setView] = useState<LinkView>(LinkView.Reader);
	const [scaling, setScaling] = useState<UIScale>("100%");

	const {
		data: { entry } = {},
		isLoading,
		isError,
		error,
	} = useQuery({
		queryKey: QueryKeys.FindEntry(params.entryId),
		queryFn: async () => {
			return await client.query("entry.find", {
				entry_id: params.entryId,
				workspace_slug: params.workspaceSlug,
				collection_slug: params.collectionSlug,
			});
		},
	});

	useShortcut(["ctrl+c", "mod+c"], () => {
		copyLink();
	});

	useShortcut(["alt+r"], () => {
		if (view === LinkView.Reader) {
			setView(LinkView.Web);
		} else {
			setView(LinkView.Reader);
		}
		toast.info(
			`Switched to ${view === LinkView.Reader ? "web" : "reader"} view`,
		);
	});

	const statusColor = useMemo(() => {
		switch (entry?.status) {
			case "queued":
				return "brown";
			case "processing":
				return "amber";
			case "completed":
				return "green";
			case "failed":
				return "red";
			case "paused":
				return "bronze";
			case "canceled":
				return "gray";
			default:
				return "gray";
		}
	}, [entry?.status]);

	if (isLoading) {
		return <PageSpinner text="Loading entry..." />;
	}

	if (isError) {
		<ErrorComponent error={error} reset={null} />;
	}

	if (!entry) {
		return <ErrorComponent error={new Error("Entry not found")} reset={null} />;
	}

	function goBack() {
		if (router.history.canGoBack()) {
			return router.history.back();
		}

		navigate({
			to: "/$workspaceSlug/c/$collectionSlug",
			params: {
				workspaceSlug: params.workspaceSlug,
				collectionSlug: params.collectionSlug,
			},
		});
	}

	function zoomIn() {
		const currentScale = Number.parseInt(scaling.replace("%", ""));
		if (currentScale < 110) {
			setScaling(`${currentScale + 5}%` as UIScale);
		}
	}

	function zoomOut() {
		const currentScale = Number.parseInt(scaling.replace("%", ""));
		if (currentScale > 90) {
			setScaling(`${currentScale - 5}%` as UIScale);
		}
	}

	function copyLink() {
		if (entry?.type !== "link") {
			return;
		}

		const link = (entry?.metadata as Metadata)?.link ?? "";
		if (!link) {
			toast.error("No link found");
			return;
		}

		navigator.clipboard.writeText(link);
		toast.info("Link copied to clipboard");
	}

	if (entry.status !== "completed") {
		return (
			<PageLayout heading={entry.name} fullScreen>
				<Flex
					position="relative"
					direction="column"
					minHeight={{ initial: "90vh", md: "0" }}
					flexGrow="1"
					width="100%"
					height="100%"
					align="center"
					justify="center"
				>
					<Card variant="ghost" style={{ maxWidth: "520px" }}>
						<Flex direction="column" align="center" gap="1" px="3" py="4">
							<Badge
								size="2"
								variant="surface"
								radius="full"
								color={statusColor}
							>
								{toTitleCase(entry.status)}
							</Badge>
							<Heading
								size="4"
								weight="medium"
								color="gray"
								align="center"
								mt="1"
								highContrast
							>
								{entry.name}
							</Heading>
							<Text size="2" color="gray" align="center">
								<Show when={entry.status === "failed"}>
									The system was unable to process it.
								</Show>
								<Show when={entry.status === "canceled"}>
									It was canceled by the user. Please re-queue it to process.
								</Show>
								<Show when={entry.status === "queued"}>
									Currently waiting to be picked up by a worker. Ensure you have
									at least one plugin installed and enabled to support this
									entry type.
								</Show>
								<Show when={entry.status === "processing"}>
									It is still being processed, this may take a while.
								</Show>
								<Show when={entry.status === "paused"}>
									It is paused. Please re-queue it to process.
								</Show>
							</Text>
							<Button size="1" mt="2" onClick={goBack}>
								<CaretLeft />
								Go Back
							</Button>
						</Flex>
					</Card>
				</Flex>
			</PageLayout>
		);
	}

	return (
		<PageLayout heading={entry.name} fullScreen>
			<Flex
				position="relative"
				direction="column"
				minHeight="0"
				flexGrow="1"
				width="100%"
				height="100%"
			>
				<Flex
					width="100%"
					align="center"
					justify="between"
					px="3"
					py="2"
					gap="3"
					className="sticky top-0 right-0 left-0 z-20"
					style={{ borderBottom: "1px solid var(--gray-4)" }}
				>
					<Button size="1" variant="ghost" onClick={goBack}>
						<CaretLeft /> Back
					</Button>

					<Flex>
						<Grid mx="auto" display={{ initial: "none", xs: "grid" }}>
							<Text size="1" weight="medium" color="gray" truncate>
								{entry.name}
							</Text>
						</Grid>
					</Flex>

					<Flex gap="4" align="center">
						<Flex align="center" gap="2">
							<IconButton size="1" variant="ghost" onClick={zoomOut}>
								<Minus />
							</IconButton>
							<Badge size="1" variant="soft">
								<MagnifyingGlassPlus />
								{scaling}
							</Badge>
							<IconButton size="1" variant="ghost" onClick={zoomIn}>
								<Plus />
							</IconButton>
						</Flex>

						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<IconButton size="1" variant="ghost">
									<DotsThree size={19} color="var(--accent-11)" />
								</IconButton>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content size="2">
								{entry.type === "link" ? (
									<DropdownMenu.Item onClick={copyLink}>
										<Copy /> Copy Link
									</DropdownMenu.Item>
								) : null}
								<DropdownMenu.Item onClick={() => exportEntry(entry)}>
									<FileArchive /> Export as ZIP
								</DropdownMenu.Item>
								{entry.type === "link" ? (
									<DropdownMenu.Sub>
										<DropdownMenu.SubTrigger>View</DropdownMenu.SubTrigger>
										<DropdownMenu.SubContent>
											<DropdownMenu.RadioGroup
												value={view}
												onValueChange={(v) => setView(v as LinkView)}
											>
												<DropdownMenu.RadioItem value={LinkView.Reader}>
													Reader
												</DropdownMenu.RadioItem>
												<DropdownMenu.RadioItem value={LinkView.Web}>
													Web
												</DropdownMenu.RadioItem>
											</DropdownMenu.RadioGroup>
										</DropdownMenu.SubContent>
									</DropdownMenu.Sub>
								) : null}
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</Flex>
				</Flex>
				{entry.type !== "link" || view === LinkView.Reader ? (
					<ReaderView entry={entry} scaling={scaling} />
				) : (
					<Flex
						direction="column"
						gap="4"
						width="100%"
						height={{ initial: "90vh", md: "100%" }}
					>
						<iframe
							title={entry.name}
							src={(entry.metadata as Metadata).link}
							style={{
								width: "100%",
								height: "100%",
								border: "none",
								minHeight: 0,
								flexGrow: 1,
							}}
							loading="lazy"
							allowFullScreen
						/>
					</Flex>
				)}
			</Flex>
		</PageLayout>
	);
}
