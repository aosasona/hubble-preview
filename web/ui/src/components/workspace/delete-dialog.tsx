import { Button, Dialog, Flex } from "@radix-ui/themes";
import type { Workspace, Member } from "$/lib/server/types";
import { toast } from "sonner";
import { useRobinMutation } from "$/lib/hooks";
import QueryKeys from "$/lib/keys";

type Props = {
	page: number;
	workspace: Workspace;
	member: Member | null;
	onClose: () => void;
};

export default function DeleteDialog(props: Props) {
	const mutation = useRobinMutation("workspace.member.remove", {
		onSuccess: (data) => {
			toast.success(
				`"${data.first_name} ${data.last_name}" removed successfully`,
			);
			props.onClose();
		},
		invalidates: QueryKeys.ListWorkspaceMembers(props.workspace.id, props.page),
		retry: false,
	});

	function removeMember(email: string) {
		toast.promise(
			() => {
				return mutation.call({
					email,
					workspace_id: props.workspace.id,
				});
			},
			{
				loading: "Removing member...",
			},
		);
	}

	return (
		<Dialog.Root open={!!props.member} onOpenChange={props.onClose}>
			<Dialog.Content maxWidth="450px">
				<Dialog.Title>Remove Member</Dialog.Title>
				<Dialog.Description size="2" color="gray" mb="4">
					Are you sure you want to remove "{props.member?.first_name}{" "}
					{props.member?.last_name}" from this workspace?
				</Dialog.Description>

				<Flex
					direction="row-reverse"
					gap="3"
					align="center"
					justify="start"
					mt="4"
				>
					<Button
						type="submit"
						variant="solid"
						size="2"
						loading={mutation.isMutating}
						onClick={() => {
							if (props.member) {
								removeMember(props.member.email);
							}
						}}
					>
						Remove
					</Button>
					<Dialog.Close>
						<Button type="button" variant="soft" color="gray" size="2">
							Cancel
						</Button>
					</Dialog.Close>
				</Flex>
			</Dialog.Content>
		</Dialog.Root>
	);
}
