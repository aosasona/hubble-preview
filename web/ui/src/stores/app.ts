import { proxyWithPersist } from "./persist";

export const ACCENT_COLORS = [
	"gray",
	"gold",
	"bronze",
	"brown",
	"yellow",
	"amber",
	"orange",
	"tomato",
	"red",
	"ruby",
	"crimson",
	"pink",
	"plum",
	"purple",
	"violet",
	"iris",
	"indigo",
	"blue",
	"cyan",
	"teal",
	"jade",
	"green",
	"grass",
	"lime",
	"mint",
	"sky",
] as const;

// eslint-disable-next-line @typescript-eslint/no-unused-vars
const GRAY_COLORS = [
	"auto",
	"gray",
	"mauve",
	"slate",
	"sage",
	"olive",
	"sand",
] as const;

const BACKGROUND_COLORS = {
	light: "#fefefe",
	dark: "#101010",
};

export const UI_SCALES = ["90%", "95%", "100%", "105%", "110%"] as const;

export type ColorScheme = "inherit" | "light" | "dark";
export type AccentColor = (typeof ACCENT_COLORS)[number];
export type GrayColor = (typeof GRAY_COLORS)[number];
export type UIScale = (typeof UI_SCALES)[number];

type Dialog =
	| "createCollection"
	| "deleteSelectedEntries"
	| "importItem"
	| "mobileSidebar";

export type AppStore = {
	/** The color scheme to use: this can be one of "light", "dark" or "system" (inherit) */
	colorScheme: ColorScheme;

	/** The accent color to use */
	accentColor: AccentColor;

	/** The gray color to use */
	grayColor: GrayColor;

	/** The UI scale percentage */
	uiScale: UIScale;

	/** Dialogs keeps track of the various global dialogs in the system */
	dialogs: Record<Dialog, boolean>;

	set: <T extends keyof AppStore>(key: T, value: AppStore[T]) => void;
	setColorScheme: (scheme: ColorScheme) => void;
	toggleColorScheme: () => void;
	switchToSystemColorScheme: () => void;

	openDialog: (dialog: Dialog) => void;
	closeDialog: (dialog: Dialog) => void;
	setDialogState: (dialog: Dialog, value: boolean) => void;

	toggleMobileSidebar: () => void;
};

const appStore = proxyWithPersist<AppStore>("app", {
	defaultValue: {
		colorScheme: "dark",
		accentColor: "orange",
		grayColor: "auto",
		uiScale: "100%",
		dialogs: {
			createCollection: false,
			importItem: false,
			mobileSidebar: false,
			deleteSelectedEntries: false,
		},

		set(key, value) {
			appStore[key] = value;
		},

		setColorScheme(scheme: ColorScheme) {
			appStore.colorScheme = scheme;
		},

		toggleColorScheme() {
			appStore.setColorScheme(
				appStore.colorScheme === "light" ? "dark" : "light",
			);
		},

		switchToSystemColorScheme() {
			appStore.setColorScheme("inherit");
		},

		setDialogState(dialog: Dialog, value: boolean) {
			appStore.dialogs[dialog] = value;
		},

		openDialog(dialog: Dialog) {
			appStore.dialogs[dialog] = true;
		},

		closeDialog(dialog: Dialog) {
			appStore.dialogs[dialog] = false;
		},

		toggleMobileSidebar() {
			appStore.dialogs.mobileSidebar = !appStore.dialogs.mobileSidebar;
		},
	},
	excludeKeys: ["dialogs"],
});

export const DARK_MODE_MEDIA_QUERY = "(prefers-color-scheme: dark)";

export function handleColorSchemeChange(e: { isDarkMode: boolean }) {
	let scheme = appStore.colorScheme;
	if (scheme === "inherit") {
		scheme = e.isDarkMode ? "dark" : "light";
	}

	const themeColorMeta = document.querySelector('meta[name="theme-color"]');
	const documentElement = document.documentElement;

	if (appStore.colorScheme === "inherit") {
		// Remove the color scheme class
		documentElement.classList.remove(
			"dark",
			"light",
			"dark-theme",
			"light-theme",
		);

		documentElement.classList.add(scheme);
	}

	// Update the theme-color meta
	if (themeColorMeta) {
		themeColorMeta.setAttribute("content", BACKGROUND_COLORS[scheme]);
	}
}

export default appStore;
