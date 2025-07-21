import type { QueryResult } from "$/lib/server";
import { Dialog, Button, Flex, Callout } from "@radix-ui/themes";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { useForm } from "react-hook-form";
import * as Form from "@radix-ui/react-form";
import Input from "../form/input";
import { Info, Warning } from "@phosphor-icons/react";
import { useRobinMutation } from "$/lib/hooks";

type Account = QueryResult<"mfa.state">["accounts"][0];
type Props = {
	account: Account | null;
	onClose: () => void;
};

const formSchema = v.object({
	password: v.pipe(
		v.string(),
		v.nonEmpty("Password is required"),
		v.minLength(6, "Password must be at least 6 characters"),
	),
});

type FormSchema = v.InferOutput<typeof formSchema>;

export default function DeleteAccountDialog(props: Props) {
	const form = useForm<FormSchema>({
		resolver: valibotResolver(formSchema),
	});

	const mutation = useRobinMutation("mfa.delete-account", {
		onSuccess: () => {
			props.onClose();
		},
		invalidates: ["mfa.state"],
		setFormError: form.setError,
	});

	async function onSubmit(data: FormSchema) {
		await mutation.call({
			account_id: props.account?.id ?? "",
			password: data.password,
		});
	}

	return (
		<Dialog.Root
			open={props.account != null}
			onOpenChange={(open) => {
				if (!open) props.onClose();
			}}
		>
			<Dialog.Content maxWidth="385px">
				<Dialog.Title>
					Remove{" "}
					{props?.account?.meta?.name
						? `"${props.account.meta.name}"`
						: props?.account?.type === "email"
							? "email"
							: "authenticator"}
				</Dialog.Title>
				<Dialog.Description size="2">
					If you choose to continue, this{" "}
					{props.account?.type === "email" ? "email" : "authenticator"} will no
					longer be available for multi-factor authentication.
				</Dialog.Description>

				<Form.Root onSubmit={form.handleSubmit(onSubmit)}>
					<Flex direction="column" gap="5" mt="3">
						<Flex direction="column" gap="2" mt="3">
							<Input
								register={form.register}
								type="password"
								name="password"
								label="Password"
								errors={form.formState.errors}
								textFieldProps={{ autoFocus: true }}
								required
							/>
							<Callout.Root color="red" size="1" variant="surface">
								<Callout.Icon>
									<Info />
								</Callout.Icon>
								<Callout.Text>Enter your password to continue</Callout.Text>
							</Callout.Root>
						</Flex>

						<Flex gap="2" align="center" justify="end">
							<Dialog.Close>
								<Button type="button" variant="surface" color="gray">
									Cancel
								</Button>
							</Dialog.Close>

							<Button type="submit" loading={form.formState.isSubmitting}>
								<Warning /> Remove
							</Button>
						</Flex>
					</Flex>
				</Form.Root>
			</Dialog.Content>
		</Dialog.Root>
	);
}
