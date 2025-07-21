import {
	Button,
	Card,
	Flex,
	Heading,
	IconButton,
	Text,
} from "@radix-ui/themes";
import { createLazyFileRoute, Link, useNavigate } from "@tanstack/react-router";
import * as Form from "@radix-ui/react-form";
import { useForm, useWatch } from "react-hook-form";
import Input from "$/components/form/input";
import { toast } from "sonner";
import { useEffect, useState } from "react";
import { Eye, EyeSlash } from "@phosphor-icons/react";
import { useRobinMutation } from "$/lib/hooks";
import * as v from "valibot";
import { toTitleCase } from "$/lib/utils";
import { valibotResolver } from "@hookform/resolvers/valibot";

const signUpSchema = v.object({
	first_name: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "First name must be at least 2 characters"),
		v.maxLength(50, "First name must be at most 50 characters"),
		v.regex(/^[a-zA-Z\s]*$/, "First name must contain only letters"),
		v.transform((value) => toTitleCase(value)),
	),
	last_name: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Last name must be at least 2 characters"),
		v.maxLength(50, "Last name must be at most 50 characters"),
		v.regex(/^[a-zA-Z\s]*$/, "Last name must contain only letters"),
		v.transform((value) => toTitleCase(value)),
	),
	email: v.pipe(v.string(), v.trim(), v.email("Invalid email address")),
	username: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, "Username must be at least 2 characters"),
		v.maxLength(24, "Username must be at most 24 characters"),
		v.regex(
			/^[a-zA-Z0-9_]*$/,
			"Username must contain only letters, numbers, and underscores",
		),
	),
	password: v.pipe(v.string(), v.nonEmpty("Password is required")),
});

type SignUpSchema = v.InferOutput<typeof signUpSchema>;

function SignUp() {
	const navigate = useNavigate();
	const {
		register,
		formState: { errors, isSubmitting },
		handleSubmit,
		setValue,
		setError,
		getValues,
		control,
	} = useForm<SignUpSchema>({
		resolver: valibotResolver(signUpSchema),
	});
	const email = useWatch({ name: "email", control });

	const [showPassword, setShowPassword] = useState(false);

	useEffect(() => {
		if (!getValues("username") && email !== undefined && email?.includes("@")) {
			setValue("username", email?.split("@")[0]);
		}
	}, [email, getValues, setValue]);

	const signUpMutation = useRobinMutation("auth.sign-up", {
		onSuccess: () => {
			toast.success("Account created successfully");
			navigate({
				to: "/auth/verify-email",
				search: { email: getValues("email") },
			});
		},
		setFormError: setError,
		retry: false,
	});

	return (
		<Flex
			width="100vw"
			height={{ initial: "100dvh", xs: "100vh" }}
			align="center"
			justify="center"
		>
			<title>Sign Up</title>
			<Card variant="ghost">
				<Flex direction="column" width={{ initial: "85vw", xs: "325px" }} p="3">
					<Heading size="8" weight="bold">
						Sign up
					</Heading>

					<Flex direction="column" gap="2" mt="4">
						<Form.Root onSubmit={handleSubmit((d) => signUpMutation.call(d))}>
							<Flex direction="column" gap="2">
								<Input
									register={register}
									name="first_name"
									label="First Name"
									errors={errors}
									required
								/>
								<Input
									register={register}
									name="last_name"
									label="Last Name"
									errors={errors}
									required
								/>
								<Input
									register={register}
									type="email"
									name="email"
									label="Email"
									errors={errors}
									required
								/>
								<Input
									register={register}
									name="username"
									label="Username"
									errors={errors}
									required
								/>

								<Flex align="center">
									<Input
										register={register}
										name="password"
										type={showPassword ? "text" : "password"}
										label="Password"
										errors={errors}
										registerOptions={{
											minLength: {
												value: 6,
												message: "Password must be at least 6 characters",
											},
										}}
										required
										RightSideComponent={() => (
											<IconButton
												aria-label="Toggle password visibility"
												color="gray"
												type="button"
												onClick={() => setShowPassword(!showPassword)}
												variant="ghost"
											>
												{showPassword ? <EyeSlash /> : <Eye />}
											</IconButton>
										)}
									/>
								</Flex>
							</Flex>

							<Button
								type="submit"
								mt="5"
								style={{ width: "100%" }}
								loading={isSubmitting}
							>
								Sign up
							</Button>

							<Flex direction="column" align="center" mt="4" mx="auto">
								<Text size="2" align="center">
									Already have an account?{" "}
									<Link to="/auth/sign-in" className="link">
										Sign in
									</Link>
								</Text>
							</Flex>
						</Form.Root>
					</Flex>
				</Flex>
			</Card>
		</Flex>
	);
}

export const Route = createLazyFileRoute("/auth/sign-up")({
	component: SignUp,
});
