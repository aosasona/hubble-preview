package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/pkg/rbac"
)

type InviteStatus string

const (
	InviteStatusAccepted InviteStatus = "accepted"
	InviteStatusPending  InviteStatus = "pending"
	InviteStatusDeclined InviteStatus = "declined"
	InviteStatusRevoked  InviteStatus = "revoked"
	InviteStatusExpired  InviteStatus = "expired"
)

type (
	MemberWithUserID struct {
		MemberID int32 `json:"member_id"`
		UserID   int32 `json:"user_id"`
	}

	Workspace struct {
		InternalID           int32     `json:"-"`
		ID                   string    `json:"id"`
		Name                 string    `json:"name"`
		Slug                 string    `json:"slug"`
		OwnerID              int32     `json:"owner_id"`
		Description          string    `json:"description"`
		AvatarID             string    `json:"avatar_id"`
		EnablePublicIndexing bool      `json:"enable_public_indexing"`
		InviteOnly           bool      `json:"invite_only"`
		CreatedAt            time.Time `json:"created_at"`

		Collections []Collection `json:"collections,omitempty" mirror:"optional:false"`
	}

	MembershipStatus struct {
		MemberID int32     `json:"member_id"`
		UserID   int32     `json:"user_id"`
		Role     rbac.Role `json:"role"      mirror:"type:'user' | 'admin' | 'guest' | 'owner'"`

		IsMember bool `json:"is_member"`
	}

	WorkspaceWithMembershipStatus struct {
		ID               int32
		PublicID         pgtype.UUID
		Workspace        *Workspace
		MembershipStatus *MembershipStatus
	}

	MemberUser struct {
		ID       int32       `json:"-"`
		PublicID pgtype.UUID `json:"id" mirror:"type:string"`
	}

	Member struct {
		// ID can be a workspace_member or workspace_invites ID
		ID        int32      `json:"id"`
		InviteID  string     `json:"invite_id"  mirror:"optional:true"`
		FirstName string     `json:"first_name"`
		LastName  string     `json:"last_name"`
		Email     string     `json:"email"`
		Role      rbac.Role  `json:"role"       mirror:"type:'user' | 'admin' | 'guest' | 'owner'"`
		User      MemberUser `json:"user"`
		// Status is the invitation status, ideally this will be pending or declined since others are excluded by default
		Status    InviteStatus `json:"status"     mirror:"type:'accepted' | 'pending' | 'declined' | 'revoked' | 'expired'"`
		CreatedAt time.Time    `json:"created_at"`
	}

	InviteWorkspaceDetails struct {
		ID       int32       `json:"-"`
		PublicID pgtype.UUID `json:"id"   mirror:"type:string"`
		Name     string      `json:"name"`
		Slug     string      `json:"slug"`
	}

	InvitedDetails struct {
		UserID int32  `json:"-"`
		Email  string `json:"email"`
		Exists bool   `json:"-"`
	}

	WorkspaceInvite struct {
		ID        int32                  `json:"-"`
		InviteID  pgtype.UUID            `json:"id"         mirror:"type:string"`
		Invited   InvitedDetails         `json:"invited"`
		Workspace InviteWorkspaceDetails `json:"workspace"`
		Inviter   User                   `json:"inviter"`
		Status    InviteStatus           `json:"status"     mirror:"type:'accepted' | 'pending' | 'declined' | 'revoked' | 'expired'"`
		Role      rbac.Role              `json:"role"       mirror:"type:'user' | 'admin' | 'guest' | 'owner'"`
		InvitedAt time.Time              `json:"invited_at"`
	}
)

func (w *Workspace) From(workspace *queries.Workspace) *Workspace {
	*w = Workspace{
		InternalID:           workspace.ID,
		ID:                   workspace.PublicID.String(),
		Name:                 workspace.DisplayName,
		Slug:                 workspace.Slug.String,
		OwnerID:              workspace.OwnerID,
		Description:          workspace.Description.String,
		AvatarID:             workspace.AvatarID.String,
		EnablePublicIndexing: workspace.EnablePublicIndexing,
		InviteOnly:           workspace.InviteOnly,
		CreatedAt:            workspace.CreatedAt.Time,
	}

	return w
}

func (m WorkspaceWithMembershipStatus) IsValid() bool {
	return m.Workspace != nil && m.MembershipStatus != nil
}

func (m WorkspaceWithMembershipStatus) IsMember() bool {
	return m.MembershipStatus != nil && m.MembershipStatus.IsMember
}
