import { createFileRoute } from "@tanstack/react-router";
import {
	AlertDialog,
	Button,
	Callout,
	Checkbox,
	Dialog,
	DropdownMenu,
	Flex,
	IconButton,
	Spinner,
	Table,
	Text,
	TextField,
} from "@radix-ui/themes";
import { useRobinMutation, useRobinQuery } from "$/lib/hooks";
import * as v from "valibot";
import QueryKeys from "$/lib/keys";
import client from "$/lib/server";
import ErrorComponent from "$/components/route/error-component";
import { extractError } from "$/lib/error";
import { MEMBER_ROLE, type OneOf } from "$/lib/types";
import type { QueryResult } from "$/lib/server";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import { useMemo, useState } from "react";
import { Link } from "@tanstack/react-router";
import Show from "$/components/show";
import {
	WarningCircle,
	MagnifyingGlass,
	ArrowLeft,
	ArrowRight,
	ArrowUp,
	ArrowDown,
	DotsThree,
} from "@phosphor-icons/react";
import PageLayout from "$/components/layout/page-layout";
import * as Form from "@radix-ui/react-form";
import CustomTextArea from "$/components/form/custom-text-area";
import { useForm } from "react-hook-form";
import { valibotResolver } from "@hookform/resolvers/valibot";
import type { Collection, Workspace } from "$/lib/server/types";
import { toast } from "sonner";

const DEFAULT_PAGE_SIZE = 50;
const paramsSchema = v.object({
	page: v.optional(v.number(), 1),
});

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/c/$collectionSlug/members",
)({
	validateSearch: paramsSchema,
	component: RouteComponent,
	beforeLoad: ({ params }) => params,
	loader: ({ params }) => {
		return client.queries.collectionMemberStatus({
			workspace_slug: params.workspaceSlug,
			collection_slug: params.collectionSlug,
		});
	},
	errorComponent: ErrorComponent,
});

type Key = keyof OneOf<QueryResult<"collection.members.all">["members"]>;
type Order = "asc" | "desc";
type SortBy = {
	key: Key;
	order: Order;
};

function RouteComponent() {
	const auth = useSnapshot(stores.auth);
	const search = Route.useSearch();
	const {
		workspace,
		collection,
		membership_status: membership,
	} = Route.useLoaderData();
	const params = Route.useParams();

	const [showAddMemberDialog, setShowAddMemberDialog] = useState(false);
	const [showRemoveMemberDialog, setShowRemoveMemberDialog] = useState(false);

	const [selectedMembers, setSelectedMembers] = useState<string[]>([]);
	const [searchQuery, setSearchQuery] = useState("");
	const [sorting, setSortBy] = useState<SortBy>({
		key: "first_name",
		order: "asc",
	});

	const query = useRobinQuery(
		"collection.members.all",
		{
			collection_id: collection.id,
			workspace_id: workspace.id,
			pagination: {
				page: search.page,
				per_page: DEFAULT_PAGE_SIZE,
			},
		},
		{
			queryKey: QueryKeys.ListCollectionMembers(
				workspace.id,
				collection.id,
				search.page,
			),
		},
	);

	const removeMemberMutation = useRobinMutation("collection.members.remove", {
		onSuccess: (data) => {
			toast.success(
				`Removed ${data.removed_count} member${data.removed_count > 1 ? "s" : ""}`,
			);
		},
		invalidates: QueryKeys.ListCollectionMembers(
			workspace.id,
			collection.id,
			search.page,
		),
	});

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
	}, [query.data?.members, searchQuery, sorting]);

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

	function toggleSelection(email: string) {
		if (selectedMembers.includes(email)) {
			setSelectedMembers((prev) => prev.filter((id) => id !== email));
		} else {
			setSelectedMembers((prev) => [...prev, email]);
		}
	}

	function removeSelectedMembers() {
		toast.promise(
			async () => {
				return await removeMemberMutation.call({
					workspace_id: workspace.id,
					collection_id: collection.id,
					emails: selectedMembers,
				});
			},
			{
				loading: `Removing ${selectedMembers.length} member${selectedMembers.length > 1 ? "s" : ""}...`,
				finally: () => {
					setSelectedMembers([]);
				},
			},
		);

		setShowRemoveMemberDialog(false);
	}

	return (
		<PageLayout
			heading={collection.name}
			header={{
				parent: "settings",
				items: [
					{ title: "Collections" },
					{
						title: collection.name,
						url: `/${workspace.slug}/settings/workspace/c/${collection.slug}`,
					},
					{ title: "Members" },
				],
			}}
			showHeader
		>
			<Flex direction="column" gap="1" width="100%" mt="2">
				<Link
					to="/$workspaceSlug/settings/workspace/c/$collectionSlug"
					params={{
						workspaceSlug: params.workspaceSlug,
						collectionSlug: params.collectionSlug,
					}}
					className="!text-[var(--accent-9)] !my-[var(--space-1)] !no-underline hover:!text-[var(--accent-contrast)] flex w-max items-center gap-1 rounded-[var(--radius-2)] border border-transparent bg-[var(--accent-surface)] px-2 py-[var(--space-1)] transition-all hover:bg-[var(--accent-9)]"
				>
					<ArrowLeft />
					<Text size="1">Go back</Text>
				</Link>

				<Text size="2" color="gray">
					View and manage members in this collection.
				</Text>

				<Show when={query.isLoading}>
					<Flex width="100%" height="300px" align="center" justify="center">
						<Spinner />
					</Flex>
				</Show>

				<Show when={query.isError && query.error != null}>
					<Callout.Root color="red" variant="surface" mt="4">
						<Callout.Icon>
							<WarningCircle />
						</Callout.Icon>
						<Callout.Text>{extractError(query.error)?.message}</Callout.Text>
					</Callout.Root>
				</Show>

				<Show when={query.isSuccess && query.data != null}>
					<Flex direction="column" gap="1">
						<Flex justify="between" align="center" my="4" gap="2" wrap="wrap">
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

							<Flex align="center" gap="2">
								{membership.role === "owner" || membership.role === "admin" ? (
									<AddMemberDialog
										open={showAddMemberDialog}
										onOpenChange={setShowAddMemberDialog}
										workspace={workspace}
										collection={collection}
										page={search.page}
									/>
								) : null}
								{selectedMembers.length > 0 ? (
									<DropdownMenu.Root>
										<DropdownMenu.Trigger>
											<IconButton color="gray" variant="soft">
												<DotsThree size={16} />
											</IconButton>
										</DropdownMenu.Trigger>
										<DropdownMenu.Content size="2" align="start">
											<DropdownMenu.Item
												color="red"
												onSelect={() => {
													setShowRemoveMemberDialog(true);
												}}
											>
												<WarningCircle />
												Remove {selectedMembers.length} member
												{selectedMembers.length > 1 ? "s" : ""}
											</DropdownMenu.Item>
										</DropdownMenu.Content>
									</DropdownMenu.Root>
								) : null}
							</Flex>
						</Flex>

						<Table.Root size="1" variant="ghost" className="w-full">
							<Table.Header>
								<Table.Row>
									<Table.ColumnHeaderCell />
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
											<Checkbox
												checked={selectedMembers.includes(member.email)}
												onCheckedChange={() => toggleSelection(member.email)}
												disabled={
													member.role === "owner" ||
													membership.role === "user" ||
													membership.role === "guest"
												}
											/>
										</Table.RowHeaderCell>
										<Table.Cell>
											<Flex gap="2" align="center">
												<Text>
													{member.first_name} {member.last_name}{" "}
													{auth?.user?.user_id === member.user.id
														? "(You)"
														: null}
												</Text>
											</Flex>
										</Table.Cell>
										<Table.Cell>
											<Text color="gray">{member.email}</Text>
										</Table.Cell>
										<Table.Cell>
											<Text color="gray" weight="medium" highContrast>
												{MEMBER_ROLE[member.role]}
											</Text>
										</Table.Cell>
										<Table.Cell>{/* TODO: add actions */}</Table.Cell>
									</Table.Row>
								))}
							</Table.Body>
						</Table.Root>

						<Flex align="center" justify="between" width="100%" mt="4">
							<Button
								variant="ghost"
								size="2"
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
								variant="ghost"
								size="2"
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
				</Show>
			</Flex>

			{/* MARK: remove member dialog */}
			<AlertDialog.Root
				open={showRemoveMemberDialog}
				onOpenChange={setShowRemoveMemberDialog}
			>
				<AlertDialog.Content maxWidth="450px">
					<AlertDialog.Title mb="1">Remove member(s)</AlertDialog.Title>
					<AlertDialog.Description color="gray" size="2" mb="4">
						Are you sure you want to remove {selectedMembers.length} member
						{selectedMembers.length > 1 ? "s" : ""}?
					</AlertDialog.Description>

					<Flex direction="row" gap="3" align="center" justify="end">
						<AlertDialog.Cancel>
							<Button variant="soft" color="gray" size="2">
								Cancel
							</Button>
						</AlertDialog.Cancel>
						<Button
							variant="solid"
							color="red"
							size="2"
							onClick={removeSelectedMembers}
							loading={removeMemberMutation.isMutating}
						>
							Remove
						</Button>
					</Flex>
				</AlertDialog.Content>
			</AlertDialog.Root>
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

const addMembersSchema = v.object({
	workspace_id: v.pipe(v.string(), v.nonEmpty(), v.uuid()),
	collection_id: v.pipe(v.string(), v.nonEmpty(), v.uuid()),
	emails: v.pipe(
		v.string(),
		v.trim(),
		v.nonEmpty("Emails cannot be empty"),
		v.transform((value) => {
			return value
				.split(",")
				.map((email) => email.trim())
				.filter((email) => email.length > 0);
		}),
		v.array(
			v.pipe(
				v.string(),
				v.nonEmpty("Email cannot be empty"),
				v.email((v) => `${v.input} is not a valid email address`),
			),
		),
	),
});

type AddMembersSchema = v.InferOutput<typeof addMembersSchema>;

type AddMemberDialogProps = {
	open: boolean;
	onOpenChange: (v: boolean) => void;
	workspace: Workspace;
	collection: Collection;
	page: number;
};
function AddMemberDialog(props: AddMemberDialogProps) {
	const form = useForm<AddMembersSchema>({
		resolver: valibotResolver(addMembersSchema),
		defaultValues: { emails: [] },
	});

	const mutation = useRobinMutation("collection.members.add", {
		onSuccess: (data) => {
			toast.success(
				`Added ${data.added_count} member${data.added_count > 1 ? "s" : ""}`,
			);
			props.onOpenChange(false);
		},
		setFormError: form.setError,
		invalidates: QueryKeys.ListCollectionMembers(
			props.workspace.id,
			props.collection.id,
			props.page,
		),
	});

	return (
		<Dialog.Root
			open={props.open}
			onOpenChange={(v) => {
				props.onOpenChange(v);
				form.reset();
			}}
		>
			<Dialog.Trigger>
				<Button size="2">Add Member(s)</Button>
			</Dialog.Trigger>
			<Dialog.Content maxWidth="450px">
				<Dialog.Title mb="2">
					Add members to "{props.collection.name}"
				</Dialog.Title>
				<Dialog.Description size="2" color="gray" mb="4">
					You can add multiple members at once by separating their email
					addresses with commas.
				</Dialog.Description>

				<Form.Root onSubmit={form.handleSubmit((data) => mutation.call(data))}>
					<input
						{...form.register("workspace_id")}
						type="hidden"
						value={props.workspace.id}
					/>
					<input
						{...form.register("collection_id")}
						type="hidden"
						value={props.collection.id}
					/>
					<CustomTextArea
						register={form.register}
						name="emails"
						label="Emails"
						placeholder="Enter email addresses separated by commas"
						errors={form.formState.errors}
						hideLabel
					/>

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
						>
							Add Members
						</Button>
						<Dialog.Close>
							<Button type="button" variant="soft" color="gray" size="2">
								Cancel
							</Button>
						</Dialog.Close>
					</Flex>
				</Form.Root>
			</Dialog.Content>
		</Dialog.Root>
	);
}
