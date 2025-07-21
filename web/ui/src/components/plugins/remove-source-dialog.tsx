import { AlertDialog, Flex, Button } from "@radix-ui/themes";
import type { PluginSource, Workspace } from "$/lib/server/types";
import { useRobinMutation } from "$/lib/hooks";
import { toast } from "sonner";

type Props = {
	workspace: Workspace;
	source: PluginSource | null;
	onClose: () => void;
};

export default function RemoveSourceDialog(props: Props) {
	const mutation = useRobinMutation("plugin.source.remove", {
		onSuccess: () => {
			toast.success(`Removed source "${props.source?.name}"`);
			props.onClose();
		},
		retry: false,
		invalidates: ["plugin.source.list"],
	});

	return (
		<AlertDialog.Root open={!!props.source} onOpenChange={props.onClose}>
			<AlertDialog.Content maxWidth="450px">
				<AlertDialog.Title>Remove source</AlertDialog.Title>
				<AlertDialog.Description size="2" color="gray">
					"{props.source?.name}" and all plugins installed from it will be
					removed from your current workspace ({props.workspace.name}), are you
					sure you want to continue?
				</AlertDialog.Description>

				<Flex justify="end" mt="4" gap="3">
					<AlertDialog.Cancel>
						<Button variant="soft" color="gray">
							Cancel
						</Button>
					</AlertDialog.Cancel>
					<Button
						variant="solid"
						color="red"
						loading={mutation.isMutating}
						onClick={() => {
							if (props.source) {
								mutation.call({
									workspace_id: props.workspace.id,
									source_id: props.source.id,
								});
							}
						}}
					>
						Remove
					</Button>
				</Flex>
			</AlertDialog.Content>
		</AlertDialog.Root>
	);
}
