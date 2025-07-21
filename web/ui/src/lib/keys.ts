const QueryKeys = {
	InitiateMfaAuthSession: (id: string) => ["mfa", "initiate-auth-session", id],
	VerifyMfaAuthSession: (id: string) => ["mfa", "verify-auth-session", id],

	FindWorkspace: (slug: string) => ["workspace", "find", slug],
	FindAllWorkspaceEntries: (slug: string) => ["workspace.entries.all", slug],
	FindAllCollectionEntries: (workspaceSlug: string, collectionSlug: string) => [
		"collection.entries.all",
		workspaceSlug,
		collectionSlug,
	],
	ListWorkspaceMembers: (id: string, page: number) => [
		"workspace.members",
		id,
		page,
	],
	ListCollectionMembers: (
		workspaceId: string,
		collectionId: string,
		page: number,
	) => ["collection.members", workspaceId, collectionId, page],
	FindEntry: (id: string) => ["entry", "find", id],
};

export default QueryKeys;
