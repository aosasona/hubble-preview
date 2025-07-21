import { Outlet, createRootRoute } from "@tanstack/react-router";
import RootLayout from "$/components/layout/root-layout";

export const Route = createRootRoute({
	component: () => (
		<RootLayout>
			<Outlet />
		</RootLayout>
	),
});
