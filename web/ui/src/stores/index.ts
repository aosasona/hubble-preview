import client, { type QueryResult } from "$/lib/server";
import appStore from "./app";
import authStore from "./auth";
import layoutStore from "./layout";
import workspaceStore from "./workspace";
import uploadsStore from "./uploads";
import entriesListStore from "./entries-list";
import queryClient from "$/lib/query-client";

async function load(): Promise<QueryResult<"me"> | null> {
	try {
		const data = await client.queries.me();

		if (!data) {
			workspaceStore.state = "error";
			authStore.state = "error";
			return null;
		}

		// Set the user data
		authStore.setUser(data.user);

		// Set workspace data
		workspaceStore.setWorkspaces(data.workspaces);

		authStore.state = "loaded";
		workspaceStore.state = "loaded";

		return data;
	} catch (error) {
		if (import.meta.env.DEV) console.error(error);

		queryClient.invalidateQueries({ queryKey: ["me"] }); // Invalidate the query to refetch on next mount

		authStore.state = "error";
		workspaceStore.state = "error";

		return null;
	}
}

export default {
	app: appStore,
	auth: authStore,
	workspace: workspaceStore,
	layout: layoutStore,
	uploads: uploadsStore,
	entriesList: entriesListStore,
	load,
};
