import { useState } from "react";
import { Fragment } from "react";
import ProgressIndicator from "../form/progress-indicator";
import * as Form from "@radix-ui/react-form";
import * as v from "valibot";
import { valibotResolver } from "@hookform/resolvers/valibot";
import { AnimatePresence } from "motion/react";
import AnimatedDialogView from "../animated-dialog-view";
import { Box, Button, Code, Dialog, Flex, Text } from "@radix-ui/themes";
import { useForm } from "react-hook-form";
import Input from "../form/input";
import { ArrowLeft, ArrowRight, Check } from "@phosphor-icons/react";
import {
	InputOTP,
	InputOTPGroup,
	InputOTPSeparator,
	InputOTPSlot,
} from "../form/input-otp";
import { REGEXP_ONLY_DIGITS } from "input-otp";
import BackupCodes from "./backup-codes";
import { useRobinMutation } from "$/lib/hooks";

enum Stage {
	ProvideAccountName = 1,
	ShowKeyAndQRCode = 2,
	EnterVerificationCode = 3,
	Done = 4,
}

type Props = {
	backToSelectionView: () => void;
};

const nameSchema = v.object({
	name: v.pipe(
		v.string(),
		v.nonEmpty("An account name is required"),
		v.minLength(2, "Account name must be at least 2 characters"),
		v.maxLength(32, "Account name cannot be longer than 32 characters"),
		v.regex(
			/^[a-zA-Z0-9_ ]+$/,
			"Only letters, numbers, spaces, and underscores are allowed",
		),
	),
});

const verificationFormSchema = v.object({
	session_id: v.pipe(
		v.string(),
		v.nonEmpty("Session ID is required"),
		v.uuid(),
	),
	code: v.pipe(
		v.string(),
		v.nonEmpty("The One-Time Password is required"),
		v.regex(/^[0-9]{6}$/, "Invalid code"),
	),
});

type NameFormSchema = v.InferOutput<typeof nameSchema>;
type VerificationFormSchema = v.InferOutput<typeof verificationFormSchema>;

export default function NewTotpAccount(props: Props) {
	const [stage, setStage] = useState<Stage>(Stage.ProvideAccountName);

	const nameForm = useForm<NameFormSchema>({
		resolver: valibotResolver(nameSchema),
	});

	const verificationForm = useForm<VerificationFormSchema>({
		resolver: valibotResolver(verificationFormSchema),
	});

	const startSessionMutation = useRobinMutation("mfa.start-totp-enrollment", {
		onSuccess: () => setStage(Stage.ShowKeyAndQRCode),
		setFormError: nameForm.setError,
	});

	const completeSessionMutation = useRobinMutation(
		"mfa.complete-totp-enrollment",
		{
			onSuccess: () => {
				setStage(Stage.Done);
			},
			invalidates: ["mfa.state"],
		},
	);

	async function handleNameSubmit(data: NameFormSchema) {
		await startSessionMutation.call(data.name);
	}

	async function handleVerificationSubmit(data: VerificationFormSchema) {
		await completeSessionMutation.call(data);
	}

	return (
		<Fragment>
			<ProgressIndicator total={4} completed={stage} inDialog />

			<AnimatePresence>
				{/* MARK: Provide account name */}
				<AnimatedDialogView
					key={Stage.ProvideAccountName}
					visible={stage === Stage.ProvideAccountName}
				>
					<Dialog.Title size="6">Details</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						Provide an alias for this account to help you identify it later (for
						example; "Ente" or "Work").
					</Dialog.Description>

					<Form.Root
						onSubmit={nameForm.handleSubmit(handleNameSubmit)}
						style={{ marginTop: "var(--space-4)" }}
					>
						<Input
							register={nameForm.register}
							name="name"
							type="text"
							label="Name"
							errors={nameForm.formState.errors}
							textFieldProps={{ autoFocus: true }}
							required
						/>

						<Flex direction="row-reverse" justify="start" gap="3" py="1" mt="4">
							<Button
								type="submit"
								loading={startSessionMutation.isMutating}
								disabled={!nameForm.formState.isValid}
							>
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

				{/* MARK: Show QR Code and key */}
				<AnimatedDialogView
					key={Stage.ShowKeyAndQRCode}
					visible={stage === Stage.ShowKeyAndQRCode}
				>
					<Dialog.Title size="6">Setup Authenticator</Dialog.Title>
					<Dialog.Description size="2" color="gray" align="left">
						Scan the QR code below with your authenticator app to set up this
						account.
					</Dialog.Description>

					<Flex
						direction="column"
						justify="center"
						align="center"
						gap="5"
						mt="4"
					>
						<img
							src={startSessionMutation.data?.image}
							alt="QR Code"
							className="w-full md:size-52"
						/>

						<Text size="2" color="gray" align="left">
							Alternatively, if you are unable to scan the QR code, you can type
							in the following key into your authenticator app:
						</Text>

						<Code size="4" variant="outline">
							{startSessionMutation.data?.secret}
						</Code>
					</Flex>
					<Flex direction="row-reverse" justify="start" gap="3" mt="6" py="1">
						<Button
							type="button"
							onClick={() => setStage(Stage.EnterVerificationCode)}
						>
							Continue
							<ArrowRight />
						</Button>
						<Button
							type="button"
							variant="surface"
							color="gray"
							onClick={() => setStage(Stage.ProvideAccountName)}
						>
							<ArrowLeft />
							Previous
						</Button>
					</Flex>
				</AnimatedDialogView>

				{/* MARK: Enter verification code */}
				<AnimatedDialogView
					key={Stage.EnterVerificationCode}
					visible={stage === Stage.EnterVerificationCode}
				>
					<Dialog.Title size="6">Verify Authenticator</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						Ener the 6-digit code from your authenticator app to complete the
						setup.
					</Dialog.Description>

					<Form.Root
						onSubmit={verificationForm.handleSubmit(handleVerificationSubmit)}
					>
						<input
							{...verificationForm.register("session_id")}
							type="hidden"
							value={startSessionMutation.data?.session_id}
						/>
						<Flex justify="center" my="5">
							<InputOTP
								{...verificationForm.register("code", {
									required: "The One-Time Password is required",
								})}
								type="text"
								inputMode="numeric"
								maxLength={6}
								onChange={(value) => verificationForm.setValue("code", value)}
								pattern={REGEXP_ONLY_DIGITS}
							>
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
							</InputOTP>
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
								onClick={() => setStage(Stage.ShowKeyAndQRCode)}
							>
								<ArrowLeft />
								Previous
							</Button>
						</Flex>
					</Form.Root>
				</AnimatedDialogView>

				{/* MARK: Done */}
				<AnimatedDialogView key={Stage.Done} visible={stage === Stage.Done}>
					<Dialog.Title size="6">Account setup complete</Dialog.Title>
					<Dialog.Description size="2" color="gray">
						{completeSessionMutation.data?.backup_codes?.length
							? "Please save these backup codes in a safe place. You can use them to sign in if you lose access to your email."
							: "Your authenticator app has been set up successfully and is ready to use."}
					</Dialog.Description>

					<Box my="5">
						{completeSessionMutation.data?.backup_codes?.length ? (
							<BackupCodes codes={completeSessionMutation.data?.backup_codes} />
						) : null}
					</Box>

					<Dialog.Close>
						<Button style={{ width: "100%" }}>Done</Button>
					</Dialog.Close>
				</AnimatedDialogView>
			</AnimatePresence>
		</Fragment>
	);
}
