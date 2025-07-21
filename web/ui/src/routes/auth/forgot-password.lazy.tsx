import { Button, Card, Flex, Heading, Text } from "@radix-ui/themes";
import { createLazyFileRoute, Link, useNavigate } from "@tanstack/react-router";
import * as Form from "@radix-ui/react-form";
import { useForm } from "react-hook-form";
import { ArrowLeft } from "@phosphor-icons/react";
import Input from "$/components/form/input";
import { toast } from "sonner";
import { useRobinMutation } from "$/lib/hooks";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";

export const Route = createLazyFileRoute("/auth/forgot-password")({
	component: ForgotPassword,
});

const forgotPasswordSchema = v.object({
	email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
	scope: v.pipe(
		v.picklist(["reset", "change"] as const),
		v.nonEmpty("Scope is required"),
	),
});

type ForgotPasswordSchema = v.InferOutput<typeof forgotPasswordSchema>;

function ForgotPassword() {
	const navigate = useNavigate();
	const { formState, register, handleSubmit, setError, getValues } =
		useForm<ForgotPasswordSchema>({
			resolver: valibotResolver(forgotPasswordSchema),
			defaultValues: { email: "", scope: "reset" },
		});

	const mutation = useRobinMutation("auth.request-password-reset", {
		onSuccess: (data) => {
			toast.success(data.message);
			navigate({
				to: "/auth/change-password",
				search: { email: getValues("email") },
			});
		},
		setFormError: setError,
		retry: false,
	});

	return (
		<Flex
			direction="column"
			width="100vw"
			height={{ initial: "100dvh", xs: "100vh" }}
			align="center"
			justify="center"
			gap="4"
		>
			<title>Reset Password</title>
			<Card variant="ghost">
				<Flex
					direction="column"
					width={{ initial: "85vw", xs: "325px" }}
					py="3"
					px="2"
					gap="3"
				>
					<Heading size="8" weight="bold">
						Reset password
					</Heading>

					<Text color="gray" size="2">
						Enter your email address and we'll send you a link to reset your
						password.
					</Text>

					<Form.Root onSubmit={handleSubmit((d) => mutation.call(d))}>
						<Flex direction="column" gap="4">
							<Input
								register={register}
								name="email"
								label="Email"
								errors={formState.errors}
							/>

							<Button type="submit" loading={formState.isSubmitting}>
								Continue
							</Button>
						</Flex>
					</Form.Root>
				</Flex>
			</Card>

			<Flex mb="2">
				<Link to="/auth/sign-in" className="flex items-center gap-1">
					<ArrowLeft />
					<Text size="2">Back to sign in</Text>
				</Link>
			</Flex>
		</Flex>
	);
}
