package api

import (
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/robin"
)

type collectionHandler struct {
	*baseHandler
}

var ErrNotCollectionMember = apperrors.BadRequest(
	"Oops! You are not a member of this collection",
)

// Delete implements CollectionHandler.
func (c *collectionHandler) Delete(
	ctx *robin.Context,
	request DeleteCollectionRequest,
) (models.Workspace, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return models.Workspace{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return models.Workspace{}, err
	}

	var workspaceID, collectionID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return models.Workspace{}, err
	}
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return models.Workspace{}, err
	}

	//nolint:exhaustruct
	result, err := c.repos.CollectionRepository().
		FindWithMembershipStatus(&repository.FindWithMembershipStatusArgs{
			UserID:       auth.UserID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		})
	if err != nil {
		return models.Workspace{}, err
	}

	if result.Workspace == nil {
		return models.Workspace{}, apperrors.BadRequest("workspace not found")
	}

	if !result.MembershipStatus.Role.Can(rbac.PermDeleteCollection) {
		return models.Workspace{}, apperrors.New(
			"you do not have permission to delete this collection",
			http.StatusForbidden,
		)
	}

	if err := c.repos.CollectionRepository().Delete(result.Collection.InternalID); err != nil {
		return models.Workspace{}, err
	}

	return *result.Workspace, nil
}

// Leave implements CollectionHandler.
func (c *collectionHandler) Leave(
	ctx *robin.Context,
	request LeaveCollectionRequest,
) (workspaceSlug string, err error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return "", err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return "", err
	}

	var workspaceID, collectionID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return "", err
	}
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return "", err
	}

	//nolint:exhaustruct
	result, err := c.repos.CollectionRepository().
		FindWithMembershipStatus(&repository.FindWithMembershipStatusArgs{
			UserID:       auth.UserID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		})
	if err != nil {
		return "", err
	}

	if !result.MembershipStatus.IsMember {
		return "", ErrNotCollectionMember
	}

	if result.Collection.OwnerID == auth.UserID {
		return "", apperrors.BadRequest("you need to transfer ownership before leaving")
	}

	if err := c.repos.CollectionRepository().Leave(result.Collection.InternalID, result.MembershipStatus.UserID); err != nil {
		return "", err
	}

	return result.Workspace.Slug, nil
}

// RemoveMembers implements CollectionHandler.
func (c *collectionHandler) RemoveMembers(
	ctx *robin.Context,
	request RemoveCollectionMembersRequest,
) (RemoveCollectionMembersResponse, error) {
	var response RemoveCollectionMembersResponse

	user, err := authlib.ExtractUser(ctx, c.repos.UserRepository())
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	var workspaceID, collectionID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return response, err
	}
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return response, err
	}

	//nolint:exhaustruct
	result, err := c.repos.CollectionRepository().
		FindWithMembershipStatus(&repository.FindWithMembershipStatusArgs{
			UserID:       user.ID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		})
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermRemoveMemberFromCollection) {
		return response, apperrors.New(
			"you do not have permission to remove members from this collection",
			http.StatusForbidden,
		)
	}

	// Check if the user is trying to remove themselves
	if slices.Contains(request.Emails, user.Email) {
		return response, apperrors.BadRequest("you cannot remove yourself from the collection here")
	}

	isOwner := result.MembershipStatus.Role == rbac.RoleOwner
	removedCount, err := c.repos.CollectionRepository().
		RemoveMembers(repository.RemoveCollectionMembersArgs{
			InitiatorIsOwner: isOwner,
			WorkspaceID:      result.Workspace.InternalID,
			CollectionID:     result.Collection.InternalID,
			Emails:           request.Emails,
		})
	if err != nil {
		return response, err
	}

	if removedCount == 0 {
		return response, apperrors.BadRequest("no members were removed")
	}

	response.RemovedCount = removedCount
	return response, nil
}

// AddMembers implements CollectionHandler.
func (c *collectionHandler) AddMembers(
	ctx *robin.Context,
	request AddCollectionMembersRequest,
) (AddCollectionMembersResponse, error) {
	var response AddCollectionMembersResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	var workspaceID, collectionID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return response, err
	}
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return response, err
	}

	//nolint:exhaustruct
	result, err := c.repos.CollectionRepository().
		FindWithMembershipStatus(&repository.FindWithMembershipStatusArgs{
			UserID:       auth.UserID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		})
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermAddMemberToCollection) {
		return response, apperrors.New(
			"you do not have permission to add members to this collection",
			http.StatusForbidden,
		)
	}

	added, err := c.repos.CollectionRepository().AddMembers(repository.AddCollectionMembersArgs{
		WorkspaceID:  workspaceID,
		CollectionID: collectionID,
		Emails:       request.Emails,
	})
	if err != nil {
		return response, err
	}

	if len(added) == 0 {
		return response, apperrors.New(
			"no members were added",
			http.StatusBadRequest,
		)
	}

	response.AddedCount = len(added)
	return response, nil
}

// ListMembers implements CollectionHandler.
func (c *collectionHandler) ListMembers(
	ctx *robin.Context,
	request ListCollectionMembersRequest,
) (ListMembersResponse, error) {
	user, err := authlib.ExtractUser(ctx, c.repos.UserRepository())
	if err != nil {
		return ListMembersResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return ListMembersResponse{}, err
	}

	var workspaceID, collectionID pgtype.UUID
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return ListMembersResponse{}, err
	}
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return ListMembersResponse{}, err
	}

	result, err := c.repos.CollectionRepository().FindWithMembershipStatus(
		//nolint:exhaustruct
		&repository.FindWithMembershipStatusArgs{
			UserID:       user.ID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		},
	)
	if err != nil {
		return ListMembersResponse{}, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermListCollectionMembers) {
		return ListMembersResponse{}, apperrors.Forbidden(
			"you do not have permission to list members of this collection",
		)
	}

	//nolint:exhaustruct
	list, err := c.repos.CollectionRepository().FindMembers(repository.FindMembersArgs{
		WorkspaceID:  workspaceID,
		CollectionID: collectionID,
		Pagination:   request.Pagination,
	})
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

// Update implements CollectionHandler.
func (c *collectionHandler) Update(
	ctx *robin.Context,
	request UpdateCollectionRequest,
) (UpdateCollectionResponse, error) {
	var response UpdateCollectionResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	var collectionID, workspaceID pgtype.UUID
	if collectionID, err = lib.UUIDFromString(request.CollectionID); err != nil {
		return response, err
	}
	if workspaceID, err = lib.UUIDFromString(request.WorkspaceID); err != nil {
		return response, err
	}

	result, err := c.repos.CollectionRepository().FindWithMembershipStatus(
		//nolint:exhaustruct
		&repository.FindWithMembershipStatusArgs{
			UserID:       auth.UserID,
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		},
	)
	if err != nil {
		return response, err
	}

	// Check name
	if request.Name != "" && request.Name != result.Collection.Name {
		// Ensure the user can update the collection name
		if !result.MembershipStatus.Role.Can(rbac.PermRenameCollection) {
			return response, apperrors.NewValidationError(apperrors.ErrorMap{
				"name": {"Insufficient permissions to change collection name"},
			})
		}

		// Check if the name is already taken
		log.Debug().Msgf("Checking if collection name %s exists", request.Name)
		if exists, _ := c.repos.CollectionRepository().
			NameExists(result.Workspace.InternalID, request.Name); exists {
			return response, apperrors.NewValidationError(apperrors.ErrorMap{
				"name": {"Another collection with this name already exists"},
			})
		}

		request.Name = strings.TrimSpace(request.Name)
	}

	// Check slug
	if request.Slug != "" && request.Slug != result.Collection.Slug {
		if !result.MembershipStatus.Role.Can(rbac.PermChangeCollectionSlug) {
			return response, apperrors.NewValidationError(apperrors.ErrorMap{
				"slug": {"Insufficient permissions to change collection slug"},
			})
		}

		if slugExists, _ := c.repos.CollectionRepository().SlugExists(result.Workspace.InternalID, request.Slug); slugExists {
			return response, apperrors.NewValidationError(apperrors.ErrorMap{
				"slug": {"Another collection with this slug already exists"},
			})
		}

		request.Slug = strings.TrimSpace(strings.ToLower(request.Slug))
	}

	// Check description
	if request.Description != "" && request.Description != result.Collection.Description {
		if !result.MembershipStatus.Role.Can(rbac.PermChangeCollectionDescription) {
			return response, apperrors.NewValidationError(apperrors.ErrorMap{
				"description": {"Insufficient permissions to change collection description"},
			})
		}

		request.Description = strings.TrimSpace(request.Description)
	}

	// Update collection
	updated, err := c.repos.CollectionRepository().
		UpdateDetails(repository.UpdateCollectionDetailsArgs{
			Name:         request.Name,
			Description:  request.Description,
			Slug:         request.Slug,
			CollectionID: result.Collection.InternalID,
		})
	if err != nil {
		return response, err
	}

	return UpdateCollectionResponse{
		WorkspaceSlug: result.Workspace.Slug,
		Collection:    updated,
	}, nil
}

// LoadCollectionDetails implements CollectionHandler.
func (c *collectionHandler) LoadCollectionDetails(
	ctx *robin.Context,
	request LoadCollectionDetailsRequest,
) (LoadCollectionDetailsResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return LoadCollectionDetailsResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return LoadCollectionDetailsResponse{}, err
	}

	//nolint:exhaustruct
	result, err := c.repos.CollectionRepository().FindWithMembershipStatus(
		&repository.FindWithMembershipStatusArgs{
			UserID:         auth.UserID,
			WorkspaceSlug:  request.WorkspaceSlug,
			CollectionSlug: request.CollectionSlug,
		},
	)
	if err != nil {
		return LoadCollectionDetailsResponse{}, err
	}

	if !result.MembershipStatus.IsMember {
		return LoadCollectionDetailsResponse{}, ErrNotCollectionMember
	}

	if result.Collection == nil {
		return LoadCollectionDetailsResponse{}, apperrors.BadRequest("collection not found")
	}
	if result.MembershipStatus == nil {
		return LoadCollectionDetailsResponse{}, apperrors.BadRequest("workspace not found")
	}
	if result.Workspace == nil {
		return LoadCollectionDetailsResponse{}, apperrors.BadRequest("workspace not found")
	}

	return LoadCollectionDetailsResponse{
		Collection:       *result.Collection,
		MembershipStatus: *result.MembershipStatus,
		Workspace:        *result.Workspace,
	}, nil
}

func (c *collectionHandler) Create(
	ctx *robin.Context,
	data CreateCollectionRequest,
) (CreateCollectionResponse, error) {
	var response CreateCollectionResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	// Check if workspace exists
	var workspace *models.WorkspaceWithMembershipStatus
	if workspace, err = c.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{Slug: data.WorkspaceSlug}, //nolint:exhaustruct
		auth.UserID,
	); err != nil {
		return response, err
	}

	// Check if user is a member of the workspace
	if !workspace.IsValid() || !workspace.IsMember() {
		return CreateCollectionResponse{}, ErrNotCollectionMember
	}

	// Check if user can create collection in workspace
	if !workspace.MembershipStatus.Role.Can(rbac.PermCreateCollection) {
		return CreateCollectionResponse{}, apperrors.New(
			"you do not have permission to create a collection in this workspace",
			http.StatusForbidden,
		)
	}

	// Create collection
	//nolint:exhaustruct
	newCollection := models.Collection{
		Name:        lib.ToTitleCase(strings.TrimSpace(data.Name)),
		WorkspaceID: workspace.ID,
		Slug:        strings.TrimSpace(strings.ToLower(data.Slug)),
		Description: data.Description,
	}
	collection, err := c.repos.CollectionRepository().Create(repository.CreateCollectionParams{
		Collection: &newCollection,
		Owner: struct {
			UserID int32
		}{UserID: auth.UserID},
		AddAllMembers: &data.AssignAllMembers,
	})
	if err != nil {
		return response, err
	}

	response.WorkspaceSlug = workspace.Workspace.Slug
	response.Collection = collection
	return response, nil
}

var _ CollectionHandler = (*collectionHandler)(nil)
