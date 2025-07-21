import {
	Badge,
	Button,
	Callout,
	Dialog,
	Flex,
	Switch,
	Text,
	Tooltip,
} from "@radix-ui/themes";
import * as Form from "@radix-ui/react-form";
import * as v from "valibot";
import { useForm } from "react-hook-form";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { useRobinMutation } from "$/lib/hooks";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import { MixedNameRegex } from "$/lib/utils";
import Input from "../form/input";
import CustomTextArea from "../form/custom-text-area";
import FieldError from "../form/field-error";
import QueryKeys from "$/lib/keys";
import { useNavigate, useParams } from "@tanstack/react-router";
import { useMemo } from "react";

const newCollectionSchema = v.object({
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
	slug: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.minLength(2, "Slug must be at least 2 characters"),
			v.maxLength(80, "Slug must be at most 80 characters"),
			v.regex(
				/^[a-zA-Z0-9-]*$/,
				"Slug must contain only letters, numbers, and hyphens",
			),
		),
	),
	description: v.pipe(
		v.string(),
		v.trim(),
		v.maxLength(256, "Description must be at most 256 characters"),
	),
	assign_all_members: v.boolean(),
});

type NewCollectionSchema = v.InferOutput<typeof newCollectionSchema>;

type Props = {
	open: boolean;
	onClose: () => void;
	onOpen: () => void;
};

export default function NewCollectionDialog(props: Props) {
	const navigate = useNavigate();
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const workspace = useSnapshot(stores.workspace);
	const app = useSnapshot(stores.app);

	const form = useForm<NewCollectionSchema>({
		resolver: valibotResolver(newCollectionSchema),
		defaultValues: {
			assign_all_members: true,
		},
	});

	const mutation = useRobinMutation("collection.create", {
		setFormError: form.setError,
		invalidates: [QueryKeys.FindWorkspace(params.workspaceSlug), "me"],
		retry: false,
		onSuccess: (data) => {
			workspace.addCollection(data.collection);

			navigate({
				to: "/$workspaceSlug/c/$collectionSlug",
				params: {
					workspaceSlug: data.workspace_slug,
					collectionSlug: data.collection.slug,
				},
			});

			form.reset();
			props.onClose();
		},
	});

	const name = form.watch("name");
	const description = form.watch("description");
	const assignAllMembers = form.watch("assign_all_members");
	const hasEditedForm = useMemo(() => {
		return (
			(name !== undefined && name !== "") ||
			(description !== undefined && description !== "")
		);
	}, [description, name]);

	const workspaceName = useMemo(
		() =>
			workspace.findWorkspaceBySlug(params.workspaceSlug)?.name ??
			"current workspace",
		[params.workspaceSlug, workspace],
	);

	async function onSubmit(data: NewCollectionSchema) {
		await mutation.call({
			workspace_slug: params.workspaceSlug,
			...data,
		});
	}

	return (
		<Dialog.Root
			open={props.open}
			onOpenChange={(value) => {
				form.reset();

				// Close the dialog if the user is not in the middle of a mutation and has not entered any data
				if (!value && !mutation.isMutating && !hasEditedForm) {
					props.onClose();
				}
			}}
		>
			<Dialog.Content maxWidth="400px">
				<Tooltip content="Current workspace">
					<Badge color={app.accentColor} variant="surface">
						{workspaceName}
					</Badge>
				</Tooltip>

				<Dialog.Title mt="3" mb="1" size="6">
					New collection
				</Dialog.Title>
				<Dialog.Description size="2" color="gray" mb="4">
					Create a new collection to organize your files.
				</Dialog.Description>
				<Form.Root onSubmit={form.handleSubmit(onSubmit)}>
					<Flex direction="column" gap="4">
						<Input
							register={form.register}
							name="name"
							type="text"
							label="Name"
							errors={form.formState.errors}
							textFieldProps={{ autoFocus: true }}
							required
						/>

						<CustomTextArea
							register={form.register}
							name="description"
							label="Description"
							placeholder="e.g. Research papers for the next project"
							errors={form.formState.errors}
						/>

						<Flex direction="column" gap="2">
							<Flex align="center">
								<Switch
									size="2"
									checked={assignAllMembers}
									onCheckedChange={(v) =>
										form.setValue("assign_all_members", v)
									}
								/>
								<Text size="2" color="gray" ml="2">
									Assign all members
								</Text>
							</Flex>
							<Callout.Root size="1" color="amber" variant="surface">
								<Callout.Text size="1">
									{assignAllMembers
										? "All users in the workspace will have access to this collection."
										: "Only you will have access to this collection until you share it with others."}
								</Callout.Text>
							</Callout.Root>
							<FieldError
								name="assign_all_members"
								errors={form.formState.errors}
							/>
						</Flex>

						{/* MARK: Buttons */}
						<Flex direction="row-reverse" gap="3">
							<Button type="submit" loading={form.formState.isSubmitting}>
								Create
							</Button>

							<Button
								type="button"
								variant="soft"
								color="gray"
								onClick={props.onClose}
							>
								Cancel
							</Button>
						</Flex>
					</Flex>
				</Form.Root>
			</Dialog.Content>
		</Dialog.Root>
	);
}
