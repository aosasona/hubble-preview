import PageLayout from "$/components/layout/page-layout";
import PageSpinner from "$/components/page-spinner";
import ErrorComponent from "$/components/route/error-component";
import Show from "$/components/show";
import DeleteDialog from "$/components/workspace/delete-dialog";
import InviteDialog from "$/components/workspace/invite-dialog";
import { InviteStatus } from "$/components/workspace/invite-status";
import MemberActions from "$/components/workspace/member-actions";
import { useRobinQuery } from "$/lib/hooks";
import QueryKeys from "$/lib/keys";
import type { QueryResult } from "$/lib/server";
import type { Member } from "$/lib/server/types";
import { MEMBER_ROLE } from "$/lib/types";
import type { OneOf } from "$/lib/types";
import stores from "$/stores";
import {
	ArrowDown,
	ArrowLeft,
	ArrowRight,
	ArrowUp,
	Clock,
	Hourglass,
	MagnifyingGlass,
	X,
} from "@phosphor-icons/react";
import {
	Button,
	DropdownMenu,
	Flex,
	Table,
	Text,
	TextField,
} from "@radix-ui/themes";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import { useMemo, useState } from "react";
import { useSnapshot } from "valtio";

enum View {
	All = "All",
	Members = "Members",
	Invites = "Invites",
}

type Key = keyof OneOf<QueryResult<"workspace.members.all">["members"]>;
type Order = "asc" | "desc";
type SortBy = {
	key: Key;
	order: Order;
};

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/members",
)({
	component: RouteComponent,
});

function RouteComponent() {
	const { status, workspace } = Route.useLoaderData();
	const auth = useSnapshot(stores.auth);

	const params = Route.useParams();
	const search = Route.useSearch();

	const [view, setView] = useState<View>(View.All);
	const [sorting, setSortBy] = useState<SortBy>({
		key: "first_name",
		order: "asc",
	});
	const [searchQuery, setSearchQuery] = useState("");
	const [openInviteDialog, setOpenInviteDialog] = useState(false);
	const [deleteMember, setDeleteMember] = useState<Member | null>(null);

	const query = useRobinQuery(
		"workspace.members.all",
		{
			workspace_id: workspace.id ?? "",
			pagination: {
				page: search.page,
				per_page: 50,
			},
		},
		{
			queryKey: QueryKeys.ListWorkspaceMembers(workspace.id, search.page),
			retry: 1,
		},
	);

	const members = useMemo(() => {
		let data = query.data?.members ?? [];

		if (searchQuery) {
			data = data.filter((member) => {
				return (
					member.email?.toLocaleLowerCase().includes(searchQuery) ||
					member.first_name
						?.toLocaleLowerCase()
						.includes(searchQuery.toLocaleLowerCase()) ||
					member.last_name
						?.toLocaleLowerCase()
						.includes(searchQuery.toLocaleLowerCase())
				);
			});
		}

		// Filter data based on view
		if (view === View.Members) {
			data = data.filter((member) => member.status === "accepted");
		} else if (view === View.Invites) {
			data = data.filter(
				(member) =>
					member.status === "pending" ||
					member.status === "revoked" ||
					member.status === "declined",
			);
		}

		// Sort data
		data = data.sort((a, b) => {
			const key = sorting.key as keyof typeof a;
			const aValue = a?.[key] ?? "";
			const bValue = b?.[key] ?? "";

			if (aValue < bValue) {
				return sorting.order === "asc" ? -1 : 1;
			}
			if (aValue > bValue) {
				return sorting.order === "asc" ? 1 : -1;
			}
			return 0;
		});

		return data;
	}, [query.data?.members, searchQuery, sorting, view]);

	function toggleSortBy(column: Key) {
		if (sorting.key === column) {
			setSortBy({
				key: column,
				order: sorting.order === "asc" ? "desc" : "asc",
			});
			return;
		}

		setSortBy({ key: column, order: "asc" });
	}

	if (query.isLoading) {
		return <PageSpinner text="Loading members..." />;
	}

	if (query.isError && query.error) {
		return <ErrorComponent error={query.error} reset={() => {}} />;
	}

	return (
		<PageLayout heading="Members" header={{ parent: "settings" }} showHeader>
			<Flex width="100%" direction="column" mt="4" gap="4" align="start">
				<Flex align="center" gap="2" justify="between" width="100%" wrap="wrap">
					<Flex align="center" gap="2">
						<TextField.Root
							name="search"
							variant="surface"
							color="gray"
							placeholder="Search by name or email"
							size="2"
							value={searchQuery}
							onChange={(e) => setSearchQuery(e.target.value)}
							style={{ width: "min(300px, 75vw)" }}
						>
							<TextField.Slot side="left">
								<MagnifyingGlass />
							</TextField.Slot>
						</TextField.Root>

						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								<Button variant="surface" color="gray" size="2">
									{view}
									<DropdownMenu.TriggerIcon />
								</Button>
							</DropdownMenu.Trigger>
							<DropdownMenu.Content size="2" align="start">
								<DropdownMenu.Item onSelect={() => setView(View.All)}>
									All
								</DropdownMenu.Item>
								<DropdownMenu.Item onSelect={() => setView(View.Members)}>
									Members
								</DropdownMenu.Item>
								<DropdownMenu.Item onSelect={() => setView(View.Invites)}>
									Invites
								</DropdownMenu.Item>
							</DropdownMenu.Content>
						</DropdownMenu.Root>
					</Flex>

					{status.role === "admin" || status.role === "owner" ? (
						<InviteDialog
							open={openInviteDialog}
							onOpenChange={setOpenInviteDialog}
						/>
					) : null}
				</Flex>

				<Table.Root size="1" variant="ghost" className="w-full">
					<Table.Header>
						<Table.Row>
							<Table.ColumnHeaderCell>
								<HeaderCell
									title="Name"
									column="first_name"
									sorting={sorting}
									toggle={toggleSortBy}
								/>
							</Table.ColumnHeaderCell>
							<Table.ColumnHeaderCell>
								<HeaderCell
									title="Email"
									column="email"
									sorting={sorting}
									toggle={toggleSortBy}
								/>
							</Table.ColumnHeaderCell>
							<Table.ColumnHeaderCell>
								<HeaderCell
									title="Role"
									column="role"
									sorting={sorting}
									toggle={toggleSortBy}
								/>
							</Table.ColumnHeaderCell>
							<Table.ColumnHeaderCell>{null}</Table.ColumnHeaderCell>
						</Table.Row>
					</Table.Header>

					<Table.Body>
						{members.map((member) => (
							<Table.Row key={member.id}>
								<Table.RowHeaderCell>
									<Flex gap="2" align="center">
										<Show when={member.status === "pending"}>
											<InviteStatus
												icon={Hourglass}
												text="Pending"
												color="amber"
											/>
										</Show>
										<Show when={member.status === "declined"}>
											<InviteStatus icon={X} text="Declined" color="red" />
										</Show>
										<Show when={member.status === "expired"}>
											<InviteStatus icon={Clock} text="Expired" color="gray" />
										</Show>
										{member.first_name ? (
											<Text>
												{member.first_name} {member.last_name}{" "}
												{auth.user?.user_id === member.user.id ? "(You)" : null}
											</Text>
										) : (
											<Text color="gray" className="italic">
												Unknown
											</Text>
										)}
									</Flex>
								</Table.RowHeaderCell>
								<Table.Cell>
									<Text color="gray">{member.email}</Text>
								</Table.Cell>
								<Table.Cell>
									<Text color="gray" weight="medium" highContrast>
										{MEMBER_ROLE[member.role]}
									</Text>
								</Table.Cell>
								<Table.Cell>
									{["admin", "owner"].includes(status.role) &&
									member.role !== "owner" &&
									member.user.id !== auth.user?.user_id ? (
										<MemberActions
											workspace={workspace}
											member={member}
											status={status}
											page={search.page}
											onDelete={setDeleteMember}
										/>
									) : null}
								</Table.Cell>
							</Table.Row>
						))}
					</Table.Body>
				</Table.Root>

				<Flex align="center" justify="end" width="100%" gap="3" mt="4">
					<Button
						size="1"
						color="gray"
						variant="soft"
						disabled={!query.data?.pagination.previous_page}
						asChild
					>
						<Link
							to="/$workspaceSlug/settings/workspace"
							params={{ workspaceSlug: params.workspaceSlug }}
							search={{ page: query.data?.pagination.previous_page ?? 0 }}
						>
							<ArrowLeft size={12} />
							Previous
						</Link>
					</Button>

					<Button
						variant="solid"
						size="1"
						disabled={!query.data?.pagination.next_page}
						asChild
					>
						<Link
							to="/$workspaceSlug/settings/workspace"
							params={{ workspaceSlug: params.workspaceSlug }}
							search={{ page: query.data?.pagination.next_page ?? 1 }}
							className="text-[var(--accent-foreground)] hover:text-inherit"
						>
							Next
							<ArrowRight size={12} />
						</Link>
					</Button>
				</Flex>
			</Flex>

			<DeleteDialog
				page={search.page}
				member={deleteMember}
				workspace={workspace}
				onClose={() => setDeleteMember(null)}
			/>
		</PageLayout>
	);
}

function HeaderCell(props: {
	title: string;
	column: Key;
	sorting: SortBy;
	toggle: (column: Key) => void;
}) {
	return (
		<Button
			size="1"
			variant="ghost"
			color="gray"
			onClick={() => props.toggle(props.column)}
		>
			{props.title}
			{props.column === props.sorting.key ? (
				props.sorting.order === "asc" ? (
					<ArrowUp size={13} />
				) : (
					<ArrowDown size={13} />
				)
			) : null}
		</Button>
	);
}
