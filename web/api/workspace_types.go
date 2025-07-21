package api

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/robin"
)

type WorkspaceHandler interface {
	// Create a new workspace
	Create(ctx *robin.Context, request CreateWorkspaceRequest) (CreateWorkspaceResponse, error)

	// Find a workspace by slug
	Find(ctx *robin.Context, filters FindWorkspaceRequest) (FindWorkspaceResponse, error)

	// Update a workspace
	Update(ctx *robin.Context, request UpdateWorkspaceRequest) (CreateWorkspaceResponse, error)

	// Delete a workspace
	Delete(ctx *robin.Context, request DeleteWorkspaceRequest) (DeleteWorkspaceResponse, error)

	// Load the membership status of the current user in a workspace
	LoadMemberStatus(
		ctx *robin.Context,
		request LoadMemberStatusRequest,
	) (LoadMemberStatusResponse, error)

	// Invite users to a workspace
	InviteUsers(ctx *robin.Context, request InviteUsersRequest) (InviteUsersResponse, error)

	// FindInvite finds an invite by ID and the current user
	// NOTE: To accept an invite, the user must be authenticated to the same email address
	FindInvite(ctx *robin.Context, request FindInviteRequest) (FindInviteResponse, error)

	UpdateInviteStatus(
		ctx *robin.Context,
		request UpdateInviteStatusRequest,
	) (UpdateInviteStatusResponse, error)

	// ListMembers lists all members of a workspace
	ListMembers(ctx *robin.Context, request ListMembersRequest) (ListMembersResponse, error)

	RemoveMember(ctx *robin.Context, request RemoveMemberRequest) (RemoveMemberResponse, error)

	ChangeMemberRole(
		ctx *robin.Context,
		request ChangeMemberRoleRequest,
	) (ChangeMemberRoleResponse, error)
}

type (
	CreateWorkspaceRequest struct {
		Name        string `json:"name"        validate:"required,workspace_name,min=2,max=64"`
		Slug        string `json:"slug"        validate:"optional_slug,max=80"`
		Description string `json:"description" validate:"ascii,min=0,max=512"                  mirror:"optional:true"`
	}

	UpdateWorkspaceRequest struct {
		Name        string `json:"name"         validate:"optional_workspace_name"`
		Slug        string `json:"slug"         validate:"optional_slug,max=80"`
		Description string `json:"description"  validate:"ascii,min=0,max=512"     mirror:"optional:true"`
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	}

	CreateWorkspaceResponse struct {
		Workspace *models.Workspace `json:"workspace" mirror:"optional:false"`
	}

	FindWorkspaceRequest struct {
		Slug string `json:"slug" validate:"required,slug"`
	}

	FindWorkspaceResponse struct {
		Workspace   *models.Workspace   `json:"workspace"   mirror:"optional:false"`
		Collections []models.Collection `json:"collections" mirror:"type:import('./types').Collection[]"`
	}

	InviteUsersRequest struct {
		Emails      []string `json:"emails"       validate:"required,min=1,max=50,dive,email"`
		WorkspaceID string   `json:"workspace_id" validate:"required,uuid"`
	}

	InviteUsersResponse struct {
		Message      string            `json:"message"`
		InvitedUsers []string          `json:"invited_users"`
		Errors       map[string]string `json:"errors"`
	}

	ChangeMemberRoleRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"                         mirror:"type:string"`
		UserID      string `json:"user_id"      validate:"required,uuid"                         mirror:"type:string"`
		Role        string `json:"role"         validate:"required,oneof=guest user admin owner" mirror:"type:'user' | 'admin' | 'owner' | 'guest'"`
	}

	ChangeMemberRoleResponse struct {
		UserID  pgtype.UUID `json:"user_id" mirror:"type:string"`
		NewRole rbac.Role   `json:"role"    mirror:"type:'user' | 'admin' | 'owner' | 'guest'" validate:"required,oneof=user admin"`
	}

	ListMembersRequest struct {
		WorkspaceID string                      `json:"workspace_id" validate:"required,uuid"`
		Pagination  repository.PaginationParams `json:"pagination"`
	}

	ListMembersResponse struct {
		Members    []models.Member            `json:"members"`
		Pagination repository.PaginationState `json:"pagination"`
	}

	FindInviteRequest struct {
		InviteID string `json:"invite_id" validate:"required,uuid"`
	}

	FindInviteResponse struct {
		Invite models.WorkspaceInvite `json:"invite"`
	}

	UpdateInviteStatusRequest struct {
		WorkspaceID string              `json:"workspace_id" validate:"required,uuid"`
		InviteID    string              `json:"invite_id"    validate:"required,uuid"`
		Status      models.InviteStatus `json:"status"       validate:"required,oneof=accepted declined revoked" mirror:"type:'accepted' | 'declined' | 'revoked'"`
	}

	UpdateInviteStatusResponse struct {
		Status    models.InviteStatus `json:"status"    mirror:"type:'accepted' | 'declined' | 'revoked'"`
		Workspace *models.Workspace   `json:"workspace"`
	}

	LoadMemberStatusRequest struct {
		WorkspaceSlug string `json:"workspace_slug" validate:"required,slug"`
	}

	LoadMemberStatusResponse struct {
		Workspace models.Workspace        `json:"workspace"`
		Status    models.MembershipStatus `json:"status"`
	}

	RemoveMemberRequest struct {
		Email       string `json:"email"        validate:"required,email"`
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	}

	RemoveMemberResponse struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		MemberID  int32  `json:"member_id"`
	}

	DeleteWorkspaceRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	}

	DeleteWorkspaceResponse struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	}
)
