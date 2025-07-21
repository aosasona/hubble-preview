import Logo from "$/components/logo";
import { GithubLogo } from "@phosphor-icons/react";
import { Box, Button, Flex, Heading, Text } from "@radix-ui/themes";
import { createLazyFileRoute, Link } from "@tanstack/react-router";

export const Route = createLazyFileRoute("/landing")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<Box>
			<Flex
				width="100vw"
				px={{ initial: "4", xs: "5" }}
				py={{ initial: "4", md: "3" }}
				align="center"
				className="sticky top-0 z-50 bg-[var(--gray-2)]/40 border-b border-[var(--gray-3)]"
			>
				<Flex align="center" gap="2">
					<Logo variant="plain" size="md" />
					<Heading size="5" weight="bold">
						Hubble
					</Heading>
				</Flex>

				<Flex ml="auto" align="center" gap={{ initial: "2", xs: "3", xl: "4" }}>
					<Button size="2" variant="surface" radius="full" asChild>
						<Link to="/auth/sign-in">Sign in</Link>
					</Button>
				</Flex>
			</Flex>

			<Flex direction="column" align="center" width="100%">
				<Flex
					width={{ initial: "100vw", xs: "50rem", lg: "75rem" }}
					direction="column"
					align="center"
					justify="center"
					gap="3"
					className="text-center py-16"
				>
					<Heading className="!text-4xl lg:!text-8xl leading-tight">
						All your knowledge, intelligently organized.
					</Heading>

					<Text size={{ initial: "3", xs: "4" }} color="gray" mt="2">
						A unified, secure and extensible knowledge-base software for teams
						and individuals
					</Text>

					<Flex
						direction={{ initial: "column", md: "row" }}
						align="center"
						gap={{ initial: "7", md: "5" }}
						mt="5"
					>
						<Button size="4" variant="solid" radius="full" asChild>
							<Link to="/auth/sign-in" className="unstyled">
								Get Started
							</Link>
						</Button>

						<Button size="4" variant="ghost" radius="full" asChild>
							<a
								href="https://github.com/aosasona/hubble"
								target="blank"
								className="unstyled"
							>
								<GithubLogo /> Star on GitHub
							</a>
						</Button>
					</Flex>
				</Flex>
			</Flex>
		</Box>
	);
}
