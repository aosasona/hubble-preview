import { RouterProvider, createRouter } from "@tanstack/react-router";
import { Theme } from "@radix-ui/themes";
import { StrictMode, useEffect, useMemo } from "react";

// Import the generated route tree
import { routeTree } from "./routeTree.gen";
import { Toaster } from "sonner";

import { QueryClientProvider } from "@tanstack/react-query";
import { subscribe, useSnapshot } from "valtio";
import { handleColorSchemeChange, DARK_MODE_MEDIA_QUERY } from "@stores/app";
import queryClient from "./lib/query-client";

import ErrorComponent from "./components/route/error-component";
import NotFound from "./components/route/not-found";
import stores from "./stores";
import { HotkeysProvider } from "react-hotkeys-hook";
import SplashScreen from "./components/splash-screen";

// Create a new router instance
const router = createRouter({
	routeTree,
	defaultNotFoundComponent: NotFound,
	defaultPendingComponent: SplashScreen,
	defaultPendingMs: 900,
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router;
	}
}

export default function App() {
	const appState = useSnapshot(stores.app);

	// React to colorscheme changes
	useEffect(() => {
		const systemThemeQuery = window.matchMedia(DARK_MODE_MEDIA_QUERY);

		// Dispatch the initial color scheme change event on startup
		handleColorSchemeChange({ isDarkMode: systemThemeQuery.matches });

		// Watch for changes in the color scheme
		systemThemeQuery.addEventListener("change", (e) => {
			handleColorSchemeChange({ isDarkMode: e.matches });
		});

		// Dispatch the handleColorSchemeChange function when the store value changes
		const unsubscribe = subscribe(stores.app, () => {
			handleColorSchemeChange({ isDarkMode: systemThemeQuery.matches });
		});

		return () => {
			systemThemeQuery.removeEventListener("change", (e) => {
				handleColorSchemeChange({ isDarkMode: e.matches });
			});

			unsubscribe();
		};
	}, []);

	const isDarkMode = useMemo(() => {
		return (
			appState.colorScheme === "dark" ||
			(appState.colorScheme === "inherit" &&
				window.matchMedia(DARK_MODE_MEDIA_QUERY).matches)
		);
	}, [appState.colorScheme]);

	return (
		<StrictMode>
			<HotkeysProvider>
				<QueryClientProvider client={queryClient}>
					<Theme
						accentColor={appState.accentColor}
						appearance={appState.colorScheme}
						panelBackground="solid"
						radius="large"
						scaling={appState.uiScale}
						grayColor={appState.grayColor}
					>
						<RouterProvider
							router={router}
							defaultErrorComponent={ErrorComponent}
						/>
					</Theme>
				</QueryClientProvider>
				<Toaster
					richColors
					position="bottom-right"
					theme={isDarkMode ? "dark" : "light"}
					className="pointer-events-auto"
				/>
			</HotkeysProvider>
		</StrictMode>
	);
}
