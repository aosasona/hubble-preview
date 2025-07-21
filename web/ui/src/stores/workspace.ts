import QueryKeys from "$/lib/keys";
import queryClient from "$/lib/query-client";
import type { QueryResult } from "$/lib/server";
import type { Collection } from "$/lib/server/types";
import type { FieldOf, OneOf } from "$/lib/types";
import { ACCENT_COLORS, type AccentColor } from "./app";
import LocalStorage from "./local-storage";
import { proxyWithPersist } from "./persist";

type Workspace = OneOf<FieldOf<QueryResult<"me">, "workspaces">>;

type StoreKeys = "active" | "collections";
const localStore = new LocalStorage<StoreKeys>("user:workspace");

type SetActiveCollectionParams = { slug: string } | { collection: Collection };

type WorkspaceStore = {
	state: "loading" | "loaded" | "error";

	activeCollection: Collection | null;
	activeWorkspace: Workspace | null;
	activeWorkspaceCollections: Collection[];
	shouldSaveActiveWorkspace: boolean;

	all: Workspace[];
	defaultWorkspaceId: string | null;

	/** The current color scheme */
	recentSearchQueries: Array<string>;

	// Computed color based on the active workspace name
	accentColor: AccentColor;
	computedAccentColorsCache: Record<string, AccentColor>;
	computeAccentColor(name: string): AccentColor;

	/** Set the workspaces */
	setWorkspaces: (workspaces: Workspace[]) => void;
	/** Add a new workspace */
	addWorkspace: (workspace: Workspace) => void;

	/** Add a new collection to the active workspace */
	addCollection: (collection: Collection) => void;

	/** Set the current collection */
	setActiveCollection: (params: SetActiveCollectionParams) => void;

	/** Remove the active collection */
	removeActiveCollection: () => void;

	/** Remove a collection from the active workspace */
	removeCollection: (collectionId: string) => void;

	/** Set the active workspace and its collections and persist it if enabled */
	setActiveWorkspace: (workspace: Workspace, collections: Collection[]) => void;
	/** Toggle the active workspace persistence */
	toggleSaveActiveWorkspace: () => void;

	/** Set the default workspace */
	setDefaultWorkspace: (workspaceId: string) => void;
	/** Remove the default workspace */
	removeDefaultWorkspace: () => void;

	/** Find a workspace by its ID */
	findWorkspaceById: (id: string) => Workspace | null;

	/** Find a workspace by its slug */
	findWorkspaceBySlug: (slug: string) => Workspace | null;

	/** Find a collection by its ID */
	findCollectionById: (id: string) => Collection | null;

	/** Find a collection by its ID in the active workspace */
	findWorkspaceCollectionById: (args: {
		workspaceId: string;
		collectionId: string;
	}) => Collection | null;

	/** Add a recent search query to the list */
	addRecentSearchQuery: (query: string) => void;

	/** Remove a recent search query from the list */
	removeRecentSearchQuery: (query: string) => void;

	/** Clear all recent search queries */
	clearRecentSearchQueries: () => void;
};

function getCachedCollections(): Collection[] {
	const activeWorkspace = localStore.get<Workspace>("active");
	if (!activeWorkspace) return [];

	const cachedQuery = queryClient.getQueryData(
		QueryKeys.FindWorkspace(activeWorkspace.slug),
	);

	if (!cachedQuery) {
		return localStore.get<Collection[]>("collections", []) ?? [];
	}

	return (cachedQuery as QueryResult<"workspace.find">)
		?.collections as Collection[];
}

export function calculateColorHash(name: string): AccentColor {
	let hash = 0;

	for (let i = 0; i < name.length; i++) {
		hash = (hash * 31 + name.charCodeAt(i)) & 0xffffffff; // ensure 32 bits integer
	}

	return ACCENT_COLORS[Math.abs(hash) % ACCENT_COLORS.length];
}

const workspaceStore = proxyWithPersist<WorkspaceStore>("workspace", {
	defaultValue: {
		state: "loading",

		all: [],
		defaultWorkspaceId: null,
		activeCollection: null,

		recentSearchQueries: [],

		shouldSaveActiveWorkspace: false,
		activeWorkspace: localStore.get<Workspace>("active"),
		activeWorkspaceCollections: getCachedCollections(),

		accentColor: "orange",
		computedAccentColorsCache: {},
		computeAccentColor(name: string): AccentColor {
			if (name in workspaceStore.computedAccentColorsCache) {
				return workspaceStore.computedAccentColorsCache[name];
			}

			const color = calculateColorHash(name?.trim()?.toLowerCase());
			workspaceStore.computedAccentColorsCache[name] = color;

			return color;
		},

		setWorkspaces(workspaces) {
			workspaceStore.all = workspaces;
		},

		addWorkspace(workspace) {
			workspaceStore.all.push(workspace);
		},

		addCollection(collection) {
			workspaceStore.activeWorkspaceCollections.push(collection);
		},

		setActiveCollection(params) {
			// Attempt to find the collection by its slug
			if ("slug" in params) {
				const collection = workspaceStore.activeWorkspaceCollections.find(
					(collection) => collection.slug === params.slug,
				);

				if (collection) {
					workspaceStore.activeCollection = collection;
				}

				return;
			}

			workspaceStore.activeCollection = params.collection;
		},

		removeActiveCollection() {
			workspaceStore.activeCollection = null;
		},

		removeCollection(collectionId) {
			workspaceStore.activeWorkspaceCollections =
				workspaceStore.activeWorkspaceCollections.filter(
					(collection) => collection.id !== collectionId,
				);
		},

		setActiveWorkspace(workspace, collections) {
			workspaceStore.activeWorkspace = workspace;
			workspaceStore.activeWorkspaceCollections = collections ?? [];
			workspaceStore.state = "loaded";

			if (workspaceStore.shouldSaveActiveWorkspace) {
				localStore.set("active", workspace);
				localStore.set("collections", collections);
			}

			workspaceStore.accentColor = workspaceStore.computeAccentColor(
				workspace.name,
			);
		},

		toggleSaveActiveWorkspace() {
			workspaceStore.shouldSaveActiveWorkspace =
				!workspaceStore.shouldSaveActiveWorkspace;

			// Clear the active workspace if the user disabled the persistence
			if (!workspaceStore.shouldSaveActiveWorkspace) {
				localStore.remove("active");
				localStore.remove("collections");
			}
		},

		setDefaultWorkspace(workspaceId) {
			workspaceStore.defaultWorkspaceId = workspaceId;
		},

		removeDefaultWorkspace() {
			workspaceStore.defaultWorkspaceId = null;
		},

		findWorkspaceById(id: string): Workspace | null {
			return (
				workspaceStore.all.find((workspace) => workspace.id === id) ?? null
			);
		},

		findWorkspaceBySlug(slug: string): Workspace | null {
			return (
				workspaceStore.all.find((workspace) => workspace.slug === slug) ?? null
			);
		},

		findCollectionById(id: string): Collection | null {
			return (
				workspaceStore.activeWorkspaceCollections.find(
					(collection) => collection.id === id,
				) ?? null
			);
		},

		findWorkspaceCollectionById(args: {
			workspaceId: string;
			collectionId: string;
		}): Collection | null {
			const { workspaceId, collectionId } = args;
			const workspace = workspaceStore.all.find(
				(workspace) => workspace.id === workspaceId,
			);

			if (!workspace) {
				return null;
			}

			return (
				workspace.collections?.find(
					(collection) => collection.id === collectionId,
				) ?? null
			);
		},

		// TODO: save queries in the local storage WITH the workspace prepended so that queries are unique to the workspace
		addRecentSearchQuery(query: string) {
			// If the array is not defined, initialize it
			if (typeof workspaceStore.recentSearchQueries === "undefined") {
				workspaceStore.recentSearchQueries = [];
			}

			if (query.length === 0) {
				return;
			}

			if (workspaceStore.recentSearchQueries.includes(query)) {
				workspaceStore.removeRecentSearchQuery(query);
			}

			workspaceStore.recentSearchQueries.push(query);
		},

		removeRecentSearchQuery(query: string) {
			workspaceStore.recentSearchQueries.splice(
				workspaceStore.recentSearchQueries.indexOf(query),
				1,
			);
		},

		clearRecentSearchQueries() {
			workspaceStore.recentSearchQueries.splice(
				0,
				workspaceStore.recentSearchQueries.length,
			);
		},
	},
	excludeKeys: [
		"state",
		"activeCollection",
		"activeWorkspace",
		"activeWorkspaceCollections",
	],
});

export default workspaceStore;
