import { Button, Flex, Card, Grid, Text } from "@radix-ui/themes";
import { useState } from "react";
import { toast } from "sonner";
import { Check, Copy, DownloadSimple } from "@phosphor-icons/react";

type Props = {
	codes: string[];
};

export default function BackupCodes(props: Props) {
	const [copied, setCopied] = useState(false);

	function copyBackupCodes() {
		if (!props.codes) return;
		navigator.clipboard.writeText(props.codes?.join("\n"));
		toast.success("Backup codes copied to clipboard.");
		setCopied(true);
	}

	function downloadCodes() {
		if (!props.codes) return;

		const blob = new Blob([props.codes.join("\n")], {
			type: "text/plain;charset=utf-8",
		});
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = "backup-codes.txt";
		a.click();
		URL.revokeObjectURL(url);
		toast.success("Backup codes downloaded.");
	}

	return (
		<Flex direction="column" gap="4">
			<Card>
				<Grid columns="2" gap="2" align="center" justify="center" p="2">
					{props.codes.map((code, idx) => (
						<Flex key={code} gap="2" align="end">
							<Text size="2" weight="medium" color="gray">
								{idx + 1}.
							</Text>
							<Text size="3" weight="medium">
								{code}
							</Text>
						</Flex>
					))}
				</Grid>
			</Card>

			<Flex align="center" justify="end" gap="4">
				<Button variant="ghost" onClick={() => downloadCodes()}>
					<DownloadSimple /> Download
				</Button>
				<Button variant="surface" onClick={() => copyBackupCodes()}>
					{copied ? <Check /> : <Copy />}
					{copied ? "Copied" : "Copy"}
				</Button>
			</Flex>
		</Flex>
	);
}
