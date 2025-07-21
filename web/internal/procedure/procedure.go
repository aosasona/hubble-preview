package procedure

// Queries
var (
	Ping = "ping"
	Me   = "me"

	MfaLoadSession = "mfa.load-session"
	MfaState       = "mfa.state"

	FindWorkspace  = "workspace.find"
	FindInvite     = "workspace.invite.find"
	FindCollection = "collection.find"

	LoadCollectionMemberStatus = "collection.member.status"
	LoadWorkspaceMemberStatus  = "workspace.member.status"

	ListWorkspaceEntries  = "workspace.entries.all"
	ListCollectionEntries = "collection.entries.all"
	ListWorkspaceMembers  = "workspace.members.all"
	ListCollectionMembers = "collection.members.all"

	ListPluginSources = "plugin.source.list"
	ListPlugins       = "plugin.list"
)

// Mutations
var (
	SignIn                   = "auth.sign-in"
	SignUp                   = "auth.sign-up"
	SignOut                  = "auth.sign-out"
	VerifyEmail              = "auth.verify-email"
	RequestEmailVerification = "auth.request-email-verification"
	RequestPasswordReset     = "auth.request-password-reset"
	ChangePassword           = "auth.change-password"

	RequestEmailChange = "user.request-email-change"
	VerifyEmailChange  = "user.verify-email-change"
	SaveProfile        = "user.save-profile"

	MfaInitiateAuth          = "mfa.initiate-auth-session"
	MfaVerifyAuth            = "mfa.verify-auth-session"
	MfaResendEmail           = "mfa.resend-email"
	MfaCreateEmailAccount    = "mfa.create-email-account"
	MfaActivateEmailAccount  = "mfa.activate-email-account"
	MfaRenameAccount         = "mfa.rename-account"
	MfaDeleteAccount         = "mfa.delete-account"
	MfaSetDefaultAccount     = "mfa.set-default-account"
	MfaRegenerateBackupCodes = "mfa.regenerate-backup-codes"
	MfaStartTotpEnrolment    = "mfa.start-totp-enrollment"
	MfaCompleteTotpEnrolment = "mfa.complete-totp-enrollment"

	CreateWorkspace             = "workspace.create"
	InviteUsersToWorkspace      = "workspace.invite"
	UpdateWorkspaceInviteStatus = "workspace.invite.status.update"
	ChangeWorkspaceMemberRole   = "workspace.member.role.update"
	RemoveMemberFromWorkspace   = "workspace.member.remove"
	UpdateWorkspaceDetails      = "workspace.details.update"
	DeleteWorkspace             = "workspace.delete"

	CreateCollection            = "collection.create"
	DeleteCollection            = "collection.delete"
	AddMembersToCollection      = "collection.members.add"
	RemoveMembersFromCollection = "collection.members.remove"
	LeaveCollection             = "collection.leave"
	UpdateCollectionDetails     = "collection.details.update"

	GetLinkMetadata = "get-link-metadata"
	ImportEntries   = "entry.import"
	DeleteEntries   = "entry.delete"
	RequeueEntries  = "entry.requeue"
	FindEntry       = "entry.find"
	SearchEntries   = "entry.search"

	FindPluginSource   = "plugin.source.find"
	AddPluginSource    = "plugin.source.add"
	RemovePluginSource = "plugin.source.remove"

	InstallPlugin = "plugin.install"
	UpdatePlugin  = "plugin.update"
	RemovePlugin  = "plugin.remove"
)
