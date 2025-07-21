import { useMemo } from "react";
import { extractError } from "$/lib/error";
import { Callout, Flex, Heading } from "@radix-ui/themes";
type ErrorProps = {
	error: Error;
};

export default function ErrorPage(props: ErrorProps) {
	const message = useMemo(() => {
		const err = extractError(props.error);
		if (err) {
			return err.message;
		}

		return "An error occurred. Try refreshing the page.";
	}, [props.error]);

	return (
		<Flex
			direction="column"
			flexGrow="1"
			width="100%"
			height="100%"
			minHeight="0"
			align="center"
			justify="center"
		>
			<Flex direction="column" align="center" width="400px" gap="2">
				<Heading>An error occurred</Heading>
				<Callout.Root color="red" variant="surface">
					<Callout.Text align="center">{message}</Callout.Text>
				</Callout.Root>
			</Flex>
		</Flex>
	);
}
