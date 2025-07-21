import lightLogo from "$/assets/light-logo-transparent.svg";
import darkLogo from "$/assets/dark-logo-transparent.svg";
import { Box, Flex, Text } from "@radix-ui/themes";

type Props = {
	description?: string;
};

export default function SplashScreen(props: Props) {
	return (
		<Flex
			width="100%"
			height="100vh"
			direction="column"
			gap="3"
			align="center"
			justify="center"
		>
			<Box width="42px" className="aspect-square animate-pulse">
				<img src={lightLogo} alt="Light Logo" className="hidden dark:block" />
				<img src={darkLogo} alt="Dark Logo" className="block dark:hidden" />
			</Box>
			{props.description ? (
				<Box width={{ initial: "95vw", xs: "385px" }} mx="auto">
					<Text size="1" color="gray" align="center">
						{props.description}
					</Text>
				</Box>
			) : null}
		</Flex>
	);
}
