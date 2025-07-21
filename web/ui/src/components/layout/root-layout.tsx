import useShortcut from "$/lib/hooks/use-shortcut";
import stores from "$/stores";
import { Box } from "@radix-ui/themes";
import type { ReactNode } from "react";
import { useSnapshot } from "valtio";

type Props = {
	children: ReactNode;
};

export default function RootLayout({ children }: Props) {
	const app = useSnapshot(stores.app);

	useShortcut(["alt+d", "ctrl+shift+d"], app.toggleColorScheme);
	useShortcut(["alt+0", "ctrl+shift+0"], app.switchToSystemColorScheme);

	return <Box>{children}</Box>;
}
