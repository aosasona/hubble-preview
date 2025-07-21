import { createFileRoute, Link } from "@tanstack/react-router";
import {
	AlertDialog,
	Badge,
	Button,
	Flex,
	Heading,
	Separator,
	Text,
} from "@radix-ui/themes";
import { useForm } from "react-hook-form";
import { useRobinMutation } from "$/lib/hooks";
import * as Form from "@radix-ui/react-form";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";
import PageLayout from "$/components/layout/page-layout";
import RowInput from "$/components/form/row-input";
import QueryKeys from "$/lib/keys";
import RowTextArea from "$/components/form/row-text-area";
import { toast } from "sonner";
import { MixedNameRegex } from "$/lib/utils";
import client from "$/lib/server";
import ErrorComponent from "$/components/route/error-component";
import { useEffect } from "react";
import { ArrowSquareOut, SignOut, Warning } from "@phosphor-icons/react";
import stores from "$/stores";
import Show from "$/components/show";

const paramsSchema = v.object({
	page: v.optional(v.number(), 1),
});

export const Route = createFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/c/$collectionSlug/",
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

const formSchema = v.object({
	name: v.pipe(
		v.string(),
		v.trim(),
		v.nonEmpty("Name is required"),
		v.minLength(1, "Name must be at least 1 character"),
		v.maxLength(64, "Name must be at most 64 characters"),
		v.regex(
			MixedNameRegex,
			"Name must contain only letters, numbers, hyphens, underscores and spaces",
		),
	),
	slug: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Slug must be at least 2 characters"),
		v.maxLength(80, "Slug must be at most 80 characters"),
		v.regex(
			/^[a-zA-Z0-9-]*$/,
			"Slug must contain only letters, numbers, and hyphens",
		),
	),
	description: v.pipe(
		v.string(),
		v.trim(),
		v.maxLength(256, "Description must be at most 256 characters"),
	),
	collection_id: v.pipe(
		v.string(),
		v.trim(),
		v.uuid("Invalid collection ID"),
		v.minLength(1, "Collection ID is required"),
	),
	workspace_id: v.pipe(
		v.string(),
		v.trim(),
		v.uuid("Invalid workspace ID"),
		v.minLength(1, "Workspace ID is required"),
	),
});

type FormSchema = v.InferOutput<typeof formSchema>;

function RouteComponent() {
	const navigate = Route.useNavigate();
	const {
		workspace,
		collection,
		membership_status: membership,
	} = Route.useLoaderData();

	const form = useForm<FormSchema>({
		resolver: valibotResolver(formSchema),
		defaultValues: {
			name: collection.name,
			slug: collection.slug,
			description: collection.description ?? "",
		},
	});

	useEffect(() => {
		form.reset();
		form.setValue("name", collection.name);
		form.setValue("slug", collection.slug);
		form.setValue("description", collection.description ?? "");
	}, [collection, form]);

	const updateDetailsMutation = useRobinMutation("collection.details.update", {
		invalidates: [QueryKeys.FindWorkspace(workspace.slug), "workspace.find"],
		onSuccess: (data) => {
			navigate({
				to: "/$workspaceSlug/settings/workspace/c/$collectionSlug",
				params: {
					workspaceSlug: data.workspace_slug,
					collectionSlug: data.collection.slug,
				},
			});
			toast.success(`"${collection.name}" updated successfully`);
		},
		setFormError: form.setError,
	});

	const leaveMutation = useRobinMutation("collection.leave", {
		onSuccess: (workspaceSlug) => {
			navigate({
				to: "/$workspaceSlug/settings/workspace",
				params: { workspaceSlug: workspaceSlug },
			});
		},
		invalidates: QueryKeys.FindWorkspace(workspace.slug),
	});

	const deleteMutation = useRobinMutation("collection.delete", {
		onSuccess: (w) => {
			stores.workspace.removeCollection(collection.id);
			navigate({
				to: "/$workspaceSlug/settings/workspace",
				params: { workspaceSlug: w.slug },
			});
		},
		invalidates: QueryKeys.FindWorkspace(workspace.slug),
	});

	const name = form.watch("name");
	const slug = form.watch("slug");
	const description = form.watch("description");
	const hasUpdatedForm =
		name !== workspace.name ||
		slug !== workspace.slug ||
		description !== workspace.description;

	return (
		<PageLayout
			heading={collection.name}
			header={{
				parent: "settings",
				items: [{ title: "Collections" }, { title: collection.name }],
			}}
			showHeader
		>
			<title>
				{collection.name} - {workspace.name}
			</title>
			<Flex width="100%" direction="column" mt="4" gap="6">
				<Form.Root
					onSubmit={form.handleSubmit((data) => {
						updateDetailsMutation.call(data);
					})}
				>
					<Flex direction="column" gap="2">
						<Text size="2" color="gray">
							Change the name, URL, and description of this collection.
						</Text>
						<Flex direction="column" gap="5" mt="2">
							<Flex direction="column" gap={{ initial: "3", md: "2" }}>
								<input
									{...form.register("workspace_id")}
									type="hidden"
									value={workspace.id}
								/>
								<input
									{...form.register("collection_id")}
									type="hidden"
									value={collection.id}
								/>
								<RowInput
									register={form.register}
									name="name"
									label="Name"
									errors={form.formState.errors}
									textFieldProps={{
										disabled: membership.role !== "owner",
									}}
								/>

								<RowInput
									register={form.register}
									name="slug"
									label="URL"
									errors={form.formState.errors}
									textFieldProps={{ disabled: membership.role !== "owner" }}
									LeftSideComponent={() => (
										<Badge
											color="gray"
											size="2"
											variant="soft"
											radius="small"
											mx="-1"
										>
											{window.location.origin?.replace(/https?:\/\//, "")}/
										</Badge>
									)}
								/>

								<RowTextArea
									register={form.register}
									name="description"
									label="Description"
									placeholder="What is this collection for?"
									textAreaProps={{
										disabled:
											membership.role !== "owner" &&
											membership.role !== "admin",
									}}
									errors={form.formState.errors}
								/>
							</Flex>
						</Flex>
					</Flex>

					<Flex justify="end" mt="4">
						<Button
							type="submit"
							loading={updateDetailsMutation.isMutating}
							disabled={!hasUpdatedForm}
						>
							Save changes
						</Button>
					</Flex>
				</Form.Root>

				<Separator orientation="horizontal" style={{ width: "100%" }} />

				{/* MARK: manage collection members */}
				<Flex direction="column" align="start" gap="2">
					<Heading size="5">Members</Heading>
					<Text size="2" color="gray">
						View, find, and manage members in this collection.
					</Text>

					<Link
						to="/$workspaceSlug/settings/workspace/c/$collectionSlug/members"
						params={{
							workspaceSlug: workspace.slug,
							collectionSlug: collection.slug,
						}}
					>
						<Flex align="center" gap="2">
							<Text>View &amp; manage members</Text>
							<ArrowSquareOut size={16} />
						</Flex>
					</Link>
				</Flex>

				<Show when={membership.role !== "owner"}>
					<Separator orientation="horizontal" style={{ width: "100%" }} />

					{/* MARK: leave collection */}
					<Flex direction="column" align="start" gap="1">
						<Heading size="5">Leave collection</Heading>
						<Text size="2" color="gray">
							You will no longer have access to this collection and all of its
							content.
						</Text>

						<AlertDialog.Root>
							<AlertDialog.Trigger>
								<Button variant="solid" color="red" mt="3">
									<SignOut />
									Leave collection
								</Button>
							</AlertDialog.Trigger>

							<AlertDialog.Content maxWidth="450px">
								<AlertDialog.Title>Leave collection</AlertDialog.Title>
								<AlertDialog.Description size="2" color="gray">
									This action cannot be undone, to regain access to this
									workspace, you would need an invite.
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
										loading={leaveMutation.isMutating}
										onClick={() => {
											leaveMutation.call({
												workspace_id: workspace.id,
												collection_id: collection.id,
											});
										}}
									>
										I understand, leave collection
									</Button>
								</Flex>
							</AlertDialog.Content>
						</AlertDialog.Root>
					</Flex>
				</Show>

				<Show when={membership.role === "owner"}>
					<Separator orientation="horizontal" style={{ width: "100%" }} />

					{/* MARK: delete collection */}
					<Flex direction="column" align="start" gap="1">
						<Heading size="5">Delete collection</Heading>
						<Text size="2" color="gray">
							Delete this collection, all of its members and content; including
							all entries.
						</Text>

						<AlertDialog.Root>
							<AlertDialog.Trigger>
								<Button variant="solid" color="red" mt="3">
									<Warning />
									Delete collection
								</Button>
							</AlertDialog.Trigger>

							<AlertDialog.Content maxWidth="450px">
								<AlertDialog.Title>Delete collection</AlertDialog.Title>
								<AlertDialog.Description size="2" color="gray">
									Are you sure you want to delete "{collection.name}"? All data
									associated with this collection will be deleted.
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
										loading={deleteMutation.isMutating}
										onClick={() => {
											deleteMutation.call({
												workspace_id: workspace.id,
												collection_id: collection.id,
											});
										}}
									>
										I understand, delete collection
									</Button>
								</Flex>
							</AlertDialog.Content>
						</AlertDialog.Root>
					</Flex>
				</Show>
			</Flex>
		</PageLayout>
	);
}
