import { ArrowClockwise, Key, Mailbox, Warning } from "@phosphor-icons/react";
import {
	Box,
	Button,
	Callout,
	Card,
	Flex,
	Heading,
	RadioCards,
	ScrollArea,
	Separator,
	Skeleton,
	Strong,
	Text,
} from "@radix-ui/themes";
import { Link } from "@tanstack/react-router";
import { createLazyFileRoute } from "@tanstack/react-router";
import * as Form from "@radix-ui/react-form";
import { useForm } from "react-hook-form";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { REGEXP_ONLY_DIGITS, REGEXP_ONLY_DIGITS_AND_CHARS } from "input-otp";
import FieldError from "$/components/form/field-error";
import {
	InputOTP,
	InputOTPGroup,
	InputOTPSeparator,
	InputOTPSlot,
} from "$/components/form/input-otp";
import QueryKeys from "$/lib/keys";
import { toast } from "sonner";
import { redactEmail } from "$/lib/utils";
import Show from "$/components/show";
import { useMemo } from "react";
import { useRobinMutation } from "$/lib/hooks";
import stores from "$/stores";
import { useSnapshot } from "valtio";

export const Route = createLazyFileRoute("/auth/mfa/$sessionId")({
	component: RouteComponent,
});

const mfaFormSchema = v.object({
	sessionId: v.string(),
	code: v.union([
		v.pipe(v.string(), v.regex(/^[a-zA-Z0-9_]{8}$/, "Invalid code provided")),
		v.pipe(
			v.string(),
			v.transform((v) => v.replace(/\D/g, "")),
			v.minLength(6),
			v.maxLength(6),
		),
	]),
});

type MfaFormSchema = v.InferOutput<typeof mfaFormSchema>;

function RouteComponent() {
	const { session, accounts } = Route.useLoaderData();
	const { useBackupCode } = Route.useSearch();
	const navigate = Route.useNavigate();

	const auth = useSnapshot(stores.auth);
	const workspaces = useSnapshot(stores.workspace);

	const {
		register,
		handleSubmit,
		setValue,
		formState: { errors, isSubmitting },
	} = useForm<MfaFormSchema>({
		resolver: valibotResolver(mfaFormSchema),
		defaultValues: {
			sessionId: session.id,
		},
	});

	const isEmail = useMemo(() => session.type === "email", [session.type]);

	const initiationMutation = useRobinMutation("mfa.initiate-auth-session", {
		mutationKey: QueryKeys.InitiateMfaAuthSession(session.id),
		onSuccess: (data) => {
			toast.success("Session initiated successfully");
			return navigate({ to: `/auth/mfa/${data.session_id}` });
		},
	});

	const verificationMutation = useRobinMutation("mfa.verify-auth-session", {
		mutationKey: QueryKeys.VerifyMfaAuthSession(session.id),
		onSuccess: (data) => {
			if (data.user) auth.setUser(data?.user);
			if (data.workspaces) workspaces.setWorkspaces(data.workspaces);

			if (!data.workspaces || data.workspaces.length === 0) {
				return navigate({ to: "/workspace/new" });
			}

			toast.success(data.message);

			return navigate({ to: "/" });
		},
	});

	const resendMutation = useRobinMutation("mfa.resend-email");

	async function onSubmit(data: MfaFormSchema) {
		await verificationMutation.call({
			session_id: session.id,
			code: data.code,
			use_backup_code: useBackupCode,
		});
	}

	return (
		<Flex
			direction="column"
			width="100vw"
			minHeight={{ initial: "100vh", xs: "100vh" }}
			align="center"
			justify="center"
		>
			<title>Confirm Sign-in</title>
			<Card variant="ghost">
				<Skeleton loading={initiationMutation.isMutating}>
					<Flex
						direction="column"
						gap="2"
						width={{ initial: "90vw", xs: "345px" }}
						p="3"
					>
						<Heading size={{ initial: "5", md: "6" }} weight="bold">
							Two-factor Authentication
						</Heading>

						{useBackupCode ? (
							<Text color="gray" size="2">
								To complete the sign-in process, please enter one of the backup
								codes provided to you during setup.
							</Text>
						) : (
							<Text color="gray" size="2">
								To complete the sign-in process, please enter the code{" "}
								<Show when={isEmail}>
									sent to{" "}
									{session.meta.email ? (
										<Strong>{redactEmail(session.meta?.email)}</Strong>
									) : (
										"your email"
									)}
								</Show>
								<Show when={!isEmail}>from your authenticator app</Show>
							</Text>
						)}

						<Form.Root onSubmit={handleSubmit(onSubmit)}>
							<Box mt="3" mb="5">
								<Form.FormField name="token">
									<Flex justify="center" mt="1">
										<InputOTP
											{...register("code", {
												required:
													"The two-factor authentication code is required",
											})}
											type="text"
											inputMode={isEmail ? "text" : "numeric"}
											maxLength={isEmail ? 8 : 6}
											onChange={(value) => setValue("code", value)}
											pattern={
												isEmail
													? REGEXP_ONLY_DIGITS_AND_CHARS
													: REGEXP_ONLY_DIGITS
											}
										>
											{isEmail ? (
												<>
													<InputOTPGroup>
														<InputOTPSlot index={0} />
														<InputOTPSlot index={1} />
														<InputOTPSlot index={2} />
														<InputOTPSlot index={3} />
														<InputOTPSlot index={4} />
														<InputOTPSlot index={5} />
														<InputOTPSlot index={6} />
														<InputOTPSlot index={7} />
													</InputOTPGroup>
												</>
											) : (
												<>
													<InputOTPGroup>
														<InputOTPSlot index={0} />
														<InputOTPSlot index={1} />
														<InputOTPSlot index={2} />
													</InputOTPGroup>
													<InputOTPSeparator />
													<InputOTPGroup>
														<InputOTPSlot index={3} />
														<InputOTPSlot index={4} />
														<InputOTPSlot index={5} />
													</InputOTPGroup>
												</>
											)}
										</InputOTP>
									</Flex>

									<FieldError errors={errors} name="code" />
								</Form.FormField>

								{isEmail && !useBackupCode ? (
									<Flex justify="end" align="center" gap="2" mt="2">
										<Button
											type="button"
											variant="ghost"
											onClick={() => {
												resendMutation.call({
													session_id: session.id,
													scope: "login",
												});
											}}
											loading={resendMutation.isMutating}
										>
											<ArrowClockwise /> Resend code
										</Button>
									</Flex>
								) : null}
							</Box>

							<Button
								type="submit"
								style={{ width: "100%" }}
								loading={isSubmitting}
							>
								Verify
							</Button>
						</Form.Root>

						<Separator
							style={{ width: "100%", flexGrow: 1, marginBlock: "1rem" }}
							orientation="horizontal"
						/>

						{accounts.length > 1 ? (
							<Box>
								<Text color="gray" size="2">
									Or continue using any of the following alternative methods:
								</Text>
								<ScrollArea style={{ maxHeight: "400px" }}>
									<Box>
										<RadioCards.Root mt="3" columns="1" gap="2">
											{accounts.map((account, idx) => {
												if (useBackupCode || account.id !== session.account_id)
													return (
														<RadioCards.Item
															value={account.id}
															key={account.id}
															style={{ width: "100%" }}
															onClick={() => {
																initiationMutation.call({
																	account_id: account.id,
																	prev_session_id: session.id,
																});
															}}
														>
															<Flex
																width="100%"
																gap="3"
																align="center"
																justify="start"
															>
																{account.type === "email" ? (
																	<Mailbox
																		size={20}
																		className="text-[var(--gray-10)]"
																	/>
																) : (
																	<Key
																		size={20}
																		className="text-[var(--gray-10)]"
																	/>
																)}
																<Flex direction="column">
																	<Text size="2" weight="bold">
																		{account.name ?? `Account ${idx + 1}`}
																	</Text>
																	<Text size="1" color="gray">
																		{account.email ?? account.id}
																	</Text>
																</Flex>
															</Flex>
														</RadioCards.Item>
													);
											})}
										</RadioCards.Root>
									</Box>
								</ScrollArea>
							</Box>
						) : null}

						{!useBackupCode ? (
							<Callout.Root mt="4" color="amber" variant="surface">
								<Callout.Icon>
									<Warning />
								</Callout.Icon>
								<Callout.Text size="1">
									If you don't have access to your email or authenticator app,
									click{" "}
									<Link
										to="/auth/mfa/$sessionId"
										params={{ sessionId: session.id }}
										search={{ useBackupCode: true }}
									>
										here
									</Link>{" "}
									to use one of the backup codes provided to you during setup.
								</Callout.Text>
							</Callout.Root>
						) : null}
					</Flex>
				</Skeleton>
			</Card>

			<Link to="/auth/sign-in" style={{ marginTop: "1rem" }}>
				Back to sign-in
			</Link>
		</Flex>
	);
}
