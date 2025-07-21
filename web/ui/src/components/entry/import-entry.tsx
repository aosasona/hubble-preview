import useShortcut from "$/lib/hooks/use-shortcut";
import stores from "$/stores";
import type { NewEntry } from "$/stores/uploads";
import { CaretRight, Clipboard } from "@phosphor-icons/react";
import {
	Badge,
	Button,
	Dialog,
	DropdownMenu,
	Flex,
	IconButton,
	Inset,
	ScrollArea,
	Separator,
	Text,
	Tooltip,
} from "@radix-ui/themes";
import { toast } from "sonner";
import { useSnapshot } from "valtio";
import UploadListItem from "./upload-list-item";
import { AnimatePresence, motion } from "motion/react";
import { useEffect, useMemo, useState } from "react";
import DropIcon from "./drop-icon";
import { extractError } from "$/lib/error";
import type { MutationResult } from "$/lib/server";
import { makeApiCall } from "$/lib/server/request";
import { useQueryClient } from "@tanstack/react-query";
import QueryKeys from "$/lib/keys";

type Props = {
	open: boolean;
	onClose: () => void;
	onOpen: () => void;
};

export default function ImportEntryDialog(props: Props) {
	const queryClient = useQueryClient();

	const app = useSnapshot(stores.app);
	const uploads = useSnapshot(stores.uploads);
	const workspace = useSnapshot(stores.workspace);

	const [workspaceId, setWorkspaceId] = useState<string | null>(
		workspace.activeWorkspace?.id ?? null,
	);
	const [collectionId, setCollectionId] = useState<string | null>(
		workspace.activeCollection?.id ?? null,
	);

	useEffect(() => {
		setWorkspaceId(workspace.activeWorkspace?.id ?? null);
		setCollectionId(workspace.activeCollection?.id ?? null);
	}, [workspace.activeCollection, workspace.activeWorkspace?.id]);

	useShortcut(["mod+v", "ctrl+v"], () => pasteFromClipboard(), {
		scopes: ["import-entry"],
	});

	const hasEntries = useMemo(
		() => Object.keys(uploads.entries).length > 0,
		[uploads.entries],
	);

	const selectedCollection = useMemo(() => {
		return workspace.findWorkspaceCollectionById({
			workspaceId: workspaceId ?? "",
			collectionId: collectionId ?? "",
		});
	}, [collectionId, workspace, workspaceId]);

	const collections = useMemo(() => {
		if (!workspaceId) {
			return [];
		}

		const wk = workspace.findWorkspaceById(workspaceId);
		if (!wk) {
			return [];
		}

		return wk.collections ?? [];
	}, [workspaceId, workspace]);

	const selectedWorkspace = useMemo(() => {
		return workspace.findWorkspaceById(workspaceId ?? "");
	}, [workspaceId, workspace]);

	function onDropHandler(event: React.DragEvent) {
		event.preventDefault();
		const files = event.dataTransfer.files;
		for (let i = 0; i < files.length; i++) {
			const file = files[i];
			uploads.add({
				type: "file",
				name: file.name,
				tags: [],
				status: "success",
			});
			uploads.addFile(file);
		}
	}

	function pasteFromClipboard() {
		// Add the clipboard content to the form if it is a valid link
		navigator.clipboard.readText().then((text) => {
			if (!text) {
				toast.error("No URL found in clipboard");
				return;
			}

			// Open the dialog if it is not already open
			if (!props.open) {
				props.onOpen();
			}

			uploads.addLink(text);
		});
	}

	// Prevent the default behavior of the browser when dragging files over the window
	function onDragOverHandler(event: React.DragEvent) {
		event.preventDefault();
	}

	function onFileInputChange(event: React.ChangeEvent<HTMLInputElement>) {
		const files = event.target.files;
		if (!files) {
			return;
		}

		for (let i = 0; i < files.length; i++) {
			const file = files[i];
			uploads.add({
				type: "file",
				name: file.name,
				tags: [],
				status: "success",
			});
			uploads.addFile(file);
		}
	}

	async function upload() {
		let count = 0;
		const form = new FormData();

		if (!workspaceId) {
			toast.error("No workspace selected!");
			return;
		}

		if (!collectionId) {
			toast.error("No collection selected");
			return;
		}

		form.append("workspace_id", workspaceId ?? "");
		form.append("collection_id", collectionId ?? "");

		for (const value of Object.values(uploads.entries)) {
			switch (value.type) {
				case "link":
					form.append("links", value.link);
					count++;
					break;
				case "file": {
					const file = uploads.files.find((f) => f.name === value.name);
					if (file) {
						form.append("files", file);
						count++;
					}
					break;
				}
			}
		}

		// Clear the entries
		uploads.reset();

		// Reset the workspace and collection ids
		setWorkspaceId(workspace.activeWorkspace?.id ?? null);
		setCollectionId(workspace.activeCollection?.id ?? null);

		// Close the dialog
		props.onClose();

		// Upload the files
		toast.promise(
			() => {
				return makeApiCall<MutationResult<"entry.import">>("entry/import", {
					method: "POST",
					body: form,
					credentials: "include",
				});
			},
			{
				loading: `Uploading ${count} entr${count === 1 ? "y" : "ies"}`,
				success: (data) => {
					const count = data?.entries.length;
					queryClient.invalidateQueries({
						queryKey: QueryKeys.FindAllWorkspaceEntries(
							selectedWorkspace?.slug ?? "",
						),
					});
					queryClient.invalidateQueries({
						queryKey: QueryKeys.FindAllCollectionEntries(
							selectedWorkspace?.slug ?? "",
							selectedCollection?.slug ?? "",
						),
					});

					const c = workspace.findWorkspaceCollectionById({
						workspaceId: data.workspace_id,
						collectionId: data.collection_id,
					});

					return `Added ${count} entr${count > 1 ? "ies" : "y"} to "${c?.name ?? "current"}" collection`;
				},
				error: (error) => extractError(error)?.message,
			},
		);
	}

	return (
		<Dialog.Root
			open={props.open}
			onOpenChange={(value) => {
				// Close the dialog if the user is not in the middle of a mutation and has not entered any data
				if (!value && !hasEntries) {
					props.onClose();
				}
			}}
		>
			<Dialog.Content
				maxWidth="450px"
				maxHeight="760px"
				style={{ display: "flex", flexDirection: "column", height: "100%" }}
				onDrop={onDropHandler}
				onDragOver={onDragOverHandler}
			>
				<Flex align="center" gap="1">
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<Badge
								color="gray"
								size="2"
								variant="surface"
								style={{ cursor: "pointer" }}
							>
								<Flex align="center" gap="2">
									<Text size="1" weight="medium" color="gray">
										{selectedWorkspace?.name ?? "Select workspace"}
									</Text>
									<DropdownMenu.TriggerIcon />
								</Flex>
							</Badge>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content size="2" align="start">
							<DropdownMenu.RadioGroup value={workspaceId ?? ""}>
								{workspace.all.map((wk) => (
									<DropdownMenu.RadioItem
										key={wk.id}
										value={wk.id}
										onSelect={() => {
											setWorkspaceId(wk.id);
											setCollectionId(null);
										}}
									>
										<Flex align="center" gap="2">
											<Text>{wk.name}</Text>
										</Flex>
									</DropdownMenu.RadioItem>
								))}
							</DropdownMenu.RadioGroup>
						</DropdownMenu.Content>
					</DropdownMenu.Root>

					<CaretRight size={12} className="text-[var(--gray-9)]" />

					{collections.length > 0 ? (
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Badge
									color={app.accentColor}
									size="2"
									variant="surface"
									style={{ cursor: "pointer" }}
								>
									<Flex align="center" gap="2">
										<Text size="1" weight="medium" color="gray" highContrast>
											{selectedCollection?.name ?? "Select collection"}
										</Text>
										<DropdownMenu.TriggerIcon />
									</Flex>
								</Badge>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content size="2" align="start">
								<DropdownMenu.RadioGroup value={collectionId ?? ""}>
									{(collections ?? []).map((collection) => (
										<DropdownMenu.RadioItem
											key={collection.id}
											value={collection.id}
											onSelect={() => setCollectionId(collection.id)}
										>
											<Flex align="center" gap="2">
												<Text>{collection.name}</Text>
											</Flex>
										</DropdownMenu.RadioItem>
									))}
								</DropdownMenu.RadioGroup>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					) : (
						<Badge variant="surface" color="gray" size="2">
							<Flex align="center" gap="2">
								<Text size="1" weight="medium" color="gray">
									No collections
								</Text>
							</Flex>
						</Badge>
					)}
				</Flex>

				<Dialog.Title mt="3" mb="1" size="5">
					Create entry
				</Dialog.Title>
				<Dialog.Description size="2" color="gray" mb="4">
					Paste a link, drag and drop or{" "}
					<Button
						asChild
						variant="ghost"
						color={app.accentColor}
						className="hover:!bg-transparent hover:!text-[var(--accent-8)] cursor-pointer transition-all"
					>
						<label htmlFor="files">select files</label>
					</Button>{" "}
					on your computer to create an entry.
				</Dialog.Description>

				<Flex direction="column" flexGrow="1" minHeight="0">
					<Inset side="x">
						<Separator style={{ width: "100%" }} />
					</Inset>

					<ScrollArea
						scrollbars="vertical"
						style={{ flexGrow: 1, minHeight: 0 }}
					>
						{hasEntries ? (
							<Flex direction="column" py="3" className="transition-all">
								<AnimatePresence>
									{Object.entries(uploads.entries).map(([id, entry]) => (
										<motion.div
											key={id}
											className="w-full"
											initial={{ opacity: 0, y: 10, paddingBlock: 0 }}
											animate={{
												opacity: 1,
												y: 0,
												height: "auto",
												paddingBlock: "4px",
											}}
											exit={{
												opacity: 0,
												x: 500,
												height: 0,
												overflow: "hidden",
												paddingBlock: 0,
											}}
											transition={{ duration: 0.2 }}
										>
											<UploadListItem id={id} entry={entry as NewEntry} />
										</motion.div>
									))}
								</AnimatePresence>
							</Flex>
						) : (
							<AnimatePresence>
								<motion.div
									key={hasEntries ? "entries" : "empty"}
									initial={{ opacity: 0, height: "25px" }}
									animate={{ opacity: 1, height: "auto" }}
									exit={{ opacity: 0, height: 0 }}
									style={{ flexGrow: 1, overflow: "hidden" }}
									transition={{ duration: 0.2 }}
								>
									<Flex direction="column" align="center" gap="1" py="6">
										<DropIcon />
										<Text
											size="1"
											color="gray"
											align="center"
											mt="3"
											mx="autp"
											style={{ width: "80%" }}
										>
											<Button
												color={app.accentColor}
												size="1"
												variant="ghost"
												onClick={pasteFromClipboard}
												className="hover:!bg-transparent hover:!text-[var(--accent-8)] cursor-pointer transition-all"
											>
												Paste a link
											</Button>{" "}
											or drop a file here to begin.
										</Text>
									</Flex>
								</motion.div>
							</AnimatePresence>
						)}
					</ScrollArea>
				</Flex>

				<Flex direction="column" gap="3">
					<Inset side="x">
						<Separator style={{ width: "100%" }} />
					</Inset>

					<input
						type="file"
						id="files"
						className="hidden"
						onChange={onFileInputChange}
						multiple
					/>

					<Flex direction="row" align="center" justify="between">
						<Tooltip content="Paste from clipboard" delayDuration={100}>
							<IconButton
								variant="surface"
								color="gray"
								size="2"
								onClick={pasteFromClipboard}
							>
								<Clipboard />
							</IconButton>
						</Tooltip>
						<Flex direction="row-reverse" gap="3">
							<Button
								onClick={upload}
								disabled={!hasEntries || !collectionId || !workspaceId}
							>
								Import
							</Button>
							<Button
								color="gray"
								variant="soft"
								onClick={() => {
									uploads.reset();
									props.onClose();
								}}
							>
								Cancel
							</Button>
						</Flex>
					</Flex>
				</Flex>
			</Dialog.Content>
		</Dialog.Root>
	);
}
