import { Sidebar, X } from "@phosphor-icons/react";
import { Box, Flex, IconButton } from "@radix-ui/themes";
import { AnimatePresence, motion } from "motion/react";
import SidebarContent from "./sidebar-content";
import stores from "$/stores";
import { useSnapshot } from "valtio";

export default function MobileSidebar() {
	const app = useSnapshot(stores.app);

	return (
		<Box
			display={{ initial: "block", md: "none" }}
			position="sticky"
			top="0"
			left="0"
			right="0"
			className="z-[9999] border-b border-b-[var(--gray-3)] bg-background"
		>
			<Flex height="36px" px="3" align="center">
				<IconButton
					variant="ghost"
					color="gray"
					onClick={() => app.toggleMobileSidebar()}
				>
					<Sidebar size={20} />
				</IconButton>
			</Flex>

			<AnimatePresence>
				{app.dialogs.mobileSidebar && (
					<motion.div className="fixed top-0 right-0 bottom-0 left-0 z-[9999] h-screen w-screen">
						<motion.div
							initial={{ opacity: 0, x: "-100%" }}
							animate={{ opacity: 1, x: 0 }}
							exit={{ opacity: 0, x: "-100%" }}
							transition={{ duration: 0.125 }}
							className="h-screen w-[85vw] xs:max-w-[225px] border-r border-r-[var(--gray-2)] bg-background sm:max-w-[300px]"
						>
							<SidebarContent
								header={
									<Flex justify="end">
										<IconButton
											color="gray"
											variant="ghost"
											size="2"
											onClick={() => app.toggleMobileSidebar()}
										>
											<X size={18} />
										</IconButton>
									</Flex>
								}
							/>
						</motion.div>
						<motion.div
							initial={{ opacity: 0 }}
							animate={{ opacity: 0.6 }}
							exit={{ opacity: 0 }}
							transition={{ duration: 0.125 }}
							className="fixed top-0 right-0 bottom-0 left-0 z-[-1] bg-black"
							onClick={() => app.toggleMobileSidebar()}
						/>
					</motion.div>
				)}
			</AnimatePresence>
		</Box>
	);
}
