import { createLazyFileRoute } from "@tanstack/react-router";
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
import Show from "$/components/show";
import { SignOut, Trash } from "@phosphor-icons/react";
import { useSnapshot } from "valtio";
import stores from "$/stores";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/workspace/",
)({
	component: RouteComponent,
});

const formSchema = v.object({
	name: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Name must be at least 2 characters"),
		v.maxLength(64, "Name must be at most 64 characters"),
		v.regex(
			/^[a-zA-Z0-9-_ ]+$/,
			"Name must contain only letters, numbers, and spaces",
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
	description: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(512, "Cannot be more than 512 characters"),
		),
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
	const { workspace, status } = Route.useLoaderData();

	const auth = useSnapshot(stores.auth);

	const form = useForm<FormSchema>({
		resolver: valibotResolver(formSchema),
		defaultValues: {
			name: workspace.name,
			slug: workspace.slug,
			description: workspace.description ?? "",
		},
	});

	const removeMemberMutation = useRobinMutation("workspace.member.remove", {
		onSuccess: () => {
			stores.workspace.setWorkspaces([]);
			navigate({ to: "/" });
		},
		invalidates: ["me"],
	});

	const updateWorkspaceDetailsMutation = useRobinMutation(
		"workspace.details.update",
		{
			invalidates: [QueryKeys.FindWorkspace(workspace.slug), "workspace.find"],
			onSuccess: (data) => {
				navigate({
					to: "/$workspaceSlug/settings/workspace",
					params: { workspaceSlug: data.workspace.slug },
				});
				toast.success("Workspace updated successfully");
			},
			setFormError: form.setError,
		},
	);

	const deleteWorkspaceMutation = useRobinMutation("workspace.delete", {
		onSuccess: () => {
			stores.workspace.setWorkspaces(
				stores.workspace.all.filter((w) => w.id !== workspace.id),
			);
			navigate({ to: "/" });
		},
		invalidates: ["me"],
	});

	const name = form.watch("name");
	const slug = form.watch("slug");
	const description = form.watch("description");
	const hasUpdatedForm =
		name !== workspace.name ||
		slug !== workspace.slug ||
		description !== workspace.description;

	return (
		<PageLayout heading="Workspace" header={{ parent: "settings" }} showHeader>
			<Flex width="100%" direction="column" mt="4" gap="6">
				<Form.Root
					onSubmit={form.handleSubmit((data) => {
						updateWorkspaceDetailsMutation.call(data);
					})}
				>
					<Flex direction="column" gap="1">
						<Text size="2" color="gray">
							Change the name, URL, and description of your workspace.
						</Text>
						<Flex direction="column" gap="5" mt="2">
							<Flex direction="column" gap={{ initial: "3", md: "2" }}>
								<input
									{...form.register("workspace_id")}
									name="workspace_id"
									type="hidden"
									value={workspace.id}
								/>
								<RowInput
									register={form.register}
									name="name"
									label="Name"
									errors={form.formState.errors}
									textFieldProps={{ disabled: status.role !== "owner" }}
								/>

								<RowInput
									register={form.register}
									name="slug"
									label="URL"
									errors={form.formState.errors}
									textFieldProps={{ disabled: status.role !== "owner" }}
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
									placeholder="What is this workspace for?"
									textAreaProps={{
										disabled:
											status.role !== "owner" && status.role !== "admin",
									}}
									errors={form.formState.errors}
								/>
							</Flex>
						</Flex>
					</Flex>

					<Flex justify="end" mt="4">
						<Button
							type="submit"
							loading={updateWorkspaceDetailsMutation.isMutating}
							disabled={!hasUpdatedForm}
						>
							Save changes
						</Button>
					</Flex>
				</Form.Root>

				{/* MARK: Leave workspace */}
				<Show when={workspace.owner_id !== status.user_id}>
					<Separator style={{ width: "100%" }} />

					<Flex direction="column" align="start" gap="3" width="100%">
						<Text size="2" color="gray">
							Leaving this workspace will remove your access to it. You will
							need to be invited back in order to regain access.
						</Text>

						<AlertDialog.Root>
							<AlertDialog.Trigger>
								<Button variant="solid" color="red">
									<SignOut />
									Leave workspace
								</Button>
							</AlertDialog.Trigger>

							<AlertDialog.Content maxWidth="450px">
								<AlertDialog.Title>Leave workspace</AlertDialog.Title>
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
										loading={removeMemberMutation.isMutating}
										onClick={() => {
											removeMemberMutation.call({
												workspace_id: workspace.id,
												email: auth.user?.email ?? "",
											});
										}}
									>
										I understand, leave workspace
									</Button>
								</Flex>
							</AlertDialog.Content>
						</AlertDialog.Root>
					</Flex>
				</Show>

				{/* MARK: Delete workspace */}
				<Show when={workspace.owner_id === status.user_id}>
					<Separator style={{ width: "100%" }} />

					<Flex direction="column" align="start" gap="1" width="100%">
						<Heading size="4" color="gray" highContrast>
							Delete workspace
						</Heading>

						<Text size="2" color="gray">
							This action will permanently delete this workspace and all data
							associated with it. This action cannot be undone.
						</Text>

						<AlertDialog.Root>
							<AlertDialog.Trigger>
								<Button variant="solid" color="red" mt="3">
									<Trash />
									Permanently delete workspace
								</Button>
							</AlertDialog.Trigger>

							<AlertDialog.Content maxWidth="450px">
								<AlertDialog.Title>Delete workspace</AlertDialog.Title>
								<AlertDialog.Description size="2" color="gray">
									This action cannot be undone. All data associated with this
									workspace will be permanently deleted.
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
										onClick={() => {
											deleteWorkspaceMutation.call({
												workspace_id: workspace.id,
											});
										}}
									>
										I understand, delete workspace
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
