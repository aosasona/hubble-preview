import PageLayout from "$/components/layout/page-layout";
import AccountListItem from "$/components/mfa/account-list-item";
import DeleteAccountDialog from "$/components/mfa/delete-account-dialog";
import NewMfaAccount from "$/components/mfa/new-mfa-account";
import RegenerateBackupCodes from "$/components/mfa/regenerate-backup-codes";
import RenameAccountDialog from "$/components/mfa/rename-account-dialog";
import Show from "$/components/show";
import { AppError } from "$/lib/error";
import { useRobinMutation, useRobinQuery } from "$/lib/hooks";
import type { QueryResult } from "$/lib/server";
import stores from "$/stores";
import { Empty, Info, PaperPlaneTilt, Plus } from "@phosphor-icons/react";
import {
	Popover,
	Box,
	Button,
	Callout,
	Card,
	Flex,
	Heading,
	Separator,
	Skeleton,
	Text,
} from "@radix-ui/themes";
import { createLazyFileRoute } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { useSnapshot } from "valtio";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/security",
)({
	component: RouteComponent,
});

type Account = QueryResult<"mfa.state">["accounts"][0];

function RouteComponent() {
	const auth = useSnapshot(stores.auth);
	const [account, setAccount] = useState<Account | null>(null);
	const [action, setAction] = useState<"rename" | "delete" | null>(null);

	const { data: mfa, isLoading: isLoadingMfa } = useRobinQuery("mfa.state");

	const pwResetMutation = useRobinMutation("auth.request-password-reset", {
		retry: false,
	});

	const { call: setPreferredAccount } = useRobinMutation(
		"mfa.set-default-account",
		{
			invalidates: ["mfa.state"],
			onSuccess: (data) => {
				const msg = data?.name
					? `"${data.name?.slice(0, 32)}" will now be used as your default account`
					: "Account set as default";
				toast.success(msg);
			},
		},
	);

	function initiatePasswordReset() {
		if (!auth.user?.email) throw new AppError("No email found");
		pwResetMutation.call({ email: auth.user?.email, scope: "change" });
	}

	function onDialogClose() {
		setAccount(null);
		setAction(null);
	}

	return (
		<PageLayout heading="Security" header={{ parent: "settings" }} showHeader>
			<Flex direction="column" mt="6" gap="6">
				{!isLoadingMfa && !mfa?.enabled ? <MfaReminder /> : null}

				{/* Change password */}
				<Box>
					<Heading size="3">Change Password</Heading>

					<Flex
						direction={{ initial: "column", md: "row" }}
						justify="between"
						align={{ initial: "start", md: "center" }}
						gap="4"
						mt="2"
					>
						<Box>
							<Text color="gray">We recommend using a strong password.</Text>
						</Box>

						<Popover.Root>
							<Popover.Trigger>
								<Button
									loading={pwResetMutation.isMutating}
									ml={{ initial: "auto", md: "0" }}
								>
									<PaperPlaneTilt /> Request email
								</Button>
							</Popover.Trigger>
							<Popover.Content maxWidth="300px">
								<Text size="2">
									We will send you an email with a link to change your password.
									Do you want to proceed?
								</Text>

								<Flex gap="3" mt="4" justify="end">
									<Popover.Close>
										<Button variant="soft" color="gray">
											Cancel
										</Button>
									</Popover.Close>
									<Button
										variant="solid"
										onClick={() => initiatePasswordReset()}
										loading={pwResetMutation.isMutating}
									>
										Continue
									</Button>
								</Flex>
							</Popover.Content>
						</Popover.Root>
					</Flex>
				</Box>

				<Separator style={{ width: "100%" }} />

				{/* MFA */}
				<Box>
					<Heading size="3">MFA Accounts</Heading>
					<Box mt="2">
						<Text color="gray">
							You can add multiple MFA accounts to secure your account.
						</Text>
					</Box>

					<Skeleton loading={isLoadingMfa}>
						<Box minHeight="50px" minWidth="100%">
							<Show when={!mfa?.accounts?.length}>
								<Card my="4">
									<Flex
										direction="column"
										align="center"
										justify="center"
										gap="2"
									>
										<Empty size="50" className="text-[var(--gray-10)]" />
										<Text color="gray">
											No accounts set up yet. Add an account to secure your
											account.
										</Text>
									</Flex>
								</Card>
							</Show>
							<Show
								when={!!(mfa?.accounts?.length && mfa?.accounts?.length > 0)}
							>
								<Card my="5" variant="ghost">
									<Flex direction="column" gap="3">
										{mfa?.accounts?.map((account, idx) => (
											<AccountListItem
												key={account.id}
												account={account}
												idx={idx}
												lastIdx={mfa.accounts.length - 1}
												onRename={(acc) => {
													setAccount(acc);
													setAction("rename");
												}}
												onDelete={(acc) => {
													setAccount(acc);
													setAction("delete");
												}}
												setAsDefault={setPreferredAccount}
											/>
										))}
									</Flex>
								</Card>
							</Show>
						</Box>
					</Skeleton>

					<Flex justify="end" align="center" gap="4" wrap="wrap">
						{mfa?.enabled ? <RegenerateBackupCodes /> : null}

						<NewMfaAccount
							Trigger={
								<Button variant="solid">
									<Plus /> Add Account
								</Button>
							}
						/>
					</Flex>

					{action === "rename" ? (
						<RenameAccountDialog account={account} onClose={onDialogClose} />
					) : null}

					{action === "delete" ? (
						<DeleteAccountDialog account={account} onClose={onDialogClose} />
					) : null}
				</Box>
			</Flex>
		</PageLayout>
	);
}

function MfaReminder() {
	return (
		<Callout.Root color="red" variant="surface">
			<Callout.Icon>
				<Info size={16} />
			</Callout.Icon>
			<Callout.Text>
				You don't have Multi-Factor Authentication (MFA) enabled. We highly
				recommend enabling MFA to secure your account.
			</Callout.Text>
		</Callout.Root>
	);
}
