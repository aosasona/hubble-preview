import { Warning } from "@phosphor-icons/react";
import {
	createFileRoute,
	Link,
	type ErrorComponentProps,
} from "@tanstack/react-router";
import { Callout, Flex } from "@radix-ui/themes";
import { ProcedureCallError } from "$/lib/server/bindings";
import client from "$/lib/server";
import * as v from "valibot";
import { extractError } from "$/lib/error";

const searchParamsSchema = v.object({
	useBackupCode: v.optional(v.fallback(v.boolean(), false), false),
});

export const Route = createFileRoute("/auth/mfa/$sessionId")({
	validateSearch: searchParamsSchema,
	loader: ({ params }) => {
		return client.queries.mfaLoadSession(params.sessionId);
	},
	errorComponent: (error) => <ErrorComponent {...error} />,
});

function ErrorComponent(props: ErrorComponentProps) {
	return (
		<Flex
			width="100vw"
			height="100vh"
			direction="column"
			align="center"
			justify="center"
		>
			<Flex direction="column" align="center" gap="3" maxWidth="500px">
				<Callout.Root color="red" role="alert" variant="surface">
					<Callout.Icon>
						<Warning />
					</Callout.Icon>
					<Callout.Text>
						{props.error instanceof ProcedureCallError
							? (extractError(props.error)?.message ?? "An error occurred")
							: import.meta.env.DEV
								? props.error.message
								: "An error occurred"}
					</Callout.Text>
				</Callout.Root>

				<Link to="/auth/sign-in">Back to sign-in</Link>
			</Flex>
		</Flex>
	);
}
