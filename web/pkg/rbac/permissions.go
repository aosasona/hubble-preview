package rbac

import (
	"net/http"

	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
)

var ErrPermissionDenied = apperrors.New("permission denied", http.StatusForbidden)

type Permission string

const (
	PermDeleteWorkspace            Permission = "workspace:delete"
	PermRenameWorkspace            Permission = "workspace:rename"
	PermChangeWorkspaceSlug        Permission = "workspace:slug:update"
	PermChangeWorkspaceDescription Permission = "workspace:description:update"

	PermInviteUsersToWorkspace Permission = "workspace:members:invite"
	PermListWorkspaceMembers   Permission = "workspace:members:list"
	PermListWorkspaceEntries   Permission = "workspace:entries:list"
	PermRevokeWorkspaceInvite  Permission = "workspace:invite:revoke"

	PermChangeMemberRole     Permission = "workspace:members:role:update"
	PermMakeOwnerOfWorkspace Permission = "workspace:members:role:to_owner"
	PermViewMembersEmail     Permission = "workspace:members:view_email"

	PermRemoveUserFromWorkspace  Permission = "workspace:members:remove"
	PermRemoveAdminFromWorkspace Permission = "workspace:members:remove_admin"

	PermAddMemberToCollection       Permission = "collection:members:add"
	PermRemoveMemberFromCollection  Permission = "collection:members:remove"
	PermDeleteCollection            Permission = "collection:delete"
	PermRenameCollection            Permission = "collection:rename"
	PermChangeCollectionSlug        Permission = "collection:slug:update"
	PermChangeCollectionDescription Permission = "collection:description:update"

	PermListCollections       Permission = "collections:list"
	PermCreateCollection      Permission = "collection:create"
	PermListCollectionEntries Permission = "collection:entries:list"
	PermListCollectionMembers Permission = "collection:members:list"

	PermCreateEntry  Permission = "entry:create"
	PermReadEntry    Permission = "entry:read"
	PermDeleteEntry  Permission = "entry:delete"
	PermRequeueEntry Permission = "entry:requeue"
	PermSearchEntry  Permission = "entry:search"

	PermAddPluginSource    Permission = "plugin:source:add"
	PermRemovePluginSource Permission = "plugin:source:remove"
	PermViewPluginSource   Permission = "plugin:source:view"
	PermListPluginSources  Permission = "plugin:source:list"

	PermInstallPlugin   Permission = "plugin:install"
	PermUninstallPlugin Permission = "plugin:uninstall"
)

type Permissions map[Permission]Role

var permissions = Permissions{
	// Workspace
	PermDeleteWorkspace:            RoleOwner,
	PermRenameWorkspace:            RoleOwner,
	PermChangeWorkspaceSlug:        RoleOwner,
	PermChangeWorkspaceDescription: CombineRoles(RoleAdmin, RoleOwner),

	PermInviteUsersToWorkspace: CombineRoles(RoleAdmin, RoleOwner),
	PermListWorkspaceMembers:   CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermListWorkspaceEntries:   CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermRevokeWorkspaceInvite:  CombineRoles(RoleAdmin, RoleOwner),

	PermChangeMemberRole:     CombineRoles(RoleAdmin, RoleOwner),
	PermMakeOwnerOfWorkspace: RoleOwner,
	PermViewMembersEmail:     CombineRoles(RoleAdmin, RoleOwner),

	PermRemoveUserFromWorkspace:  CombineRoles(RoleAdmin, RoleOwner),
	PermRemoveAdminFromWorkspace: CombineRoles(RoleAdmin, RoleOwner),

	// Collection
	PermAddMemberToCollection:       CombineRoles(RoleAdmin, RoleOwner),
	PermRemoveMemberFromCollection:  CombineRoles(RoleAdmin, RoleOwner),
	PermDeleteCollection:            CombineRoles(RoleAdmin, RoleOwner),
	PermRenameCollection:            CombineRoles(RoleOwner, RoleAdmin),
	PermChangeCollectionSlug:        RoleOwner,
	PermChangeCollectionDescription: CombineRoles(RoleAdmin, RoleOwner),

	PermListCollections:       CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermCreateCollection:      CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermListCollectionEntries: CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermListCollectionMembers: CombineRoles(RoleAdmin, RoleOwner, RoleUser),

	// Entry
	PermCreateEntry:  CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermReadEntry:    CombineRoles(RoleAdmin, RoleOwner, RoleUser, RoleGuest),
	PermDeleteEntry:  CombineRoles(RoleAdmin, RoleOwner),
	PermRequeueEntry: CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermSearchEntry:  CombineRoles(RoleAdmin, RoleOwner, RoleUser, RoleGuest),

	// Plugin
	PermAddPluginSource:    CombineRoles(RoleAdmin, RoleOwner),
	PermRemovePluginSource: CombineRoles(RoleAdmin, RoleOwner),
	PermViewPluginSource:   CombineRoles(RoleAdmin, RoleOwner, RoleUser),
	PermListPluginSources:  CombineRoles(RoleAdmin, RoleOwner, RoleUser),

	PermInstallPlugin:   CombineRoles(RoleAdmin, RoleOwner),
	PermUninstallPlugin: CombineRoles(RoleAdmin, RoleOwner),
}
