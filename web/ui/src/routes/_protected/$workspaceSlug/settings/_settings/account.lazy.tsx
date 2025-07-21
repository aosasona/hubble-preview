import PageLayout from "$/components/layout/page-layout";
import { createLazyFileRoute, Link } from "@tanstack/react-router";
import * as Form from "@radix-ui/react-form";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import * as v from "valibot";
import { useForm } from "react-hook-form";
import { valibotResolver } from "@hookform/resolvers/valibot";
import Input from "$/components/form/input";
import {
	Box,
	Button,
	Callout,
	Flex,
	Heading,
	Dialog,
	Separator,
	Text,
} from "@radix-ui/themes";
import { ArrowClockwise, At, Warning } from "@phosphor-icons/react";
import RowInput from "$/components/form/row-input";
import { redactEmail, toTitleCase } from "$/lib/utils";
import { toast } from "sonner";
import { useMemo } from "react";
import {
	InputOTP,
	InputOTPGroup,
	InputOTPSlot,
} from "$/components/form/input-otp";
import { REGEXP_ONLY_DIGITS_AND_CHARS } from "input-otp";
import Show from "$/components/show";
import FieldError from "$/components/form/field-error";
import { useRobinMutation } from "$/lib/hooks";

export const Route = createLazyFileRoute(
	"/_protected/$workspaceSlug/settings/_settings/account",
)({
	component: RouteComponent,
});

const profileSchema = v.object({
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
});

const requestEmailOtpSchema = v.object({
	email: v.pipe(v.string(), v.trim(), v.email("Invalid email address")),
});

const verifyNewEmailSchema = v.object({
	email: v.pipe(v.string(), v.trim(), v.email("Invalid email address")),
	code: v.pipe(v.string(), v.regex(/^[a-zA-Z0-9_]{8}$/, "Invalid code.")),
});

type ProfileSchema = v.InferOutput<typeof profileSchema>;
type RequestEmailOtpSchema = v.InferOutput<typeof requestEmailOtpSchema>;
type VerifyNewEmailSchema = v.InferOutput<typeof verifyNewEmailSchema>;

function RouteComponent() {
	const auth = useSnapshot(stores.auth);
	const params = Route.useParams();

	const {
		register: profileFormRegister,
		handleSubmit: handleProfileFormSubmit,
		formState: { errors: profileFormErrors },
		setValue: setProfileFormValue,
		setError: setProfileFormError,
		...profileForm
	} = useForm<ProfileSchema>({
		resolver: valibotResolver(profileSchema),
		defaultValues: {
			first_name: auth.user?.first_name,
			last_name: auth.user?.last_name,
			username: auth.user?.username,
		},
	});

	const firstName = profileForm.watch("first_name");
	const lastName = profileForm.watch("last_name");
	const username = profileForm.watch("username");
	const hasUpdatedProfile = useMemo(() => {
		return (
			firstName?.toLowerCase() !== auth.user?.first_name?.toLowerCase() ||
			lastName?.toLowerCase() !== auth.user?.last_name?.toLowerCase() ||
			username?.toLowerCase() !== auth.user?.username?.toLowerCase()
		);
	}, [
		auth.user?.first_name,
		auth.user?.last_name,
		auth.user?.username,
		firstName,
		lastName,
		username,
	]);

	const requestEmailOtpForm = useForm<RequestEmailOtpSchema>({
		resolver: valibotResolver(requestEmailOtpSchema),
		defaultValues: {
			email: auth.user?.email,
		},
	});
	const requestEmailOtpFormValue = requestEmailOtpForm.watch("email");

	const verifyNewEmailForm = useForm<VerifyNewEmailSchema>({
		resolver: valibotResolver(verifyNewEmailSchema),
	});
	const newEmail = verifyNewEmailForm.watch("email");

	const saveProfileMutation = useRobinMutation("user.save-profile", {
		invalidates: ["me"],
		onSuccess: (data) => {
			setProfileFormValue("first_name", data.first_name);
			setProfileFormValue("last_name", data.last_name);
			setProfileFormValue("username", data.username);

			toast.success("Profile updated successfully");
		},
		setFormError: setProfileFormError,
	});

	const requestEmailMutation = useRobinMutation("user.request-email-change", {
		onSuccess: (data) => {
			verifyNewEmailForm.setValue("email", data.email);
			toast.success(`We have sent an email to ${redactEmail(data.email)}`);
		},
		setFormError: requestEmailOtpForm.setError,
	});

	const verifyEmailMutation = useRobinMutation("user.verify-email-change", {
		invalidates: ["me"],
		onSuccess: () => {
			verifyNewEmailForm.reset();
			toast.success("Email changed successfully");
		},
		setFormError: verifyNewEmailForm.setError,
	});

	return (
		<PageLayout heading="Account" header={{ parent: "settings" }} showHeader>
			<Flex direction="column" mt="6" gap="6">
				<Form.Root
					onSubmit={handleProfileFormSubmit(async (data) => {
						await saveProfileMutation.call(data);
					})}
				>
					<Box>
						<Heading size="3" mb="1">
							Personal Details
						</Heading>
						<Text size="2" color="gray">
							Manage your personal information.
						</Text>

						<Flex direction="column" gap="5" mt="4">
							<Flex direction="column" gap="3">
								<RowInput
									register={profileFormRegister}
									name="first_name"
									label="First name"
									errors={profileFormErrors}
								/>

								<RowInput
									register={profileFormRegister}
									name="last_name"
									label="Last name"
									errors={profileFormErrors}
								/>

								<RowInput
									register={profileFormRegister}
									name="username"
									label="Username"
									errors={profileFormErrors}
									LeftSideComponent={() => <At size={16} />}
								/>
							</Flex>

							<Flex justify="end">
								<Button
									type="submit"
									loading={saveProfileMutation.isMutating}
									disabled={!hasUpdatedProfile}
								>
									Save changes
								</Button>
							</Flex>
						</Flex>
					</Box>
				</Form.Root>

				<Separator orientation="horizontal" style={{ width: "100%" }} />

				<Box>
					<Heading size="3">E-mail Address</Heading>

					<Flex
						direction="column"
						align="start"
						gap={{ initial: "2", md: "4" }}
						mt="2"
					>
						<Box>
							<Text color="gray">
								We will send you a One-Time Password (OTP) to verify your new
								email address.
							</Text>
						</Box>

						<Form.Root
							onSubmit={requestEmailOtpForm.handleSubmit(
								async (d) => await requestEmailMutation.call(d),
							)}
							style={{ width: "100%" }}
						>
							<Flex
								direction={{ initial: "column", md: "row" }}
								align="start"
								gap="3"
							>
								<Input
									register={requestEmailOtpForm.register}
									name="email"
									label="E-mail address"
									errors={requestEmailOtpForm.formState.errors}
									LeftSideComponent={() => <At size={16} />}
									hideLabel
								/>
								<Button
									type="submit"
									loading={
										requestEmailOtpForm.formState.isSubmitting ||
										requestEmailMutation.isMutating
									}
									disabled={
										requestEmailOtpFormValue?.toLowerCase() ===
										auth.user?.email.toLowerCase()
									}
								>
									<Warning /> Change Email
								</Button>
							</Flex>
						</Form.Root>
						<Callout.Root color="amber" size="1" variant="surface">
							<Callout.Icon>
								<Warning />
							</Callout.Icon>
							<Callout.Text>
								If you no longer want to use your current email for multi-factor
								authentication (if enabled), you have to disable it in the{" "}
								<Link
									to="/$workspaceSlug/settings/security"
									params={{ workspaceSlug: params.workspaceSlug }}
								>
									security
								</Link>{" "}
								settings.
							</Callout.Text>
						</Callout.Root>
					</Flex>
				</Box>
			</Flex>

			<Show when={!!newEmail}>
				<Dialog.Root
					open={!!newEmail}
					onOpenChange={(open) => {
						if (!open) {
							verifyNewEmailForm.reset();
						}
					}}
				>
					<Dialog.Content maxWidth="375px" style={{ overflowX: "hidden" }}>
						<Dialog.Title>Verify New Email</Dialog.Title>
						<Dialog.Description color="gray" size="2">
							Enter the One-Time Password (OTP) sent to your new email address.
						</Dialog.Description>
						<Form.Root
							onSubmit={verifyNewEmailForm.handleSubmit((d) => {
								verifyEmailMutation.call(d);
							})}
						>
							<Box mt="4" mb="5">
								<Form.FormField name="code">
									<Flex justify="center" align="center" mt="1">
										<InputOTP
											{...verifyNewEmailForm.register("code", {
												required: "The One-Time Password is required",
											})}
											type="text"
											inputMode="text"
											maxLength={8}
											onChange={(value) =>
												verifyNewEmailForm.setValue("code", value)
											}
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
									<FieldError
										errors={verifyNewEmailForm.formState.errors}
										name="code"
									/>

									<Flex justify="end" align="center" gap="2" mt="2">
										<Button
											type="button"
											variant="ghost"
											size="1"
											onClick={async () =>
												await requestEmailMutation.call({ email: newEmail })
											}
											loading={requestEmailMutation.isMutating}
										>
											<ArrowClockwise /> Resend code
										</Button>
									</Flex>
								</Form.FormField>
							</Box>

							<Button
								type="submit"
								style={{ width: "100%" }}
								loading={verifyNewEmailForm.formState.isSubmitting}
							>
								Verify Email
							</Button>
						</Form.Root>
					</Dialog.Content>
				</Dialog.Root>
			</Show>
		</PageLayout>
	);
}
