import { Button, Dialog } from "@radix-ui/themes";
import BackupCodes from "./backup-codes";
import { ArrowClockwise } from "@phosphor-icons/react";
import { useState } from "react";
import { useRobinMutation } from "$/lib/hooks";

export default function RegenerateBackupCodes() {
	const [codes, setCodes] = useState<string[]>([]);

	const mutation = useRobinMutation("mfa.regenerate-backup-codes", {
		onSuccess: (data) => setCodes(data),
	});

	return (
		<>
			<Button
				variant="ghost"
				loading={mutation.isMutating}
				onClick={() => mutation.call()}
			>
				<ArrowClockwise />
				Regenerate Backup Codes
			</Button>

			<Dialog.Root
				open={codes.length > 0}
				onOpenChange={(open) => !open && setCodes([])}
			>
				<Dialog.Content width="375px">
					<Dialog.Title size="6">Backup Codes</Dialog.Title>
					<Dialog.Description size="2" color="gray" mb="4">
						Please save these backup codes in a safe place. You can use them to
						sign in if you lose access to your email or authenticator app.
					</Dialog.Description>

					<BackupCodes codes={codes} />
				</Dialog.Content>
			</Dialog.Root>
		</>
	);
}
