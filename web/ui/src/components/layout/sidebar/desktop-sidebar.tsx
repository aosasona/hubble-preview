import { Box } from "@radix-ui/themes";
import SidebarContent from "./sidebar-content";

export default function DesktopSidebar() {
	return (
		<Box
			display={{ initial: "none", md: "block" }}
			width={{ sm: "200px", lg: "230px", xl: "250px" }}
			minWidth={{ sm: "200px", lg: "230px", xl: "250px" }}
		>
			<SidebarContent />
		</Box>
	);
}
