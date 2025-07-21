import { Check, Dot, DotsThree, Trash } from "@phosphor-icons/react";
import {
	Badge,
	DropdownMenu,
	Flex,
	IconButton,
	Separator,
	Strong,
	Text,
	Tooltip,
} from "@radix-ui/themes";
import type { QueryResult } from "$/lib/server";
import Show from "../show";
import { useMemo } from "react";
import { redactEmail } from "$/lib/utils";

type Account = QueryResult<"mfa.state">["accounts"][0];
type Props = {
	account: Account;
	idx: number;
	lastIdx: number;
	onRename: (account: Account) => void;
	onDelete: (account: Account) => void;
	setAsDefault: (accountId: string) => void;
};
export default function AccountListItem({
	account,
	idx,
	lastIdx,
	...props
}: Props) {
	const registeredAt = useMemo(() => {
		const date = new Date(account.registered_at);
		return date.toLocaleDateString();
	}, [account.registered_at]);

	return (
		<Flex direction="column" gap="3">
			<Flex
				key={account.id}
				align="center"
				gap="3"
				justify="between"
				p="0"
				m="0"
			>
				<Flex direction="column" gap="1">
					<Flex gap="2" align="center">
						<Text size="2" weight="bold">
							{account.meta?.name ? account.meta.name : `Account ${idx + 1}`}
						</Text>

						<Show when={account.preferred}>
							<Tooltip content="This is your preferred account">
								<Badge
									color="grass"
									variant="soft"
									style={{ cursor: "pointer" }}
								>
									<Check /> Default
								</Badge>
							</Tooltip>
						</Show>
					</Flex>

					<Flex gap="1" align="center">
						{account.type === "email" ? (
							<>
								<Text size="1" color="gray">
									{/* @ts-expect-error this will always exist as long as the account is an email */}
									{redactEmail(account.meta?.email)}
								</Text>
							</>
						) : (
							<Text size="1" color="gray">
								Authenticator
							</Text>
						)}
						<Dot size={16} />
						<Text size="1" color="gray">
							Added on <Strong>{registeredAt}</Strong>
						</Text>
					</Flex>
				</Flex>

				<Flex>
					<DropdownMenu.Root>
						<DropdownMenu.Trigger>
							<IconButton variant="ghost" color="gray">
								<DotsThree size={18} />
							</IconButton>
						</DropdownMenu.Trigger>
						<DropdownMenu.Content>
							{!account.preferred ? (
								<>
									<DropdownMenu.Item
										onSelect={() => props.setAsDefault(account.id)}
										disabled={account.preferred}
									>
										Set as default
									</DropdownMenu.Item>
								</>
							) : null}

							<DropdownMenu.Item
								onSelect={() => {
									props.onRename(account);
								}}
							>
								Rename
							</DropdownMenu.Item>

							<DropdownMenu.Item
								color="red"
								onSelect={() => props.onDelete(account)}
							>
								<Trash /> Remove
							</DropdownMenu.Item>
						</DropdownMenu.Content>
					</DropdownMenu.Root>
				</Flex>
			</Flex>

			{idx !== lastIdx && (
				<Separator orientation="horizontal" style={{ width: "100%" }} />
			)}
		</Flex>
	);
}
