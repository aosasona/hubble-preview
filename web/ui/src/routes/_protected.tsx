import ErrorComponent from "$/components/route/error-component";
import AppLayout from "$/components/layout/app-layout";
import { requireAuth } from "$/lib/auth";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import SplashScreen from "$/components/splash-screen";
import { useUserData } from "$/lib/hooks";
import stores from "$/stores";

export const Route = createFileRoute("/_protected")({
	component: RouteComponent,
	beforeLoad: ({ location }) => {
		// Close the mobile navigation
		stores.app.closeDialog("mobileSidebar");
		return requireAuth(location);
	},
	pendingComponent: SplashScreen,
	errorComponent: ErrorComponent,
});

function RouteComponent() {
	const data = useUserData();

	if (data.isLoading) {
		return <SplashScreen />;
	}

	return (
		<AppLayout>
			<Outlet />
		</AppLayout>
	);
}
