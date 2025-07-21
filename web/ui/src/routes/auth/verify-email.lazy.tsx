import { Box, Button, Card, Flex, Heading, Text } from "@radix-ui/themes";
import { createLazyFileRoute, Link, useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import Input from "$/components/form/input";
import * as Form from "@radix-ui/react-form";
import { useForm } from "react-hook-form";

import {
	InputOTP,
	InputOTPGroup,
	InputOTPSlot,
} from "$/components/form/input-otp";
import FieldError from "$/components/form/field-error";
import { ArrowLeft } from "@phosphor-icons/react";
import { REGEXP_ONLY_DIGITS_AND_CHARS } from "input-otp";
import { useRobinMutation } from "$/lib/hooks";

function VerifyEmail() {
	const { email, token } = Route.useSearch();
	const navigate = useNavigate();
	const { formState, register, handleSubmit, setValue, setError } = useForm({
		defaultValues: { email, token: token ?? "" },
	});

	const verificationMutation = useRobinMutation("auth.verify-email", {
		onSuccess: () => {
			toast.success("Email verified successfully");
			navigate({
				to: "/auth/sign-in",
				search: {
					email: email,
				},
			});
		},
		setFormError: setError,
		retry: false,
	});

	const resendMutation = useRobinMutation("auth.request-email-verification", {
		onSuccess: (message) => toast.success(message),
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
			<title>Verify Email</title>
			<Card variant="ghost">
				<Flex
					direction="column"
					width={{ initial: "85vw", xs: "350px" }}
					py="3"
					px="2"
					gap="3"
				>
					<Heading size="8" weight="bold">
						Verify your email
					</Heading>

					<Text color="gray" size="2">
						Please check your inbox and click on the link we sent you to verify
						your email address or enter the token below.
					</Text>

					<Form.Root
						onSubmit={handleSubmit((d) => verificationMutation.call(d))}
					>
						<Input
							register={register}
							name="email"
							label="Email"
							textFieldProps={{ readOnly: true }}
							errors={formState.errors}
						/>

						<Box mt="3" mb="5">
							<label htmlFor="token">
								<Text size="2" weight="medium" color="gray">
									Code
								</Text>
							</label>
							<Flex justify="center" mt="1">
								<InputOTP
									{...register("token", {
										required: "The One-Time Password is required",
									})}
									type="text"
									inputMode="text"
									maxLength={8}
									onChange={(value) => setValue("token", value)}
									pattern={REGEXP_ONLY_DIGITS_AND_CHARS}
								>
									<InputOTPGroup>
										<InputOTPSlot index={0} />
										<InputOTPSlot index={1} />
										<InputOTPSlot index={2} />
										<InputOTPSlot index={3} />
									</InputOTPGroup>
									<InputOTPGroup>
										<InputOTPSlot index={4} />
										<InputOTPSlot index={5} />
										<InputOTPSlot index={6} />
										<InputOTPSlot index={7} />
									</InputOTPGroup>
								</InputOTP>
							</Flex>

							<FieldError errors={formState.errors} name="token" />
						</Box>

						<Button
							type="submit"
							style={{ width: "100%" }}
							loading={formState.isSubmitting}
						>
							Verify email
						</Button>

						<Button
							type="button"
							variant="ghost"
							mt="3"
							onClick={() => resendMutation.call(email)}
							style={{ width: "100%" }}
							loading={resendMutation.isMutating}
						>
							Didn't receive the email?
						</Button>
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

export const Route = createLazyFileRoute("/auth/verify-email")({
	component: VerifyEmail,
});
