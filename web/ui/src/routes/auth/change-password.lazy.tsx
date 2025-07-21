import * as Form from "@radix-ui/react-form";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { createLazyFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { Button, Card, Flex, Heading, Text } from "@radix-ui/themes";
import { useForm } from "react-hook-form";
import * as v from "valibot";
import Input from "$/components/form/input";
import { toast } from "sonner";
import Show from "$/components/show";
import { ArrowLeft } from "@phosphor-icons/react";
import { useRobinMutation } from "$/lib/hooks";

enum Scope {
	Reset = "reset",
	Change = "change",
}

export const Route = createLazyFileRoute("/auth/change-password")({
	component: RouteComponent,
});

const formSchema = v.pipe(
	v.object({
		email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
		token: v.pipe(
			v.string(),
			v.nonEmpty("Token is required"),
			v.regex(/^[a-zA-Z0-9]{8}$/, "Invalid token."),
		),
		new_password: v.pipe(v.string(), v.nonEmpty("Password is required")),
		confirm_password: v.pipe(
			v.string(),
			v.nonEmpty("Password confirmation is required"),
		),
	}),
	v.forward(
		v.partialCheck(
			[["new_password"], ["confirm_password"]],
			(input) => input.new_password === input.confirm_password,
			"Passwords do not match.",
		),
		["confirm_password"],
	),
);

type ChangePasswordFormSchema = v.InferOutput<typeof formSchema>;

function RouteComponent() {
	const navigate = useNavigate();
	const queryParams = Route.useSearch();
	const {
		formState: { errors, isSubmitting },
		register,
		...form
	} = useForm<ChangePasswordFormSchema>({
		resolver: valibotResolver(formSchema),
		defaultValues: {
			email: queryParams.email,
			token: queryParams.token || "",
		},
	});

	const resendMutation = useRobinMutation("auth.request-password-reset", {
		onComplete: () => form.setValue("token", ""),
		setFormError: form.setError,
		retry: false,
	});

	const passwordChangeMutation = useRobinMutation("auth.change-password", {
		onSuccess: (data) => {
			toast.success(data.message);
			navigate({ to: data.scope === Scope.Reset ? "/auth/sign-in" : "/" });
		},
		setFormError: form.setError,
		retry: false,
	});

	async function onSubmit(data: ChangePasswordFormSchema) {
		await passwordChangeMutation.call({
			...data,
			current_password: "",
			scope: queryParams.scope || Scope.Reset,
		});
	}

	const watchedToken = form.watch("token");

	return (
		<Flex
			width="100vw"
			direction="column"
			height={{ initial: "100dvh", xs: "100vh" }}
			align="center"
			justify="center"
		>
			<title>Set A New Password</title>
			<Card variant="ghost">
				<Flex
					direction="column"
					width={{ initial: "85vw", xs: "350px" }}
					py="3"
					px="2"
				>
					<Heading size="8" weight="bold">
						Change Password
					</Heading>

					<Text size="2" color="gray" mt="2">
						Complete the form below to change your password.
					</Text>

					<Flex direction="column" gap="2" mt="4">
						<Form.Root onSubmit={form.handleSubmit(onSubmit)}>
							<Flex direction="column" gap="2">
								<Input
									register={register}
									name="email"
									label="E-mail Address"
									errors={errors}
									required
									textFieldProps={{ readOnly: true }}
								/>

								<Input
									register={register}
									name="token"
									label="One-Time Token"
									errors={errors}
									textFieldProps={{
										maxLength: 8,
										minLength: 8,
										autoComplete: "off",
										autoCapitalize: "off",
										autoFocus: true,
									}}
									RightSideComponent={() => (
										<Button
											type="button"
											variant="ghost"
											loading={resendMutation.isMutating}
											onClick={() =>
												resendMutation.call({
													email: queryParams.email,
													scope: queryParams.scope || Scope.Reset,
												})
											}
										>
											Resend
										</Button>
									)}
									required
								/>

								<Show when={watchedToken?.length === 8}>
									<Input
										register={register}
										type="password"
										name="new_password"
										label="New Password"
										errors={errors}
										textFieldProps={{
											autoFocus: true,
											autoComplete: "new-password",
										}}
										required
									/>

									<Input
										register={register}
										type="password"
										name="confirm_password"
										label="Confirm Password"
										errors={errors}
										textFieldProps={{
											autoComplete: "new-password",
										}}
										required
									/>
								</Show>
							</Flex>

							<Button
								type="submit"
								mt="6"
								style={{ width: "100%" }}
								loading={isSubmitting}
								disabled={
									isSubmitting || !watchedToken || resendMutation.isMutating
								}
							>
								Continue
							</Button>
						</Form.Root>
					</Flex>
				</Flex>
			</Card>

			<Flex mt="4">
				<Link to="/auth/sign-in" className="flex items-center gap-1">
					<ArrowLeft />
					<Text size="2">Back to sign in</Text>
				</Link>
			</Flex>
		</Flex>
	);
}
