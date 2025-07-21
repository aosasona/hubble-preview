package api

import (
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/robin"
)

type CollectionHandler interface {
	// Create a new collection
	Create(ctx *robin.Context, data CreateCollectionRequest) (CreateCollectionResponse, error)

	// Update a collection's details
	Update(ctx *robin.Context, request UpdateCollectionRequest) (UpdateCollectionResponse, error)

	Delete(ctx *robin.Context, request DeleteCollectionRequest) (models.Workspace, error)

	// LoadCollectionDetails loads the details of a collectiona and the current member
	LoadCollectionDetails(
		ctx *robin.Context,
		data LoadCollectionDetailsRequest,
	) (LoadCollectionDetailsResponse, error)

	// ListMembers lists all members of a collection
	ListMembers(
		ctx *robin.Context,
		request ListCollectionMembersRequest,
	) (ListMembersResponse, error)

	// AddMembers adds members to a collection
	AddMembers(
		ctx *robin.Context,
		request AddCollectionMembersRequest,
	) (AddCollectionMembersResponse, error)

	// RemoveMembers removes members from a collection
	RemoveMembers(
		ctx *robin.Context,
		request RemoveCollectionMembersRequest,
	) (RemoveCollectionMembersResponse, error)

	// Leave removes the current user from a collection
	Leave(ctx *robin.Context, request LeaveCollectionRequest) (workspaceSlug string, err error)
}

type (
	CreateCollectionRequest struct {
		// ID of the workspace to create the collection in
		WorkspaceSlug string `json:"workspace_slug"     validate:"required,slug"`
		// Name of the collection
		Name string `json:"name"               validate:"required,mixed_name,min=1,max=64"`
		// Slug of the collection
		Slug string `json:"slug"               validate:"optional_slug,max=80"             mirror:"optional:true"`
		// Description of the collection
		Description string `json:"description"        validate:"ascii,min=0,max=512"              mirror:"optional:true"`
		// Whether to assign all members of the workspace to this collection by default
		AssignAllMembers bool `json:"assign_all_members"`
	}

	CreateCollectionResponse struct {
		WorkspaceSlug string             `json:"workspace_slug"`
		Collection    *models.Collection `json:"collection"     mirror:"optional:false"`
	}

	UpdateCollectionRequest struct {
		Name        string `json:"name"        validate:"optional_mixed_name,max=64"`
		Slug        string `json:"slug"        validate:"optional_slug,max=80"`
		Description string `json:"description" validate:"ascii,min=0,max=512"        mirror:"optional:true"`

		CollectionID string `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string `json:"workspace_id"  validate:"required,uuid"`
	}

	UpdateCollectionResponse struct {
		WorkspaceSlug string            `json:"workspace_slug"`
		Collection    models.Collection `json:"collection"     mirror:"optional:false"`
	}

	LoadCollectionDetailsRequest struct {
		CollectionSlug string `json:"collection_slug" validate:"required,slug"`
		WorkspaceSlug  string `json:"workspace_slug"  validate:"required,slug"`
	}

	LoadCollectionDetailsResponse struct {
		Collection       models.Collection       `json:"collection"`
		MembershipStatus models.MembershipStatus `json:"membership_status"`
		Workspace        models.Workspace        `json:"workspace"`
	}

	ListCollectionMembersRequest struct {
		CollectionID string                      `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string                      `json:"workspace_id"  validate:"required,uuid"`
		Pagination   repository.PaginationParams `json:"pagination"`
	}

	AddCollectionMembersRequest struct {
		CollectionID string   `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string   `json:"workspace_id"  validate:"required,uuid"`
		Emails       []string `json:"emails"        validate:"required,dive,email"`
	}

	AddCollectionMembersResponse struct {
		AddedCount int `json:"added_count"`
	}

	RemoveCollectionMembersRequest struct {
		CollectionID string   `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string   `json:"workspace_id"  validate:"required,uuid"`
		Emails       []string `json:"emails"        validate:"required,dive,email"`
	}

	RemoveCollectionMembersResponse struct {
		RemovedCount int `json:"removed_count"`
	}

	LeaveCollectionRequest struct {
		CollectionID string `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string `json:"workspace_id"  validate:"required,uuid"`
	}

	DeleteCollectionRequest struct {
		CollectionID string `json:"collection_id" validate:"required,uuid"`
		WorkspaceID  string `json:"workspace_id"  validate:"required,uuid"`
	}
)
