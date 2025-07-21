import { Flex, Heading, Text } from "@radix-ui/themes";

export default function NotFound() {
	return (
		<Flex
			flexGrow="1"
			minHeight="0"
			direction="column"
			align="center"
			justify="center"
			gap="4"
			p="4"
		>
			<Heading size="8">404 - Not Found</Heading>
			<Text size="2" color="gray">
				Sorry, the page you are looking for does not exist.
			</Text>
		</Flex>
	);
}
