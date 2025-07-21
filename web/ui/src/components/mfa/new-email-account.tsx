import { Button, Callout, Dialog, Flex, Box } from "@radix-ui/themes";
import { useState } from "react";
import { Fragment } from "react/jsx-runtime";
import { AnimatePresence } from "motion/react";
import {
	ArrowClockwise,
	ArrowLeft,
	ArrowRight,
	Check,
	Info,
} from "@phosphor-icons/react";
import Input from "../form/input";
import * as Form from "@radix-ui/react-form";
import { useForm } from "react-hook-form";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";
import type { MutationResult } from "$/lib/server";
import {
	InputOTP,
	InputOTPGroup,
	InputOTPSeparator,
	InputOTPSlot,
} from "../form/input-otp";
import { REGEXP_ONLY_DIGITS_AND_CHARS } from "input-otp";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import BackupCodes from "./backup-codes";
import ProgressIndicator from "../form/progress-indicator";
import AnimatedDialogView from "../animated-dialog-view";
import { useRobinMutation } from "$/lib/hooks";

enum Stage {
	EnterEmail = 1,
	VerifyEmail = 2,
	Done = 3,
}

type Props = {
	backToSelectionView: () => void;
};

const emailFormSchema = v.object({
	email: v.pipe(
		v.string(),
		v.email("A valid email address is required"),
		v.nonEmpty("An email address is required"),
	),
});

const verificationFormSchema = v.object({
	email: v.pipe(v.string(), v.nonEmpty("Email is required"), v.email()),
	token: v.pipe(
		v.string(),
		v.nonEmpty("The One-Time Password is required."),
		v.regex(/^[a-zA-Z0-9_]{8}$/, "Invalid token."),
	),
});

type EmailFormSchema = v.InferOutput<typeof emailFormSchema>;
type VerificationFormSchema = v.InferOutput<typeof verificationFormSchema>;

export default function NewEmailAccount(props: Props) {
	const auth = useSnapshot(stores.auth);

	const [response, setResponse] =
		useState<MutationResult<"mfa.create-email-account"> | null>(null);
	const [stage, setStage] = useState<Stage>(Stage.EnterEmail);
	const [backupCodes, setBackupCodes] = useState<string[] | null>(null);

	const emailForm = useForm<EmailFormSchema>({
		resolver: valibotResolver(emailFormSchema),
	});

	const verificationForm = useForm<VerificationFormSchema>({
		resolver: valibotResolver(verificationFormSchema),
	});

	const createEmailMfaAccount = useRobinMutation("mfa.create-email-account", {
		invalidates: ["mfa.state"],
		onSuccess: (data) => {
			setResponse(data);
			setStage(Stage.VerifyEmail);
		},
		setFormError: emailForm.setError,
	});

	const resendMfaEmailVerification = useRobinMutation("mfa.resend-email", {
		setFormError: emailForm.setError,
	});

	const activateEmailMfaAccount = useRobinMutation(
		"mfa.activate-email-account",
		{
			onSuccess: (data) => {
				setStage(Stage.Done);
				setBackupCodes(data.backup_codes);
			},
			invalidates: ["mfa.state"],
			setFormError: verificationForm.setError,
		},
	);

	async function onCreateEmailMfaAccount(data: EmailFormSchema) {
		await createEmailMfaAccount.call(data);
	}

	async function onVerificationSubmit(data: VerificationFormSchema) {
		await activateEmailMfaAccount.call({
			token: data.token,
			session_id: response?.session_id ?? "",
		});
	}

	return (
		<Fragment>
			<ProgressIndicator total={3} completed={stage} inDialog />

			<AnimatePresence>
				<AnimatedDialogView
					key={Stage.EnterEmail}
					visible={stage === Stage.EnterEmail}
				>
					<Dialog.Title size="6">Add an e-Mail address</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						Enter an email address you would like to use for multi-factor
						authentication.
					</Dialog.Description>

					<Form.Root onSubmit={emailForm.handleSubmit(onCreateEmailMfaAccount)}>
						<Flex direction="column" gap="2" mt="5">
							<Input
								register={emailForm.register}
								name="email"
								type="email"
								label="E-Mail address"
								errors={emailForm.formState.errors}
								required
							/>
							<Button
								size="1"
								variant="ghost"
								style={{ marginLeft: "auto" }}
								type="button"
								onClick={() =>
									emailForm.setValue("email", auth.user?.email ?? "")
								}
							>
								Fill with current email
							</Button>
							<Callout.Root color="orange" variant="surface">
								<Callout.Icon>
									<Info size={16} />
								</Callout.Icon>
								<Callout.Text size="1">
									This can be different from the email address linked to your
									account.
								</Callout.Text>
							</Callout.Root>
						</Flex>

						<Flex direction="row-reverse" justify="start" gap="3" mt="6" py="1">
							<Button type="submit" loading={emailForm.formState.isSubmitting}>
								Continue
								<ArrowRight />
							</Button>
							<Button
								type="button"
								variant="surface"
								color="gray"
								onClick={props.backToSelectionView}
							>
								<ArrowLeft />
								Previous
							</Button>
						</Flex>
					</Form.Root>
				</AnimatedDialogView>

				<AnimatedDialogView
					key={Stage.VerifyEmail}
					visible={stage === Stage.VerifyEmail}
				>
					<Dialog.Title size="6">Verify e-Mail address</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						Enter the code we sent to your email address to verify it.
					</Dialog.Description>

					<Form.Root
						onSubmit={verificationForm.handleSubmit(onVerificationSubmit)}
					>
						<input
							{...verificationForm.register("email")}
							type="hidden"
							value={response?.email}
						/>
						<Flex justify="center" my="5">
							<Flex direction="column" gap="2">
								<InputOTP
									{...verificationForm.register("token", {
										required: "The One-Time Password is required",
									})}
									type="text"
									inputMode="text"
									maxLength={8}
									onChange={(value) =>
										verificationForm.setValue("token", value)
									}
									pattern={REGEXP_ONLY_DIGITS_AND_CHARS}
								>
									<InputOTPGroup>
										<InputOTPSlot index={0} />
										<InputOTPSlot index={1} />
										<InputOTPSlot index={2} />
										<InputOTPSlot index={3} />
									</InputOTPGroup>

									<InputOTPSeparator />

									<InputOTPGroup>
										<InputOTPSlot index={4} />
										<InputOTPSlot index={5} />
										<InputOTPSlot index={6} />
										<InputOTPSlot index={7} />
									</InputOTPGroup>
								</InputOTP>

								<Button
									type="button"
									variant="ghost"
									onClick={() =>
										resendMfaEmailVerification.call({
											session_id: response?.session_id ?? "",
											scope: "setup",
										})
									}
									loading={resendMfaEmailVerification.isMutating}
									style={{ marginLeft: "auto" }}
								>
									<ArrowClockwise />
									Resend
								</Button>
							</Flex>
						</Flex>

						<Flex direction="row-reverse" justify="start" gap="3" py="1">
							<Button
								type="submit"
								loading={verificationForm.formState.isSubmitting}
							>
								Verify
								<Check />
							</Button>
							<Button
								type="button"
								variant="surface"
								color="gray"
								onClick={() => setStage(Stage.EnterEmail)}
							>
								<ArrowLeft />
								Previous
							</Button>
						</Flex>
					</Form.Root>
				</AnimatedDialogView>

				<AnimatedDialogView key={Stage.Done} visible={stage === Stage.Done}>
					<Dialog.Title size="6">Account setup complete</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						{backupCodes?.length
							? "Please save these backup codes in a safe place. You can use them to sign in if you lose access to your email."
							: "Your email address has been verified and added as a multi-factor authentication account."}
					</Dialog.Description>

					<Box my="5">
						{backupCodes?.length ? <BackupCodes codes={backupCodes} /> : null}
					</Box>

					<Dialog.Close>
						<Button style={{ width: "100%" }}>Done</Button>
					</Dialog.Close>
				</AnimatedDialogView>
			</AnimatePresence>
		</Fragment>
	);
}
