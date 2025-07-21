package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/pkg/rbac"
)

type (
	Collection struct {
		InternalID  int32     `json:"-"`
		ID          string    `json:"id"`
		Name        string    `json:"name"`
		Slug        string    `json:"slug"`
		WorkspaceID int32     `json:"workspace_id"`
		Description string    `json:"description"`
		AvatarID    string    `json:"avatar_id"`
		CreatedAt   time.Time `json:"created_at"`
		OwnerID     int32     `json:"owner_id"`

		MembersCount int64 `json:"members_count"`
		EntriesCount int64 `json:"entries_count"`
	}

	CollectionMember struct {
		ID               int32     `json:"member_id"`
		CollectionID     int32     `json:"collection_id"`
		UserID           int32     `json:"user_id"`
		ExtraPermissions []byte    `json:"-"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
		DeletedAt        time.Time `json:"deleted_at"`
		Role             rbac.Role `json:"role"`
	}

	CollectionWithMembershipStatus struct {
		ID               int32
		PublicID         pgtype.UUID
		Collection       *Collection
		Workspace        *Workspace
		MembershipStatus *MembershipStatus
	}
)

func (c *Collection) From(collection *queries.Collection) *Collection {
	//nolint:exhaustruct
	*c = Collection{
		InternalID:  collection.ID,
		ID:          collection.PublicID.String(),
		Name:        collection.Name,
		Slug:        collection.Slug.String,
		WorkspaceID: collection.WorkspaceID,
		Description: collection.Description.String,
		AvatarID:    collection.AvatarID.String,
		CreatedAt:   collection.CreatedAt.Time,
		OwnerID:     collection.OwnerID,
	}

	return c
}

func (cm *CollectionMember) From(member *queries.CollectionMember) *CollectionMember {
	*cm = CollectionMember{
		ID:               member.ID,
		CollectionID:     member.CollectionID,
		UserID:           member.UserID,
		Role:             member.BitmaskRole,
		CreatedAt:        member.CreatedAt.Time,
		UpdatedAt:        member.UpdatedAt.Time,
		DeletedAt:        member.DeletedAt.Time,
		ExtraPermissions: member.ExtraPermissions,
	}

	return cm
}
