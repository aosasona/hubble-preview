import { extractError } from "$/lib/error";
import { ArrowLeft, House } from "@phosphor-icons/react";
import { Box, Button, Text, Flex, Heading } from "@radix-ui/themes";
import { useRouter } from "@tanstack/react-router";
import { useMemo } from "react";

type Props = {
	error: Error;
	info?: {
		componentStack: string;
	};
	reset: (() => void) | null;
};

export default function ErrorComponent(props: Props) {
	const router = useRouter();

	const message = useMemo(() => {
		let message = "An error occurred. Please try again later.";

		const err = JSON.parse(JSON.stringify(props.error)) as {
			routerCode: string;
		};
		if (err.routerCode === "VALIDATE_SEARCH") {
			message = "Invalid search query";

			const errors = JSON.parse(props.error.message) as {
				kind: "schema" | unknown;
				expected: string;
				recieved: string;
				path: {
					key: string;
				}[];
			}[];

			if (errors.length > 0 && errors[0].kind === "schema") {
				if (errors[0].kind === "schema") {
					message = "Invalid search query";

					for (const error of errors[0].path) {
						message += `: ${error.key}, expected \`${errors[0].expected}\``;
					}
				}
			}
		} else {
			const extractedError = extractError(props.error);
			if (!extractError) return message;

			if (extractedError?.type === "validation" && extractedError?.errors) {
				const msgs = [];
				for (const [key, validationErrors] of Object.entries(
					extractedError.errors,
				)) {
					for (const validationError of validationErrors) {
						msgs.push(`'${key}': ${validationError}`);
					}

					message = msgs.join(", ");
				}
			} else if (extractedError?.message) {
				message = extractedError.message;
			}
		}

		return message.endsWith(".") ? message : `${message}.`;
	}, [props.error]);

	return (
		<Box width="100%" height="100dvh">
			<Flex
				width="100%"
				height="100%"
				direction="column"
				align="center"
				justify="center"
				gap="2"
			>
				<Heading size="5">An error occurred!</Heading>

				<Flex
					width={{ initial: "95vw", xs: "385px", md: "450px" }}
					align="center"
					justify="center"
					px="2"
					mx="auto"
				>
					<Text size="2" color="gray" wrap="pretty" className="!text-center">
						{message}
					</Text>
				</Flex>

				<Flex gap="2" mt="4">
					{router.history.canGoBack() ? (
						<Button size="1" onClick={() => router.history.back()}>
							<ArrowLeft /> Go back
						</Button>
					) : (
						<Button size="1" onClick={() => router.navigate({ to: "/" })}>
							<House /> Go home
						</Button>
					)}
					{import.meta.env.DEV && props.reset ? (
						<Button size="1" onClick={props.reset} variant="surface">
							Reset
						</Button>
					) : null}
				</Flex>
			</Flex>
		</Box>
	);
}
