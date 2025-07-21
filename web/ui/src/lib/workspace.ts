import type { Workspace } from "$/lib/server/types";
import stores from "$/stores";
import { redirect } from "@tanstack/react-router";

export function beforeWorkspaceLoad() {
	// If there are no workspaces, redirect to the workspace creation page
	if (stores.workspace.all.length === 0) {
		throw redirect({ to: "/workspace/new", replace: true });
	}

	/*
	 * If there are workspaces, redirect to the one of the following in this order:
	 *
	 * - The active workspace, if persisted
	 * - The default workspace, if set (the active workspace might have been deleted, so we fallback to the default)
	 * - The first workspace in the list
	 */
	let workspace: Workspace | null = null;

	const defaultWorkspaceId = stores.workspace.defaultWorkspaceId;
	const activeWorkspace = stores.workspace.activeWorkspace;

	if (!empty(activeWorkspace)) {
		workspace = stores.workspace.findWorkspaceById(activeWorkspace.id);
	}

	if (workspace == null && !empty(defaultWorkspaceId)) {
		workspace = stores.workspace.findWorkspaceById(defaultWorkspaceId);

		// If the default workspace is not found, clean it up
		if (!workspace) {
			stores.workspace.removeDefaultWorkspace();
		}
	}

	if (!workspace) workspace = stores.workspace.all[0];

	stores.workspace.setActiveWorkspace(workspace, []);

	throw redirect({
		to: "/$workspaceSlug",
		params: { workspaceSlug: workspace.slug },
		replace: true,
	});
}

function empty<T>(value: T | null | undefined): value is null | undefined {
	return value === null || value === undefined;
}
