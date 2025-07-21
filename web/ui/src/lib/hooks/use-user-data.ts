import stores from "$/stores";
import workspaceStore from "$/stores/workspace";
import { useEffect } from "react";
import useRobinQuery from "./use-robin-query";
import type { User, Workspace } from "../server/types";
import { useSnapshot } from "valtio";
import { extractError } from "../error";
import { useMatches, useNavigate } from "@tanstack/react-router";

type Result = {
	user: User | null;
	workspaces: Workspace[];
	isLoading: boolean;
};

export default function useUserData(): Result {
	const matches = useMatches();
	const navigate = useNavigate();

	const query = useRobinQuery("me", undefined, { retry: false });
	const user = useSnapshot(stores.auth);

	useEffect(() => {
		if (query.isFetched && query.data) {
			user.setUser(query.data.user);
			workspaceStore.setWorkspaces(query.data.workspaces);
		} else if (query.isError) {
			const err = extractError(query.error);
			const isProtectedRoute = matches.some((match) => {
				return match.id?.includes("/_protected");
			});

			if (err?.type === "authz" && isProtectedRoute) {
				// Check if we are in a protected route
				user.clear();
				navigate({
					to: "/auth/sign-in",
					search: { redirect: window.location.pathname },
					replace: true,
				});
			}
		}
	}, [
		navigate,
		query.data,
		query.isFetched,
		query.isError,
		query.error,
		user,
		matches,
	]);

	return {
		user: query.data?.user || null,
		workspaces: query.data?.workspaces || [],
		isLoading: query.isLoading,
	};
}
