import { Flex, Dialog, Button } from "@radix-ui/themes";
import { useForm } from "react-hook-form";
import * as v from "valibot";
import * as Form from "@radix-ui/react-form";
import CustomTextArea from "../form/custom-text-area";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { useRobinMutation } from "$/lib/hooks";
import { toast } from "sonner";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import { useSearch } from "@tanstack/react-router";
import QueryKeys from "$/lib/keys";

type Props = {
	open: boolean;
	onOpenChange: (value: boolean) => void;
};

const formSchema = v.object({
	workspace_id: v.pipe(v.string(), v.nonEmpty(), v.uuid()),
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

type FormSchema = v.InferOutput<typeof formSchema>;

export default function InviteDialog(props: Props) {
	const search = useSearch({
		from: "/_protected/$workspaceSlug/settings/_settings/workspace/members",
	});
	const workspace = useSnapshot(stores.workspace);
	const form = useForm<FormSchema>({
		resolver: valibotResolver(formSchema),
		defaultValues: { emails: [] },
	});

	const mutation = useRobinMutation("workspace.invite", {
		onSuccess: (data) => {
			toast.success(data.message);
			props.onOpenChange(false);
		},
		setFormError: form.setError,
		invalidates: QueryKeys.ListWorkspaceMembers(
			workspace.activeWorkspace?.id ?? "",
			search.page,
		),
	});

	function onSubmit(data: FormSchema) {
		mutation.call(data);
	}

	return (
		<Dialog.Root
			open={props.open}
			onOpenChange={(v) => {
				props.onOpenChange(v);
				form.reset();
			}}
		>
			<Dialog.Trigger>
				<Button variant="solid" size="2">
					Invite members
				</Button>
			</Dialog.Trigger>
			<Dialog.Content maxWidth="450px">
				<Dialog.Title>Invite to workspace</Dialog.Title>
				<Dialog.Description size="2" color="gray" mb="4">
					Enter the email addresses of the people you want to invite to this
					workspace.
				</Dialog.Description>

				<Form.Root onSubmit={form.handleSubmit(onSubmit)}>
					<input
						{...form.register("workspace_id")}
						type="hidden"
						value={workspace.activeWorkspace?.id}
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
							Invite
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
