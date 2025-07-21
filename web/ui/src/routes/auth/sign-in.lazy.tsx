import {
	Box,
	Button,
	Card,
	Flex,
	Heading,
	IconButton,
	Text,
} from "@radix-ui/themes";
import { createLazyFileRoute, Link, useNavigate } from "@tanstack/react-router";
import * as Form from "@radix-ui/react-form";
import { useForm } from "react-hook-form";
import Input from "$/components/form/input";
import * as v from "valibot";
import { toast } from "sonner";
import { toTitleCase } from "$/lib/utils";
import { useState } from "react";
import { Eye, EyeSlash } from "@phosphor-icons/react";
import { useRobinMutation } from "$/lib/hooks";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { useSnapshot } from "valtio";
import stores from "$/stores";

const signInSchema = v.object({
	email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
	password: v.pipe(
		v.string(),
		v.nonEmpty("Password is required"),
		v.minLength(6),
	),
});

type SignInSchema = v.InferOutput<typeof signInSchema>;

function SignIn() {
	const search = Route.useSearch();
	const workspace = useSnapshot(stores.workspace);
	const user = useSnapshot(stores.auth);

	const navigate = useNavigate();
	const {
		register,
		formState: { errors, isSubmitting },
		handleSubmit,
		setError,
	} = useForm<SignInSchema>({
		resolver: valibotResolver(signInSchema),
		defaultValues: { email: search.email, password: "" },
	});

	const [showPassword, setShowPassword] = useState(false);

	const loginMutation = useRobinMutation("auth.sign-in", {
		setFormError: setError,
		onSuccess: (data) => {
			if (data.requires_email_verification) {
				toast.success("Please verify your email to continue");
				return navigate({
					to: "/auth/verify-email",
					search: { email: data.email },
				});
			}

			if (data.mfa.enabled) {
				toast.success("You need to verify this sign-in");
				return navigate({ to: `/auth/mfa/${data.mfa.session_id}` });
			}

			if (data.user) user.setUser(data?.user);
			if (data.workspaces) workspace.setWorkspaces(data.workspaces);
			if (!data.workspaces || data.workspaces.length === 0) {
				return navigate({ to: "/workspace/new" });
			}

			toast.success(
				`Welcome back, ${toTitleCase(data?.user?.first_name ?? "")}`,
			);

			if (search.redirect) {
				return navigate({ to: search.redirect, replace: true });
			}

			return navigate({ to: "/" });
		},
	});

	return (
		<Flex
			width="100vw"
			height={{ initial: "100dvh", xs: "100vh" }}
			align="center"
			justify="center"
		>
			<title>Sign In</title>
			<Card variant="ghost">
				<Flex direction="column" width={{ initial: "85vw", xs: "325px" }} p="3">
					<Heading size="8" weight="bold">
						Sign in
					</Heading>

					<Flex direction="column" gap="2" mt="4">
						<Form.Root onSubmit={handleSubmit((d) => loginMutation.call(d))}>
							<Flex direction="column" gap="2">
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
									name="password"
									type={showPassword ? "text" : "password"}
									label="Password"
									errors={errors}
									RightSideComponent={() => (
										<IconButton
											type="button"
											variant="ghost"
											onClick={() => setShowPassword(!showPassword)}
										>
											{showPassword ? <EyeSlash /> : <Eye />}
										</IconButton>
									)}
									required
								/>
								<Flex
									align="center"
									justify="end"
									className="transition-all"
									mt="-1"
								>
									<Link to="/auth/forgot-password" className="link">
										<Text size="1">Forgot password?</Text>
									</Link>
								</Flex>
							</Flex>

							<Button
								type="submit"
								loading={isSubmitting}
								my="4"
								style={{ width: "100%" }}
							>
								Sign in
							</Button>

							<Box>
								<Text size="2">
									Don't have an account?{" "}
									<Link to="/auth/sign-up" className="link">
										Create one
									</Link>
								</Text>
							</Box>
						</Form.Root>
					</Flex>
				</Flex>
			</Card>
		</Flex>
	);
}

export const Route = createLazyFileRoute("/auth/sign-in")({
	component: SignIn,
});
