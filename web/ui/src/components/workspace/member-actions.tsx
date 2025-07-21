import { useRobinMutation } from "$/lib/hooks";
import QueryKeys from "$/lib/keys";
import type { MembershipStatus, Workspace, Member } from "$/lib/server/types";
import { MEMBER_ROLE, type MemberRole } from "$/lib/types";
import stores from "$/stores";
import {
	ArrowClockwise,
	Check,
	Copy,
	DotsThree,
	Trash,
	X,
} from "@phosphor-icons/react";
import { DropdownMenu, IconButton } from "@radix-ui/themes";
import { toast } from "sonner";
import { useSnapshot } from "valtio";
import Show from "../show";

type Props = {
	member: Member;
	workspace: Workspace;
	status: MembershipStatus;
	page: number;
	onDelete: (member: Member) => void;
};

export default function MemberActions({
	workspace,
	member,
	page,
	onDelete,
}: Props) {
	const auth = useSnapshot(stores.auth);

	const sendInviteMutation = useRobinMutation("workspace.invite", {
		onSuccess: (data) => {
			toast.success(data.message);
		},
		invalidates: QueryKeys.ListWorkspaceMembers(workspace.id, page),
		retry: false,
	});

	const revokeInviteMutation = useRobinMutation(
		"workspace.invite.status.update",
		{
			onSuccess: (data) => {
				if (data.status === "revoked") {
					toast.success("Invite revoked successfully");
				}
			},
			invalidates: QueryKeys.ListWorkspaceMembers(workspace.id, page),
			retry: false,
		},
	);

	const changeRoleMutation = useRobinMutation("workspace.member.role.update", {
		onSuccess: () => {
			toast.success("Role updated successfully");
		},
		invalidates: QueryKeys.ListWorkspaceMembers(workspace.id, page),
		retry: false,
	});

	function resendInvite(email: string) {
		toast.promise(
			() => {
				return sendInviteMutation.call({
					emails: [email],
					workspace_id: workspace.id,
				});
			},
			{ loading: "Resending invite..." },
		);
	}

	function revokeInvite(invite_id: string) {
		toast.promise(
			() => {
				return revokeInviteMutation.call({
					invite_id,
					workspace_id: workspace.id ?? "",
					status: "revoked",
				});
			},
			{ loading: "Revoking invite..." },
		);
	}

	function changeRole(role: MemberRole) {
		toast.promise(
			() => {
				return changeRoleMutation.call({
					workspace_id: workspace.id,
					user_id: member.user.id,
					role: role,
				});
			},
			{ loading: "Updating role..." },
		);
	}

	function copyInviteLink() {
		if (!member.invite_id) return;

		const scheme = window.location.protocol;
		const host = window.location.host;
		const url = `${scheme}//${host}/workspace/join/${member.invite_id}`;

		navigator.clipboard.writeText(url);
		toast.success("Invite link copied to clipboard");
	}

	return (
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				<IconButton variant="ghost" color="gray" size="1">
					<DotsThree size={16} />
				</IconButton>
			</DropdownMenu.Trigger>
			<DropdownMenu.Content size="2" align="start">
				<Show when={member.status !== "accepted"}>
					<DropdownMenu.Item onSelect={() => copyInviteLink()}>
						<Copy />
						Copy invite link
					</DropdownMenu.Item>
					<DropdownMenu.Item
						onSelect={() => {
							resendInvite(member.email);
						}}
					>
						<ArrowClockwise />
						Resend invite
					</DropdownMenu.Item>
				</Show>
				{!member.invite_id &&
				member.user.id !== auth.user?.user_id &&
				member.role !== "owner" ? (
					<DropdownMenu.Sub>
						<DropdownMenu.SubTrigger>Change role</DropdownMenu.SubTrigger>
						<DropdownMenu.SubContent>
							{Object.entries(MEMBER_ROLE).map(([role, name]) => {
								if (role === "owner") return null;

								return (
									<DropdownMenu.Item
										key={role}
										onSelect={() => {
											changeRole(role as MemberRole);
										}}
									>
										{member.role === role ? <Check /> : null}
										{name}
									</DropdownMenu.Item>
								);
							})}
						</DropdownMenu.SubContent>
					</DropdownMenu.Sub>
				) : null}
				{member.status === "accepted" || !member.invite_id ? (
					<DropdownMenu.Item color="red" onSelect={() => onDelete(member)}>
						<Trash /> Remove
					</DropdownMenu.Item>
				) : (
					<DropdownMenu.Item
						color="red"
						onSelect={() => revokeInvite(member.invite_id ?? "")}
						disabled={revokeInviteMutation.isMutating}
					>
						<X /> Revoke
					</DropdownMenu.Item>
				)}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	);
}
