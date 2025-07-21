import { Box, Flex, Heading } from "@radix-ui/themes";
import MobileSidebar from "./sidebar/mobile-sidebar";
import DesktopSidebar from "./sidebar/desktop-sidebar";
import { useSnapshot } from "valtio";
import Show from "../show";
import { CaretRight } from "@phosphor-icons/react";
import { AnimatePresence } from "motion/react";
import { motion } from "motion/react";
import stores from "$/stores";
import useProtectedKeyBindings from "$/lib/hooks/use-protected-key-bindings";
import { Link } from "@tanstack/react-router";

type Props = { children: React.ReactNode };

export default function AppLayout({ children }: Props) {
	const layout = useSnapshot(stores.layout);
	useProtectedKeyBindings();

	return (
		<Flex
			direction={{ initial: "column", md: "row" }}
			height={{ md: "100vh" }}
			minHeight={{ md: "100vh" }}
			gap={{ initial: "0", md: "2" }}
			className="md:bg-[var(--gray-2)]/40"
		>
			{/* Desktop sidebar */}
			<DesktopSidebar />

			{/* Mobile sidebar */}
			<MobileSidebar />

			<Flex
				direction="column"
				flexGrow="1"
				width="100%"
				minHeight={{ initial: "0", md: "100vh" }}
				height={{ initial: "100%", md: "100vh" }}
				maxHeight={{ md: "100vh" }}
				px={{ initial: "0", md: "2" }}
				pb={{ initial: "0", md: "2" }}
				pl="0"
				m="0"
			>
				<Flex
					display={{ initial: "none", md: "flex" }}
					width="100%"
					align="center"
					justify="center"
					pt={{ initial: "0", md: "2" }}
				>
					<Show when={layout.hasHeaderItems() && !layout.fullScreen}>
						<AnimatePresence>
							<motion.div
								initial={{ opacity: 0, scale: 0.9 }}
								animate={{ opacity: 1, scale: 1 }}
								exit={{ opacity: 0, scale: 0.9 }}
								transition={{ duration: 0.1, type: "tween" }}
								className="flex select-none items-center justify-center gap-1"
							>
								{layout.headerItems.map((item, idx) => (
									<Box key={`${item.title}-${idx}`} pb="2">
										{"component" in item ? (
											<item.component />
										) : (
											<Flex key={item.title} gap="1" align="center">
												{item.icon ? (
													<item.icon
														size={16}
														style={{
															color: `var(--${item.color ?? "gray"}-${idx === 0 && !item.color ? "contrast" : "12"})`,
														}}
													/>
												) : null}

												<motion.div
													initial={{ opacity: 0, y: 6 }}
													animate={{ opacity: 1, y: 0 }}
													exit={{ opacity: 0, y: 6 }}
													transition={{ duration: 0.1, type: "tween" }}
												>
													{item.url ? (
														<Link to={item.url} className="unstyled">
															<Heading
																size="1"
																weight="medium"
																color="gray"
																highContrast
															>
																{item.title}
															</Heading>
														</Link>
													) : (
														<Heading
															size="1"
															weight="medium"
															color="gray"
															highContrast
														>
															{item.title}
														</Heading>
													)}
												</motion.div>

												{idx !== layout.headerItems.length - 1 ? (
													<motion.div
														initial={{ opacity: 0, width: 0 }}
														animate={{ opacity: 1, width: "auto" }}
														exit={{ opacity: 0, width: 0 }}
														transition={{ duration: 0.1, type: "tween" }}
													>
														<CaretRight
															size={10}
															style={{ color: "var(--gray-12)" }}
														/>
													</motion.div>
												) : null}
											</Flex>
										)}
									</Box>
								))}
							</motion.div>
						</AnimatePresence>
					</Show>
				</Flex>
				<Box
					position="relative"
					width="100%"
					height="100%"
					minHeight="0"
					className="w-full flex-1 bg-background md:rounded-[var(--radius-1)] md:border md:border-[var(--gray-3)]"
				>
					{children}
				</Box>
			</Flex>
		</Flex>
	);
}
