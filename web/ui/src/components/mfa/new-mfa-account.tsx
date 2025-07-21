import {
	Box,
	Button,
	Dialog,
	Flex,
	Heading,
	RadioCards,
	Text,
} from "@radix-ui/themes";
import type { ReactNode } from "@tanstack/react-router";
import { useState } from "react";
import Show from "../show";
import { At, Key } from "@phosphor-icons/react";
import NewEmailAccount from "./new-email-account";
import NewTotpAccount from "./new-totp-account";
import { AnimatePresence, motion } from "motion/react";

type Props = {
	Trigger: ReactNode;
};

enum MfaType {
	Authenticator = "authenticator",
	Email = "email",
}

enum Stage {
	SelectType = 0,
	Authenticator = 1,
	Email = 2,
}

export default function NewMfaAccount(props: Props) {
	const [open, setOpen] = useState(false);
	const [mfaType, setMfaType] = useState<MfaType>(MfaType.Email);
	const [stage, setStage] = useState<Stage>(Stage.SelectType);

	function handleContinue() {
		if (mfaType === null) return;
		setStage(mfaType === MfaType.Email ? Stage.Email : Stage.Authenticator);
	}

	function handleDialogOpen(open: boolean) {
		if (!open) {
			setMfaType(MfaType.Email);
			setStage(Stage.SelectType);
		}
		setOpen(open);
	}

	return (
		<Dialog.Root open={open} onOpenChange={handleDialogOpen}>
			<Dialog.Trigger>{props.Trigger}</Dialog.Trigger>
			<Dialog.Content maxWidth="400px" className="!overflow-x-hidden">
				<Show when={mfaType === null || stage === Stage.SelectType}>
					<AnimatePresence>
						<motion.div
							initial={{ x: -100 }}
							animate={{ x: 0 }}
							exit={{ x: 100 }}
							transition={{ type: "tween", duration: 0.125 }}
						>
							<Dialog.Title size="6">Account type</Dialog.Title>
							<Dialog.Description size="2" color="gray">
								Would you like to use an authenticator app or your email
								address?
							</Dialog.Description>

							<RadioCards.Root
								columns="1"
								mt="4"
								mb="6"
								defaultValue={mfaType}
								onValueChange={(value) => setMfaType(value as MfaType)}
							>
								<RadioCards.Item value={MfaType.Email}>
									<Flex width="100%" align="center" gap="3">
										<Box p="1">
											<At size={20} />
										</Box>
										<Flex direction="column" gap="1">
											<Heading size="2">Email</Heading>
											<Text size="1" color="gray">
												Confirm sign-in attempts via email.
											</Text>
										</Flex>
									</Flex>
								</RadioCards.Item>

								<RadioCards.Item value={MfaType.Authenticator}>
									<Flex width="100%" align="center" gap="3">
										<Box p="1">
											<Key size={20} />
										</Box>
										<Flex direction="column" gap="1">
											<Heading size="2">Authenticator</Heading>
											<Text size="1" color="gray">
												Use an authenticator app to generate a code for sign-in
												attempts.
											</Text>
										</Flex>
									</Flex>
								</RadioCards.Item>
							</RadioCards.Root>

							<Flex justify="end" gap="3">
								<Dialog.Close>
									<Button variant="surface" color="gray">
										Cancel
									</Button>
								</Dialog.Close>
								<Button onClick={handleContinue} disabled={mfaType === null}>
									Continue
								</Button>
							</Flex>
						</motion.div>
					</AnimatePresence>
				</Show>

				<Show when={mfaType === MfaType.Email && stage === Stage.Email}>
					<NewEmailAccount
						backToSelectionView={() => setStage(Stage.SelectType)}
					/>
				</Show>

				<Show
					when={
						mfaType === MfaType.Authenticator && stage === Stage.Authenticator
					}
				>
					<NewTotpAccount
						backToSelectionView={() => setStage(Stage.SelectType)}
					/>
				</Show>
			</Dialog.Content>
		</Dialog.Root>
	);
}
