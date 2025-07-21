import { Flex, Spinner, Text } from "@radix-ui/themes";

type Props = {
	text?: string;
};

export default function PageSpinner(props: Props) {
	return (
		<Flex
			flexGrow="1"
			direction="column"
			width="100%"
			height="100%"
			minHeight={{ initial: "90vh", xs: "0" }}
			align="center"
			justify="center"
			gap="2"
		>
			<Spinner size="2" />
			{props.text ? (
				<Text size="1" color="gray" align="center">
					{props.text}
				</Text>
			) : null}
		</Flex>
	);
}
