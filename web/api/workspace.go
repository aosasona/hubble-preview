package api

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/mail"
	"go.trulyao.dev/hubble/web/internal/mail/templates"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
)

type workspaceHandler struct {
	*baseHandler
}

// Delete implements WorkspaceHandler.
func (w *workspaceHandler) Delete(
	ctx *robin.Context,
	request DeleteWorkspaceRequest,
) (DeleteWorkspaceResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return DeleteWorkspaceResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return DeleteWorkspaceResponse{}, err
	}

	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return DeleteWorkspaceResponse{}, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return DeleteWorkspaceResponse{}, err
	}

	if result.Workspace == nil {
		return DeleteWorkspaceResponse{}, apperrors.New("workspace not found", http.StatusNotFound)
	}

	if !result.MembershipStatus.Role.Can(rbac.PermDeleteWorkspace) {
		return DeleteWorkspaceResponse{}, apperrors.Forbidden(
			"you do not have permission to delete this workspace",
		)
	}

	if err := w.repos.WorkspaceRepository().Delete(result.PublicID); err != nil {
		return DeleteWorkspaceResponse{}, seer.Wrap(
			"delete_workspace_procedure",
			err,
			"failed to delete workspace",
		)
	}

	log.Info().
		Int32("user_id", auth.UserID).
		Str("workspace_id", result.Workspace.ID).
		Msg("workspace deleted")

	return DeleteWorkspaceResponse{
		WorkspaceID: result.Workspace.ID,
	}, nil
}

// Update implements WorkspaceHandler.
func (w *workspaceHandler) Update(
	ctx *robin.Context,
	request UpdateWorkspaceRequest,
) (CreateWorkspaceResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return CreateWorkspaceResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return CreateWorkspaceResponse{}, err
	}

	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return CreateWorkspaceResponse{}, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return CreateWorkspaceResponse{}, err
	}

	if request.Name != "" && request.Name != result.Workspace.Name {
		// Ensure the user has permission to rename the workspace
		if !result.MembershipStatus.Role.Can(rbac.PermRenameWorkspace) {
			return CreateWorkspaceResponse{}, apperrors.NewValidationError(apperrors.ErrorMap{
				"name": {"Insufficient permissions to rename the workspace"},
			})
		}

		// Ensure the name is valid
		if len(request.Name) < 2 || len(request.Name) > 64 {
			return CreateWorkspaceResponse{}, apperrors.NewValidationError(apperrors.ErrorMap{
				"name": {"Must be between 2 and 64 characters"},
			})
		}

		request.Name = strings.TrimSpace(request.Name)
	}

	// If the slug is being changed, ensure the user has permission to rename the workspace
	if request.Slug != "" && request.Slug != result.Workspace.Slug {
		if !result.MembershipStatus.Role.Can(rbac.PermChangeWorkspaceSlug) {
			return CreateWorkspaceResponse{}, apperrors.NewValidationError(apperrors.ErrorMap{
				"slug": {"Insufficient permissions to change the workspace slug"},
			})
		}

		slugExists, err := w.repos.WorkspaceRepository().SlugExists(request.Slug)
		if err != nil {
			return CreateWorkspaceResponse{}, err
		}

		if slugExists {
			return CreateWorkspaceResponse{}, apperrors.NewValidationError(apperrors.ErrorMap{
				"slug": {"Slug is already in use by another workspace"},
			})
		}

		request.Slug = strings.TrimSpace(strings.ToLower(request.Slug))
	}

	// If the description is being changed, ensure the user has permission to change the description
	canChangeDescription := result.MembershipStatus.Role.Can(rbac.PermChangeWorkspaceDescription)
	if request.Description != "" && request.Description != result.Workspace.Description &&
		!canChangeDescription {
		return CreateWorkspaceResponse{}, apperrors.NewValidationError(apperrors.ErrorMap{
			"description": {"Insufficient permissions to change the workspace description"},
		})
	}

	// Update the workspace
	updated, err := w.repos.WorkspaceRepository().
		UpdateDetails(repository.UpdateWorkspaceDetailsArgs{
			Name:        request.Name,
			Description: request.Description,
			Slug:        request.Slug,
			WorkspaceID: result.ID,
		})
	if err != nil {
		return CreateWorkspaceResponse{}, err
	}

	return CreateWorkspaceResponse{Workspace: &updated}, nil
}

// RemoveMember implements WorkspaceHandler.
func (w *workspaceHandler) RemoveMember(
	ctx *robin.Context,
	request RemoveMemberRequest,
) (RemoveMemberResponse, error) {
	var response RemoveMemberResponse

	user, err := authlib.ExtractUser(ctx, w.repos.UserRepository())
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID}, //nolint:exhaustruct
		user.ID,
	)
	if err != nil {
		return response, err
	}

	//nolint:exhaustruct
	target, err := w.repos.WorkspaceRepository().FindMember(repository.FindMemberArg{
		WorkspaceID: workspaceID,
		Email:       request.Email,
	})
	if err != nil {
		return response, err
	}

	// If the target user is the owner of the workspace, don't allow
	if result.Workspace.OwnerID == target.User.ID {
		return response, apperrors.Forbidden("you cannot remove the owner of the workspace")
	}

	var (
		isWorkspaceCreator = result.Workspace.OwnerID == user.ID
		isOwnerRole        = result.MembershipStatus.Role == rbac.RoleOwner
		isRemovingSelf     = user.ID == target.User.ID
		isLowerPrivilege   = target.Role < result.MembershipStatus.Role

		canRemoveMember = result.MembershipStatus.Role.Can(rbac.PermRemoveUserFromWorkspace)
		canRemoveAdmin  = result.MembershipStatus.Role.Can(rbac.PermRemoveAdminFromWorkspace)
	)

	// if the target user is an admin, ensure the acting user has permission to remove admins
	if target.Role == rbac.RoleAdmin && !canRemoveAdmin {
		return response, apperrors.Forbidden("you do not have permission to remove admins")
	}

	// if the user is removing themselves and user is the owner of the workspace, don't allow
	// otherwirse, treat it as them leaving the workspace
	if isRemovingSelf && isWorkspaceCreator {
		return response, apperrors.Forbidden(
			"please transfer ownership before leaving the workspace",
		)
	}

	// Ensure the user has permission to remove members if they are not removing themselves
	if !isRemovingSelf && !canRemoveMember {
		return response, apperrors.Forbidden(
			"you do not have permission to remove members from this workspace",
		)
	}

	// Ensure users can only remove others with equal or lower privilege (unless they are an owner)
	if !isRemovingSelf && (!isOwnerRole && !isWorkspaceCreator) && !isLowerPrivilege {
		return response, apperrors.Forbidden(
			"you do not have permission to remove this user from the workspace",
		)
	}

	log.Info().
		Int32("user_id", user.ID).
		Int32("target_user_id", target.User.ID).
		Str("workspace_id", result.Workspace.ID).
		Msg("removing user from workspace")

	if err := w.repos.WorkspaceRepository().RemoveMember(repository.RemoveWorkspaceMemberArgs{
		WorkspaceID:      result.ID,
		TargetUserID:     target.User.ID,
		WorkspaceOwnerID: result.Workspace.OwnerID,
	}); err != nil {
		return response, err
	}

	response.FirstName = target.FirstName
	response.LastName = target.LastName
	response.MemberID = target.ID
	return response, nil
}

// ChangeMemberRole implements WorkspaceHandler.
func (w *workspaceHandler) ChangeMemberRole(
	ctx *robin.Context,
	request ChangeMemberRoleRequest,
) (ChangeMemberRoleResponse, error) {
	response := ChangeMemberRoleResponse{} //nolint:exhaustruct

	user, err := authlib.ExtractUser(ctx, w.repos.UserRepository())
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}
	targetUserId, err := lib.UUIDFromString(request.UserID)
	if err != nil {
		return response, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID}, //nolint:exhaustruct
		user.ID,
	)
	if err != nil {
		return response, err
	}

	// Prevent self-role change
	if request.UserID == user.PublicID {
		return response, apperrors.Forbidden("you cannot change your own role")
	}

	// Check if the user has permission to change roles
	if !result.MembershipStatus.Role.Can(rbac.PermChangeMemberRole) {
		return response, apperrors.Forbidden(
			"insufficient permissions to change member role",
		)
	}
	// NOTE: if the target role is an owner, the user must have the permission to make someone owner
	canMakeOwner := result.MembershipStatus.Role.Can(rbac.PermMakeOwnerOfWorkspace) &&
		result.Workspace.OwnerID == user.ID
	role := rbac.RoleFromString(request.Role)
	if role == rbac.RoleOwner && !canMakeOwner {
		return response, apperrors.Forbidden(
			"you do not have permission to make someone an owner",
		)
	}

	err = w.repos.WorkspaceRepository().UpdateMemberRole(repository.UpdateMemberRoleArgs{
		WorkspaceID: result.ID,
		UserID:      targetUserId,
		Role:        role,
	})
	if err != nil {
		return response, err
	}

	response.UserID = targetUserId
	response.NewRole = role
	return response, nil
}

// LoadMemberStatus implements WorkspaceHandler.
func (w *workspaceHandler) LoadMemberStatus(
	ctx *robin.Context,
	request LoadMemberStatusRequest,
) (LoadMemberStatusResponse, error) {
	var response LoadMemberStatusResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{Slug: request.WorkspaceSlug},
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if result.Workspace == nil {
		return response, apperrors.New("workspace not found", http.StatusNotFound)
	}
	if result.MembershipStatus == nil {
		return response, apperrors.New("unable to load member", http.StatusNotFound)
	}
	if !result.MembershipStatus.IsMember {
		return response, apperrors.New(
			"you are not a member of this workspace",
			http.StatusForbidden,
		)
	}

	return LoadMemberStatusResponse{
		Workspace: *result.Workspace,
		Status:    *result.MembershipStatus,
	}, nil
}

// UpdateInviteStatus implements WorkspaceHandler.
func (w *workspaceHandler) UpdateInviteStatus(
	ctx *robin.Context,
	request UpdateInviteStatusRequest,
) (UpdateInviteStatusResponse, error) {
	var response UpdateInviteStatusResponse

	user, err := authlib.ExtractUser(ctx, w.repos.UserRepository())
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	// Convert IDs
	var workspaceID, inviteID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return response, err
	}
	if inviteID, err = lib.UUIDFromString(request.InviteID); err != nil {
		return response, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID},
		user.ID,
	)
	if err != nil {
		return response, err
	}

	// NOTE: If it is a revoked invite, the user must have the permission to revoke invites
	if !result.MembershipStatus.Role.Can(rbac.PermRevokeWorkspaceInvite) &&
		request.Status == models.InviteStatusRevoked {
		return response, apperrors.Forbidden(
			"you do not have permission to revoke invites for this workspace",
		)
	}

	invite, err := w.repos.WorkspaceRepository().FindInvite(repository.FindInviteArgs{
		InviteID: inviteID,
	})
	if err != nil {
		return response, err
	}

	// If the invite status is already the same as the request status, quit early
	if invite.Status == request.Status {
		response.Workspace = result.Workspace
		return response, nil
	}

	// NOTE: The invite must be pending for it to be updated by any user
	if invite.Status != models.InviteStatusPending {
		return response, apperrors.BadRequest(
			"unable to update invite status, the invite is not pending",
		)
	}

	// NOTE: If the target status is accepted or declined, the user must be the one invited
	isAccepted, isDeclined := request.Status == models.InviteStatusAccepted, request.Status == models.InviteStatusDeclined
	if invite.Invited.UserID != user.ID && (isAccepted || isDeclined) {
		return response, repository.ErrInviteForAnotherEmail
	}

	if err := w.repos.WorkspaceRepository().ValidateInvite(invite); err != nil {
		return response, err
	}

	if err := w.repos.WorkspaceRepository().UpdateInviteStatus(repository.UpdateInviteStatusArgs{
		Invite:      invite,
		Status:      request.Status,
		InvitedUser: *user,
	}); err != nil {
		return response, err
	}

	log.Info().
		Int32("user_id", user.ID).
		Str("invite_id", invite.InviteID.String()).
		Str("status", string(request.Status)).
		Msg("invite status updated")

	go func() {
		if err := w.repos.WorkspaceRepository().
			UntrackInvite(invite.Workspace.ID, invite.Invited.Email, invite.InviteID); err != nil {
			log.Error().
				Err(err).
				Int32("user_id", user.ID).
				Str("invite_id", invite.InviteID.String()).
				Msg("failed to untrack invite")
		}
	}()

	response.Status = request.Status
	if request.Status == models.InviteStatusAccepted {
		response.Workspace = result.Workspace
	}
	return response, nil
}

// FindInvite implements WorkspaceHandler.
func (w *workspaceHandler) FindInvite(
	ctx *robin.Context,
	request FindInviteRequest,
) (FindInviteResponse, error) {
	var response FindInviteResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return FindInviteResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return FindInviteResponse{}, err
	}

	inviteID, err := lib.UUIDFromString(request.InviteID)
	if err != nil {
		return FindInviteResponse{}, err
	}

	invite, err := w.repos.WorkspaceRepository().FindInvite(repository.FindInviteArgs{
		InviteID: inviteID,
	})
	if err != nil {
		return FindInviteResponse{}, err
	}

	if invite.Invited.UserID != auth.UserID {
		return FindInviteResponse{}, repository.ErrInviteForAnotherEmail
	}

	if err := w.repos.WorkspaceRepository().ValidateInvite(invite); err != nil {
		return FindInviteResponse{}, err
	}

	response.Invite = invite
	return response, nil
}

// ListMembers implements WorkspaceHandler.
func (w *workspaceHandler) ListMembers(
	ctx *robin.Context,
	request ListMembersRequest,
) (ListMembersResponse, error) {
	user, err := authlib.ExtractUser(ctx, w.repos.UserRepository())
	if err != nil {
		return ListMembersResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return ListMembersResponse{}, err
	}

	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return ListMembersResponse{}, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID},
		user.ID,
	)
	if err != nil {
		return ListMembersResponse{}, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermListWorkspaceMembers) {
		return ListMembersResponse{}, apperrors.Forbidden(
			"you do not have permission to list members of this workspace",
		)
	}

	list, err := w.repos.WorkspaceRepository().FindMembers(workspaceID, request.Pagination)
	if err != nil {
		return ListMembersResponse{}, err
	}

	// Check if the user has permission to view members' emails
	if !result.MembershipStatus.Role.Can(rbac.PermViewMembersEmail) {
		for i := range list.Members {
			if list.Members[i].User.ID == user.ID {
				continue
			}

			list.Members[i].Email = lib.RedactEmail(list.Members[i].Email, 2)
		}
	}

	return ListMembersResponse{
		Members: list.Members,
		Pagination: request.Pagination.ToState(repository.PageStateArgs{
			CurrentCount: len(list.Members),
			TotalCount:   list.Total,
		}),
	}, nil
}

// InviteUser implements WorkspaceHandler.
func (w *workspaceHandler) InviteUsers(
	ctx *robin.Context,
	request InviteUsersRequest,
) (InviteUsersResponse, error) {
	response := InviteUsersResponse{
		InvitedUsers: make([]string, 0),
		Errors:       make(map[string]string),
		Message:      "",
	}

	user, err := authlib.ExtractUser(ctx, w.repos.UserRepository())
	if err != nil {
		return InviteUsersResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return InviteUsersResponse{}, err
	}

	// Find the workspace
	workspaceID, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return InviteUsersResponse{}, err
	}

	result, err := w.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceID},
		user.ID,
	)
	if err != nil {
		return InviteUsersResponse{}, err
	}

	// Check if the user has permission to invite users
	if !result.MembershipStatus.Role.Can(rbac.PermInviteUsersToWorkspace) {
		return InviteUsersResponse{}, apperrors.Forbidden(
			"you do not have permission to invite users to this workspace",
		)
	}

	var wg sync.WaitGroup
	emails := lib.UniqueSlice(request.Emails)
	for _, email := range emails {
		if email == user.Email {
			response.Errors[email] = "you cannot invite yourself"
			continue
		}

		wg.Add(1)
		go func(email string) {
			defer wg.Done()

			if err := w.inviteUser(email, user, result); err != nil {
				message := "An unexpected error occurred"
				if e, ok := err.(*apperrors.Error); ok {
					message = e.Message()
				}

				response.Errors[email] = message
				return
			}

			response.InvitedUsers = append(response.InvitedUsers, email)
		}(email)
	}
	wg.Wait()

	if len(response.Errors) > 0 {
		log.Error().
			Int32("user_id", user.ID).
			Str("workspace_id", result.Workspace.ID).
			Any("errors", response.Errors).
			Msg("failed to invite users to workspace")

		return response, apperrors.BadRequest(fmt.Sprintf(
			"failed to invite %d %s",
			len(response.Errors),
			lib.Pluralize("user", len(response.Errors)),
		))
	}

	response.Message = fmt.Sprintf(
		"Invited %d %s",
		len(response.InvitedUsers),
		lib.Pluralize("user", len(response.InvitedUsers)),
	)
	return response, nil
}

func (w *workspaceHandler) Create(
	ctx *robin.Context,
	data CreateWorkspaceRequest,
) (CreateWorkspaceResponse, error) {
	var response CreateWorkspaceResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	//nolint:exhaustruct
	workspace := models.Workspace{
		Name:        data.Name,
		Description: data.Description,
		Slug:        strings.TrimSpace(strings.ToLower(data.Slug)),
		OwnerID:     auth.UserID,
	}

	wk, err := w.repos.WorkspaceRepository().Create(&workspace)
	if err != nil {
		return response, err
	}

	// Install core plugins
	go func(workspaceID int32) {
		err = w.pluginManager.InstallCorePlugins(workspaceID)
		if err != nil {
			log.Error().
				Err(err).
				Int32("user_id", auth.UserID).
				Str("workspace_id", wk.ID).
				Msg("failed to install core plugins")
		}
	}(wk.InternalID)

	response.Workspace = wk
	return response, nil
}

func (w *workspaceHandler) Find(
	ctx *robin.Context,
	data FindWorkspaceRequest,
) (FindWorkspaceResponse, error) {
	response := FindWorkspaceResponse{} //nolint:exhaustruct

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return FindWorkspaceResponse{}, err
	}

	result, err := w.repos.WorkspaceRepository().
		FindWithMembershipStatus(repository.PublicIdOrSlug{Slug: data.Slug}, auth.UserID)
	if err != nil {
		return FindWorkspaceResponse{}, err
	}

	// Check if user is a member of the workspace
	if !result.MembershipStatus.IsMember && result.Workspace.InviteOnly {
		return FindWorkspaceResponse{}, apperrors.New(
			"accessed to this workspace is restricted, please request an invite.",
			http.StatusForbidden,
		)
	}

	if result.Workspace == nil {
		return FindWorkspaceResponse{}, apperrors.New(
			"workspace not found",
			http.StatusNotFound,
		)
	}

	// Load collections
	collections, err := w.repos.CollectionRepository().FindByWorkspaceAndUser(
		result.ID,
		auth.UserID,
	)
	if err != nil {
		return FindWorkspaceResponse{}, err
	}

	response.Workspace = result.Workspace
	response.Collections = collections

	return response, nil
}

func (w *workspaceHandler) inviteUser(
	email string,
	user *models.User,
	result *models.WorkspaceWithMembershipStatus,
) error {
	// Check if the user is already a member
	isMember, err := w.repos.WorkspaceRepository().MemberExists(
		result.ID,
		repository.MemberLookupArgs{Email: email}, //nolint:exhaustruct
	)
	if err != nil {
		return err
	}
	if isMember {
		return apperrors.New(
			"user is already a member of this workspace",
			http.StatusConflict,
		)
	}

	// Upsert the invite
	inviteID, err := w.repos.WorkspaceRepository().UpsertWorkspaceInvite(
		repository.UpsertWorkspaceInviteParams{
			WorkspaceID: result.ID,
			Email:       email,
			Role:        rbac.RoleUser,
			InitiatorID: user.ID,
		},
	)
	if err != nil {
		return err
	}

	// Send the invite email
	if err := w.mailer.Send(email, templates.TemplateInviteUserToWorkspace, mail.WorkspaceInviteEmailParams{
		Email:         email,
		InitiatorName: fmt.Sprintf("%s %s.", user.FirstName, user.LastName[:1]),
		WorkspaceName: result.Workspace.Name,
		Host:          w.config.AppUrl,
		InviteID:      inviteID.String(),
		SentAt:        lib.ToHumanReadableDate(time.Now()),
	}); err != nil {
		log.Error().Err(err).Msg("failed to send workspace invite email")
		return apperrors.New(
			"failed to send workspace invite email",
			http.StatusInternalServerError,
		)
	}

	// Only cache the invite if the email was sent successfully
	if err := w.repos.WorkspaceRepository().TrackInviteSent(result.ID, email, inviteID); err != nil {
		log.Error().Err(err).Msg("failed to cache workspace invite")
	}

	return nil
}

var _ WorkspaceHandler = (*workspaceHandler)(nil)
