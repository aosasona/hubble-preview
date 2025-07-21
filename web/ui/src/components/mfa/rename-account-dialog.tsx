import type { QueryResult } from "$/lib/server";
import { valibotResolver } from "@hookform/resolvers/valibot";
import * as Form from "@radix-ui/react-form";
import { Button, Dialog, Flex, Text } from "@radix-ui/themes";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as v from "valibot";
import Input from "../form/input";
import { redactEmail } from "$/lib/utils";
import { useRobinMutation } from "$/lib/hooks";

type Account = QueryResult<"mfa.state">["accounts"][0];
type Props = {
	account: Account | null;
	onClose: () => void;
};

const renameFormSchema = v.object({
	name: v.pipe(
		v.string(),
		v.nonEmpty("Name is required"),
		v.maxLength(50),
		v.regex(
			/^[a-zA-Z0-9-_ ]+$/,
			"Only letters, numbers, spaces, hyphens, and underscores are allowed",
		),
	),
});

type RenameFormSchema = v.InferOutput<typeof renameFormSchema>;

export default function RenameAccountDialog(props: Props) {
	const form = useForm<RenameFormSchema>({
		resolver: valibotResolver(renameFormSchema),
		defaultValues: {
			name: props.account?.meta?.name ?? "",
		},
	});

	const mutation = useRobinMutation("mfa.rename-account", {
		onSuccess: () => {
			toast.success("Account renamed successfully");
			props.onClose();
		},
		invalidates: ["mfa.state"],
		setFormError: form.setError,
	});

	async function onSubmit(data: RenameFormSchema) {
		await mutation.call({
			account_id: props?.account?.id ?? "",
			name: data.name,
		});
	}

	const newName = form.watch("name");

	return (
		<Dialog.Root
			open={props.account != null}
			onOpenChange={(open) => {
				if (!open) {
					props.onClose();
				}
				form.reset();
			}}
		>
			<Dialog.Content maxWidth="400px">
				<Dialog.Title size="6">Rename account</Dialog.Title>
				<Dialog.Description>
					<Text size="2" color="gray">
						{props?.account?.type === "email"
							? // @ts-expect-error this will always exist as long as the account is an email
								`Set a new alias for ${redactEmail(props.account.meta?.email) ?? "this email account"} to easily identify it later.`
							: "Set a name for this account to easily identify it later."}
					</Text>
				</Dialog.Description>

				<Form.Root onSubmit={form.handleSubmit(onSubmit)}>
					<Flex direction="column" gap="5" mt="3">
						<Input
							register={form.register}
							type="text"
							name="name"
							label="Name"
							errors={form.formState.errors}
							textFieldProps={{ autoFocus: true }}
							required
						/>

						<Flex gap="2" align="center" justify="end">
							<Dialog.Close>
								<Button type="button" variant="surface" color="gray">
									Cancel
								</Button>
							</Dialog.Close>

							<Button
								type="submit"
								loading={form.formState.isSubmitting}
								disabled={newName === props.account?.meta?.name}
							>
								Save
							</Button>
						</Flex>
					</Flex>
				</Form.Root>
			</Dialog.Content>
		</Dialog.Root>
	);
}
