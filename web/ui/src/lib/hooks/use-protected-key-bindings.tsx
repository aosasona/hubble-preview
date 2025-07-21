import useShortcut from "$/lib/hooks/use-shortcut";
import stores from "$/stores";
import { useNavigate, useParams, useRouter } from "@tanstack/react-router";
import { useCallback, useMemo } from "react";
import { useSnapshot } from "valtio";

export default function useProtectedKeyBindings() {
	const params = useParams({ from: "/_protected/$workspaceSlug" });

	const navigate = useNavigate();
	const router = useRouter();

	const app = useSnapshot(stores.app);
	const workspaces = useSnapshot(stores.workspace);

	const otherWorkspaces = useMemo(() => {
		return workspaces.all.filter((w) => w.slug !== params.workspaceSlug);
	}, [params.workspaceSlug, workspaces.all]);

	const switchWorkspace = useCallback(
		(index: number) => {
			if (index < 0) return;
			if (index > otherWorkspaces.length) return;

			navigate({
				to: "/$workspaceSlug",
				params: {
					workspaceSlug: otherWorkspaces?.[index]?.slug ?? "",
				},
			});
		},
		[navigate, otherWorkspaces],
	);

	/** ==== KEYMAPS ==== **/
	// Dialogs
	useShortcut(
		["alt+c"],
		(e) => {
			app.openDialog("createCollection");

			// Prevent from inserting the character
			e.preventDefault();
			e.stopPropagation();
		},
		{
			enableOnFormTags: false,
		},
	);
	useShortcut(
		["c"],
		(e) => {
			app.openDialog("importItem");

			// Prevent from inserting the character
			e.preventDefault();
			e.stopPropagation();
		},
		{
			enableOnFormTags: false,
		},
	);

	// Navigation
	useShortcut(["alt+."], () => {
		navigate({ to: "/$workspaceSlug/settings", params });
	});
	useShortcut(["alt+a"], () => {
		navigate({ to: "/$workspaceSlug/settings/account", params });
	});
	useShortcut(["alt+p"], () => {
		navigate({ to: "/$workspaceSlug/settings/appearance", params });
	});

	// Workspace switching
	useShortcut(["alt+1"], () => switchWorkspace(0));
	useShortcut(["alt+2"], () => switchWorkspace(1));
	useShortcut(["alt+3"], () => switchWorkspace(2));
	useShortcut(["alt+4"], () => switchWorkspace(3));
	useShortcut(["alt+5"], () => switchWorkspace(4));
	useShortcut(["alt+6"], () => switchWorkspace(5));
	useShortcut(["alt+7"], () => switchWorkspace(6));
	useShortcut(["alt+8"], () => switchWorkspace(7));
	useShortcut(["alt+9"], () => switchWorkspace(8));

	// General navigation
	useShortcut(["ctrl+backspace"], () => {
		// Go back in history
		if (router.history.canGoBack()) {
			router.history.go(-1);
		}
	});

	// Move focus to the nearest `data-list` container
	useShortcut(["g"], () => {
		const list = document.querySelector("[data-focusable-list]") as HTMLElement;
		if (!list) return;

		// Focus on the one marked as the last selected in the list
		const lastFocusedIdx = list.getAttribute("data-last-focused");
		if (lastFocusedIdx) {
			const lastFocused = list.querySelector(
				`[data-list-item="${lastFocusedIdx}"]`,
			) as HTMLElement;
			if (lastFocused) {
				lastFocused.focus();
				return;
			}
		}

		// Otherwise, focus on the first element
		(list?.querySelector("[data-list-item]") as HTMLElement)?.focus();
	});

	useShortcut(["shift+g"], () => {
		const list = document.querySelector("[data-focusable-list]") as HTMLElement;
		// focus on the last element in the list
		if (list) {
			const items = list.querySelectorAll("[data-list-item]");
			(items[items.length - 1] as HTMLElement)?.focus();
		}
	});

	// Search
	useShortcut(["f", "/"], (e) => {
		e.preventDefault();
		e.stopPropagation();
		const input = document.querySelector<HTMLInputElement>("#search-entries");
		if (input) {
			input.focus();
			return;
		}

		// Else navigate to the search page
		navigate({
			to: "/$workspaceSlug/search",
			params: { workspaceSlug: params.workspaceSlug },
		});
	});
}
