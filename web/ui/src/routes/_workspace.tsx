import ErrorComponent from "$/components/route/error-component";
import { requireAuth } from "$/lib/auth";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import SplashScreen from "$/components/splash-screen";
import { useUserData } from "$/lib/hooks";

export const Route = createFileRoute("/_workspace")({
	component: RouteComponent,
	beforeLoad: ({ location }) => {
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

	return <Outlet />;
}
