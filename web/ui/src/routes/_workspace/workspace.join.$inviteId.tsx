import { useRobinMutation } from "$/lib/hooks";
import QueryKeys from "$/lib/keys";
import client from "$/lib/server";
import { Flex, Card, Text, Strong, Button } from "@radix-ui/themes";
import { createFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/_workspace/workspace/join/$inviteId")({
	component: RouteComponent,
	beforeLoad: ({ params }) => ({ ...params }),
	loader: ({ params }) => {
		return client.queries.workspaceInviteFind({ invite_id: params.inviteId });
	},
});

function RouteComponent() {
	const navigate = Route.useNavigate();
	const { invite } = Route.useLoaderData();

	// NOTE: this is just to show the spinner on the appropriate button
	const [isAccepting, setIsAccepting] = useState(false);

	const mutation = useRobinMutation("workspace.invite.status.update", {
		onSuccess: (data) => {
			if (data.status === "accepted" && data.workspace) {
				toast.success("Invitation accepted");
				navigate({
					to: "/$workspaceSlug",
					params: { workspaceSlug: data.workspace.slug },
				});
				return;
			}

			toast.success("Invitation declined");
			navigate({ to: "/" });
		},
		invalidates: QueryKeys.ListWorkspaceMembers(invite.workspace.id ?? "", 1),
	});

	function acceptInvite() {
		setIsAccepting(true);
		mutation.call({
			workspace_id: invite.workspace.id,
			invite_id: invite.id,
			status: "accepted",
		});
	}

	function declineInvite() {
		setIsAccepting(false);
		mutation.call({
			workspace_id: invite.workspace.id,
			invite_id: invite.id,
			status: "declined",
		});
	}

	return (
		<Flex
			direction="column"
			width="100vw"
			height="100vh"
			align="center"
			justify="center"
		>
			<Card variant="ghost">
				<Flex
					gap="3"
					direction="column"
					align="center"
					width={{ initial: "95vw", md: "450px" }}
					p="3"
				>
					<Text align="center" size="2">
						<Strong>{invite.inviter.first_name}</Strong> invited you to join
						their workspace (<Strong>{invite.workspace.name}</Strong>).
					</Text>

					<Text size="2" color="gray" align="center">
						By accepting this invitation, you will be able to access the
						workspace and collaborate with other members.
					</Text>

					<Flex gap="2" direction="row-reverse">
						<Button
							variant="solid"
							size="1"
							onClick={acceptInvite}
							loading={mutation.isMutating && isAccepting}
							disabled={mutation.isMutating}
						>
							Accept
						</Button>
						<Button
							variant="soft"
							color="gray"
							size="1"
							onClick={declineInvite}
							loading={mutation.isMutating && !isAccepting}
							disabled={mutation.isMutating}
						>
							Decline
						</Button>
					</Flex>
				</Flex>
			</Card>
		</Flex>
	);
}
