import { Box, Flex } from "@radix-ui/themes";
import type { ReactNode } from "react";
import SidebarUser from "./user";
import WorkspaceSwitcher from "./workspace-switcher";
import NewCollectionDialog from "$/components/collections/new-collection-dialog";
import { useSnapshot } from "valtio";
import stores from "$/stores";
import SidebarItems from "./sidebar-items";

type Props = {
	header?: ReactNode;
};

export default function SidebarContent(props: Props) {
	const app = useSnapshot(stores.app);

	return (
		<Flex width="100%" height="100%" direction="column" gap="2" p="3">
			{props?.header ? <Box>{props.header}</Box> : null}

			<Flex direction="column" minHeight="0" width="100%" height="100%" gap="2">
				<WorkspaceSwitcher />

				<SidebarItems />
			</Flex>

			<Box mt="auto">
				<SidebarUser />
			</Box>

			<NewCollectionDialog
				open={app.dialogs.createCollection}
				onOpen={() => app.openDialog("createCollection")}
				onClose={() => app.closeDialog("createCollection")}
			/>
		</Flex>
	);
}
