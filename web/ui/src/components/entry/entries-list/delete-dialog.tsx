import { useRobinMutation } from "$/lib/hooks";
import QueryKeys from "$/lib/keys";
import stores from "$/stores";
import { AlertDialog, Button, Flex } from "@radix-ui/themes";
import { useMemo } from "react";
import { toast } from "sonner";
import { useSnapshot } from "valtio";

export default function DeleteEntriesDialog() {
	const app = useSnapshot(stores.app);
	const selections = useSnapshot(stores.entriesList.selections);
	const workspace = useSnapshot(stores.workspace);

	const q = useMemo(() => {
		const searchParams = new URLSearchParams(window.location.search);
		const q = searchParams.get("q");
		if (!q) {
			return "";
		}
		return q;
	}, []);

	const mutation = useRobinMutation("entry.delete", {
		invalidates: [
			QueryKeys.FindAllCollectionEntries(
				workspace.activeWorkspace?.slug ?? "",
				workspace.activeCollection?.slug ?? "",
			),
			QueryKeys.FindAllWorkspaceEntries(workspace.activeWorkspace?.slug ?? ""),
			["entry", "search", q],
		],
		onSuccess: (data) => {
			stores.entriesList.selections.clear();
			toast.success(data.message);
		},
	});

	return (
		<AlertDialog.Root
			open={app.dialogs.deleteSelectedEntries}
			onOpenChange={(v) =>
				stores.app.setDialogState("deleteSelectedEntries", v)
			}
		>
			<AlertDialog.Content maxWidth="400px">
				<AlertDialog.Title>Delete entries</AlertDialog.Title>
				<AlertDialog.Description size="2">
					Are you sure you want to delete {selections.size} entr
					{selections.size === 1 ? "y" : "ies"}?
				</AlertDialog.Description>
				<Flex gap="3" mt="4" justify="end">
					<AlertDialog.Cancel>
						<Button variant="soft" color="gray">
							Cancel
						</Button>
					</AlertDialog.Cancel>
					<AlertDialog.Action>
						<Button
							type="button"
							variant="solid"
							color="red"
							loading={mutation.isMutating}
							onClick={() =>
								mutation.call({
									workspace_slug: workspace.activeWorkspace?.slug ?? "",
									entry_ids: Array.from(selections),
								})
							}
						>
							Delete
						</Button>
					</AlertDialog.Action>
				</Flex>
			</AlertDialog.Content>
		</AlertDialog.Root>
	);
}
