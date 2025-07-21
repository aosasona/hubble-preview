package main

import (
	"go.trulyao.dev/hubble/web/api"
	"go.trulyao.dev/hubble/web/api/middleware"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/procedure"
)

func (a *App) attachProcedures() {
	r := a.robin

	r.Use("rate_limit", a.middleware.WithRateLimit)

	auth := a.handler.Auth()
	user := a.handler.User()
	mfa := a.handler.Mfa()
	workspace := a.handler.Workspace()
	collection := a.handler.Collection()
	entry := a.handler.Entry()
	plugin := a.handler.Plugin()

	//nolint:all
	gulterInstance, err := middleware.CreateGulterInstance(&middleware.GulterOptions{
		Bucket:      "entries",
		ObjectStore: a.objectsStore.WithBucket(objectstore.BucketEntries),
	})
	if err != nil {
		panic("failed to create gulter instance: " + err.Error())
	}

	// MARK: Queries
	query(r, procedure.Ping, ping)
	query(r, procedure.MfaLoadSession, mfa.FindSession, "/mfa/session")

	// -- MARK: Mutations
	// Auth
	mutation(r, procedure.SignIn, auth.SignIn, "/auth/sign-in")
	mutation(r, procedure.SignUp, auth.SignUp, "/auth/sign-up")
	mutation(r, procedure.VerifyEmail, auth.VerifyEmail, "/auth/email/verify")
	mutation(
		r,
		procedure.RequestPasswordReset,
		auth.SendPasswordResetEmail,
		"/auth/reset/request-email",
	)
	mutation(r, procedure.ChangePassword, auth.ChangePassword, "/auth/reset/change-password")

	// MFA
	mutation(
		r,
		procedure.RequestEmailVerification,
		auth.ResendVerificationEmail,
		"/auth/email/resend",
	)
	mutation(r, procedure.MfaInitiateAuth, mfa.InitiateAuthSession, "/mfa/initiate")
	mutation(r, procedure.MfaVerifyAuth, mfa.VerifyAuthSession, "/mfa/verify")
	mutation(r, procedure.MfaResendEmail, mfa.ResendMfaEmail, "/mfa/email/resend")

	// -- MARK: PROTECTED Queries
	a.protectAll(
		query(r, procedure.Me, user.Me, "/@me"),
		query(r, procedure.MfaState, mfa.GetUserMfaState, "/mfa"),

		// Entry
		query(r, procedure.GetLinkMetadata, entry.GetLinkMetadata, "/entry/url/lookup"),
		query(r, procedure.FindEntry, entry.Find, "/entry"),
		query(r, procedure.SearchEntries, entry.Search, "/entry/search"),

		// WORKSPACE
		query(r, procedure.FindWorkspace, workspace.Find, "/workspace"),
		query(r, procedure.ListWorkspaceEntries, entry.FindWorkspaceEntries, "/workspace/entries"),
		query(r, procedure.ListWorkspaceMembers, workspace.ListMembers, "/workspace/members"),
		query(r, procedure.FindInvite, workspace.FindInvite, "/workspace/invite"),
		query(
			r,
			procedure.LoadWorkspaceMemberStatus,
			workspace.LoadMemberStatus,
			"/workspace/member/status",
		),

		// COLLECTIONS
		query(
			r,
			procedure.ListCollectionEntries,
			entry.FindCollectionEntries,
			"/collection/entries",
		),
		query(
			r,
			procedure.LoadCollectionMemberStatus,
			collection.LoadCollectionDetails,
			"/collection/member/status",
		),
		query(r, procedure.ListCollectionMembers, collection.ListMembers, "/collection/members"),

		// PLUGINS
		query(r, procedure.ListPluginSources, plugin.ListSources, "/plugin/sources"),
		query(r, procedure.ListPlugins, plugin.ListPlugins, "/plugins"),
	)

	// MARK: PROTECTED Mutations
	a.protectAll(
		// Auth
		mutation(r, procedure.SignOut, auth.SignOut, "/auth/logout"),
		mutation(r, procedure.RequestEmailChange, auth.RequestEmailChange, "/auth/email/change"),
		mutation(
			r,
			procedure.VerifyEmailChange,
			auth.VerifyEmailChange,
			"/auth/email/change/verify",
		),

		// MFA
		mutation(
			r,
			procedure.MfaCreateEmailAccount,
			mfa.CreateEmailAccount,
			"/mfa/accounts/email",
		),
		mutation(
			r,
			procedure.MfaActivateEmailAccount,
			mfa.ActivateEmailAccount,
			"/mfa/email/activate",
		),
		mutation(
			r,
			procedure.MfaSetDefaultAccount,
			mfa.SetPreferredAccount,
			"/mfa/account/preferred",
		),
		mutation(
			r,
			procedure.MfaRegenerateBackupCodes,
			mfa.RegenerateBackupCodes,
			"/mfa/backup/regenerate",
		),
		mutation(
			r,
			procedure.MfaStartTotpEnrolment,
			mfa.StartTotpEnrollmentSession,
			"/mfa/totp/start",
		),
		mutation(
			r,
			procedure.MfaCompleteTotpEnrolment,
			mfa.CompleteTotpEnrollment,
			"/mfa/totp/complete",
		),
		mutation(r, procedure.MfaRenameAccount, mfa.RenameAccount, "/mfa/account/rename"),
		mutation(r, procedure.MfaDeleteAccount, mfa.DeleteAccount, "/mfa/account/delete"),

		// User
		mutation(r, procedure.SaveProfile, user.SaveProfile, "/@me/profile"),

		// Workspace
		mutation(r, procedure.CreateWorkspace, workspace.Create, "/workspace/create"),
		mutation(r, procedure.UpdateWorkspaceDetails, workspace.Update, "/workspace/update"),
		mutation(r, procedure.DeleteWorkspace, workspace.Delete, "/workspace/delete"),
		mutation(r, procedure.InviteUsersToWorkspace, workspace.InviteUsers, "/workspace/invite"),
		mutation(
			r,
			procedure.ChangeWorkspaceMemberRole,
			workspace.ChangeMemberRole,
			"/workspace/member/role",
		),
		mutation(
			r,
			procedure.UpdateWorkspaceInviteStatus,
			workspace.UpdateInviteStatus,
			"/workspace/invite/status",
		),
		mutation(
			r,
			procedure.RemoveMemberFromWorkspace,
			workspace.RemoveMember,
			"/workspace/member/remove",
		),

		// Collections
		mutation(r, procedure.CreateCollection, collection.Create, "/collection/create"),
		mutation(r, procedure.UpdateCollectionDetails, collection.Update, "/collection/update"),
		mutation(
			r,
			procedure.AddMembersToCollection,
			collection.AddMembers,
			"/collection/member/add",
		),
		mutation(
			r,
			procedure.RemoveMembersFromCollection,
			collection.RemoveMembers,
			"/collection/member/remove",
		),
		mutation(r, procedure.LeaveCollection, collection.Leave, "/collection/leave"),
		mutation(r, procedure.DeleteCollection, collection.Delete, "/collection/delete"),

		// Entries
		mutation(
			r,
			procedure.ImportEntries,
			entry.Import,
			"/entry/import",
		).
			WithMiddleware(a.middleware.WithGulter(gulterInstance, []string{"files"})).
			WithRawPayload(api.ImportEntryPayload{}), // nolint:exhaustruct

		mutation(r, procedure.DeleteEntries, entry.Delete, "/entry/delete"),
		mutation(r, procedure.RequeueEntries, entry.Requeue, "/entry/requeue"),

		// Plugins
		mutation(r, procedure.FindPluginSource, plugin.FindSourceByURL, "/plugin/source/lookup"),
		mutation(r, procedure.AddPluginSource, plugin.AddSource, "/plugin/source/add"),
		mutation(r, procedure.RemovePluginSource, plugin.RemoveSource, "/plugin/source/remove"),
		mutation(r, procedure.InstallPlugin, plugin.InstallPlugin, "/plugin/install"),
		mutation(r, procedure.UpdatePlugin, plugin.UpdatePlugin, "/plugin/update"),
		mutation(r, procedure.RemovePlugin, plugin.RemovePlugin, "/plugin/remove"),
	)
}
