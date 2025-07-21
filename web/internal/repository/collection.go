package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/models"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/seer"
)

var ErrCollectionNotFound = apperrors.BadRequest(
	"this collection does not exist in this workspace or you don't have access to it",
)

type (
	CreateCollectionParams struct {
		Collection *models.Collection
		Owner      struct {
			UserID int32
		}
		AddAllMembers *bool
	}

	GetInternalIDParams struct {
		WorkspaceID  pgtype.UUID
		CollectionID pgtype.UUID
	}

	FindEntriesByCollectionSlugParams struct {
		WorkspaceID  string
		CollectionID string
	}

	FindWithMembershipStatusArgs struct {
		UserID int32

		WorkspaceID   pgtype.UUID
		WorkspaceSlug string

		CollectionID   pgtype.UUID
		CollectionSlug string
	}

	UpdateCollectionDetailsArgs struct {
		Name         string
		Description  string
		Slug         string
		CollectionID int32
	}

	FindMembersArgs struct {
		WorkspaceID   pgtype.UUID
		WorkspaceSlug string

		CollectionID   pgtype.UUID
		CollectionSlug string

		Pagination PaginationParams
	}

	AddCollectionMembersArgs struct {
		WorkspaceID  pgtype.UUID
		CollectionID pgtype.UUID
		Emails       []string
	}

	RemoveCollectionMembersArgs struct {
		InitiatorIsOwner bool
		WorkspaceID      int32
		CollectionID     int32
		Emails           []string
	}

	CollectionRepository interface {
		// Create a new collection
		Create(params CreateCollectionParams) (*models.Collection, error)

		// Delete a collection
		Delete(collectionID int32) error

		// UpdateDetails updates the details of a collection
		UpdateDetails(args UpdateCollectionDetailsArgs) (models.Collection, error)

		// Leave self-removes a user from a collection
		Leave(collectionID, userID int32) error

		// Exists checks if a collection exists in a workspace
		NameExists(workspaceId int32, name string) (bool, error)

		// SlugExists checks if a collection exists in a workspace
		SlugExists(workspaceId int32, slug string) (bool, error)

		// AddMembers adds multiple members by their email to a collection
		AddMembers(args AddCollectionMembersArgs) ([]models.CollectionMember, error)

		// FindMember finds a member by their user ID
		FindMember(
			workspace PublicIdOrSlug,
			collection PublicIdOrSlug,
			userId int32,
		) (models.CollectionMember, error)

		// FindMembers finds all members of a collection
		FindMembers(args FindMembersArgs) (FindMembersResult, error)

		// RemoveMembers removes multiple members by their user ID from a collection
		// NOTE: this is a soft delete, and will NOT allow removing the owner of the collection
		RemoveMembers(RemoveCollectionMembersArgs) (int, error)

		// GetIdBySlug find the internal ID of a collection by its public ID and workspace ID
		GetInternalID(GetInternalIDParams) (int32, error)

		// FindByWorkspaceAndUser finds all collections in a workspace that a user also belongs to
		FindByWorkspaceAndUser(workspaceID int32, userID int32) ([]models.Collection, error)

		FindWithMembershipStatus(
			args *FindWithMembershipStatusArgs,
		) (*models.CollectionWithMembershipStatus, error)
	}

	collectionRepo struct {
		*baseRepo
	}
)

// Delete implements CollectionRepository.
func (c *collectionRepo) Delete(collectionID int32) error {
	err := c.queries.MarkCollectionAsDeleted(context.TODO(), queries.MarkCollectionAsDeletedParams{
		CollectionID: collectionID,
	})

	return err
}

// Leave implements CollectionRepository.
func (c *collectionRepo) Leave(collectionID int32, userID int32) error {
	err := c.queries.LeaveCollection(context.TODO(), queries.LeaveCollectionParams{
		CollectionID: collectionID,
		UserID:       userID,
	})
	if err != nil {
		return err
	}

	return nil
}

// RemoveMembers implements CollectionRepository.
func (c *collectionRepo) RemoveMembers(args RemoveCollectionMembersArgs) (int, error) {
	rows, err := c.queries.RemoveCollectionMembers(
		context.TODO(),
		queries.RemoveCollectionMembersParams{
			Emails:        args.Emails,
			CollectionID:  args.CollectionID,
			WorkspaceID:   args.WorkspaceID,
			IncludeAdmins: args.InitiatorIsOwner,
			AdminRole:     rbac.RoleAdmin,
		},
	)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, apperrors.NewValidationError(apperrors.ErrorMap{
			"emails": {
				"No members removed, these emails are not part of this collection",
			},
		})
	}

	return len(rows), nil
}

// AddMembers implements CollectionRepository.
func (c *collectionRepo) AddMembers(
	args AddCollectionMembersArgs,
) ([]models.CollectionMember, error) {
	rows, err := c.queries.AddMembersToCollection(
		context.TODO(),
		queries.AddMembersToCollectionParams{
			UserRole:           rbac.RoleUser,
			CollectionPublicID: args.CollectionID,
			Emails:             args.Emails,
		}, //nolint:exhaustruct
	)
	if err != nil {
		return nil, err
	}

	members := make([]models.CollectionMember, 0)
	for _, row := range rows {
		member := models.CollectionMember{} //nolint:exhaustruct
		member.From(&row)
		members = append(members, member)
	}

	if len(members) == 0 {
		return nil, apperrors.NewValidationError(apperrors.ErrorMap{
			"emails": {
				"No members added, these emails are already members of this collection or are not part of this workspace",
			},
		})
	}

	return members, nil
}

// FindMembers implements CollectionRepository.
func (c *collectionRepo) FindMembers(args FindMembersArgs) (FindMembersResult, error) {
	rows, err := c.queries.FindCollectionMembers(
		context.TODO(),
		queries.FindCollectionMembersParams{
			Limit:          args.Pagination.Limit(),
			Offset:         args.Pagination.Offset(),
			WorkspaceID:    args.WorkspaceID,
			WorkspaceSlug:  lib.PgText(args.WorkspaceSlug),
			CollectionID:   args.CollectionID,
			CollectionSlug: lib.PgText(args.CollectionSlug),
		},
	)
	if err != nil {
		return FindMembersResult{}, err
	}

	var total int64
	members := make([]models.Member, 0)
	for i := range rows {
		row := &rows[i]
		member := models.Member{
			ID:        row.ID,
			InviteID:  row.InviteID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
			Email:     row.Email,
			Role:      rbac.Role(row.Role),
			Status:    models.InviteStatus(row.Status),
			CreatedAt: row.CreatedAt.Time,
			User: models.MemberUser{
				ID:       row.UserID,
				PublicID: row.PublicUserID,
			},
		}

		if total == 0 {
			total = row.TotalCount
		}

		members = append(members, member)
	}

	return FindMembersResult{
		Members: lib.WithMaxSize(members, args.Pagination.PerPage),
		Total:   total,
	}, nil
}

// UpdateDetails implements CollectionRepository.
func (c *collectionRepo) UpdateDetails(
	args UpdateCollectionDetailsArgs,
) (models.Collection, error) {
	result, err := c.queries.UpdateCollectionDetails(
		context.TODO(),
		queries.UpdateCollectionDetailsParams{
			Name:         lib.PgText(args.Name),
			Description:  lib.PgText(args.Description),
			Slug:         lib.PgText(args.Slug),
			CollectionID: args.CollectionID,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Collection{}, ErrCollectionNotFound
		}
		return models.Collection{}, err
	}

	collection := models.Collection{} //nolint:exhaustruct
	collection.From(&result)
	return collection, nil
}

// FindWithMembershipStatus implements CollectionRepository.
func (c *collectionRepo) FindWithMembershipStatus(
	args *FindWithMembershipStatusArgs,
) (*models.CollectionWithMembershipStatus, error) {
	row, err := c.queries.FindCollectionWithMembershipStatus(
		context.TODO(),
		queries.FindCollectionWithMembershipStatusParams{
			UserID:             args.UserID,
			WorkspacePublicID:  args.WorkspaceID,
			WorkspaceSlug:      lib.PgText(args.WorkspaceSlug),
			CollectionPublicID: args.CollectionID,
			CollectionSlug:     lib.PgText(args.CollectionSlug),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCollectionNotFound
		}

		return nil, err
	}

	collection := &models.Collection{} //nolint:exhaustruct
	collection.From(&row.Collection)

	workspace := &models.Workspace{} //nolint:exhaustruct
	workspace.From(&row.Workspace)

	result := &models.CollectionWithMembershipStatus{
		ID:         row.Collection.ID,
		PublicID:   row.Collection.PublicID,
		Collection: collection,
		Workspace:  workspace,
		MembershipStatus: &models.MembershipStatus{
			MemberID: row.MemberID.Int32,
			UserID:   row.MemberUserID.Int32,
			Role:     rbac.Role(row.Role),
			IsMember: row.IsMember,
		},
	}

	return result, nil
}

// FindMember implements CollectionRepository.
func (c *collectionRepo) FindMember(
	workspace PublicIdOrSlug,
	collection PublicIdOrSlug,
	userId int32,
) (models.CollectionMember, error) {
	result, err := c.queries.FindCollectionMember(
		context.TODO(),
		queries.FindCollectionMemberParams{
			CollectionSlug:     lib.PgText(collection.Slug),
			CollectionPublicID: collection.PublicID,
			WorkspaceSlug:      lib.PgText(workspace.Slug),
			WorkspacePublicID:  workspace.PublicID,
			UserID:             userId,
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.CollectionMember{}, apperrors.New(
				"user is not a member of this collection",
				http.StatusBadRequest)
		}
		return models.CollectionMember{}, err
	}

	m := models.CollectionMember{} //nolint:exhaustruct
	m.From(&result.CollectionMember)
	return m, nil
}

func (c *collectionRepo) GetInternalID(params GetInternalIDParams) (int32, error) {
	return c.queries.FindCollectionIdByPublicId(
		context.TODO(),
		queries.FindCollectionIdByPublicIdParams{
			WorkspaceID: params.WorkspaceID,
			PublicID:    params.CollectionID,
		},
	)
}

func (c *collectionRepo) NameExists(workspaceId int32, collectionName string) (bool, error) {
	return c.queries.CollectionNameExists(context.TODO(), queries.CollectionNameExistsParams{
		WorkspaceID:    workspaceId,
		CollectionName: strings.TrimSpace(collectionName),
	})
}

func (c *collectionRepo) SlugExists(workspaceId int32, collectionSlug string) (bool, error) {
	return c.queries.CollectionSlugExists(context.TODO(), queries.CollectionSlugExistsParams{
		WorkspaceID: workspaceId,
		Slug:        pgtype.Text{String: strings.TrimSpace(collectionSlug), Valid: true},
	})
}

// Create implements CollectionRepository.
func (c *collectionRepo) Create(params CreateCollectionParams) (*models.Collection, error) {
	if params.Collection == nil {
		return nil, seer.Wrap("create_collection", errors.New("collection is required"))
	}

	if params.AddAllMembers == nil {
		return nil, seer.Wrap("create_collection", errors.New("AddAllMembers is required"))
	}

	// Check if collection name already exists
	exists, err := c.NameExists(params.Collection.WorkspaceID, params.Collection.Name)
	if err != nil {
		return nil, seer.Wrap("check_collection_name", err)
	}
	if exists {
		return nil, apperrors.NewValidationError(apperrors.ErrorMap{
			"name": {"A collection with this name already exists"},
		})
	}

	// Set slug if not provided
	slug := params.Collection.Slug
	if slug == "" {
		slug = lib.Slugify(params.Collection.Name)
	}

	// Check if collection slug already exists
	exists, err = c.SlugExists(params.Collection.WorkspaceID, slug)
	if err != nil {
		return nil, seer.Wrap("check_collection_slug", err)
	}

	if exists {
		// If it was provided by the user, return an error immediately
		if params.Collection.Slug != "" {
			return nil, apperrors.NewValidationError(apperrors.ErrorMap{
				"slug": {"This slug is already in use"},
			})
		}

		// Otherwise, append a random slug
		slug = fmt.Sprintf("%s-%d", slug, lib.RandomInt(1000, 9999))
	}

	createdCollection, err := c.queries.CreateCollection(
		context.TODO(),
		queries.CreateCollectionParams{
			WorkspaceID: params.Collection.WorkspaceID,
			Name:        params.Collection.Name,
			Description: lib.PgText(params.Collection.Description),
			Slug:        lib.PgText(slug),
			OwnerID:     params.Owner.UserID,
		},
	)
	if err != nil {
		return nil, seer.Wrap("create_collection", err)
	}

	collection := new(models.Collection)
	collection.From(&createdCollection)

	// Assign owner to collection
	err = c.queries.CreateCollectionMember(context.TODO(), queries.CreateCollectionMemberParams{
		CollectionID: createdCollection.ID,
		UserID:       params.Owner.UserID,
		BitmaskRole:  rbac.RoleOwner,
	})
	if err != nil {
		return nil, seer.Wrap("assign_creator_to_collection", err)
	}

	if lib.Deref(params.AddAllMembers, false) {
		// Assign all workspace members to collection if required (with user role by default)
		err = c.queries.AssignAllWorkspaceMembersToCollection(
			context.TODO(),
			queries.AssignAllWorkspaceMembersToCollectionParams{
				CollectionID: createdCollection.ID,
				WorkspaceID:  createdCollection.WorkspaceID,
				Role:         rbac.RoleUser,
				OwnerID:      params.Owner.UserID,
			},
		)
		if err != nil {
			return nil, seer.Wrap("assign_all_members_to_collection", err)
		}
	}

	return collection, nil
}

func (c *collectionRepo) FindByWorkspaceAndUser(
	workspaceID int32,
	userID int32,
) ([]models.Collection, error) {
	rows, err := c.queries.FindCollectionsByWorkspaceAndUser(
		context.TODO(),
		queries.FindCollectionsByWorkspaceAndUserParams{
			UserID:      userID,
			WorkspaceID: workspaceID,
		},
	)
	if err != nil {
		return nil, seer.Wrap("find_collections_by_workspace_and_user", err)
	}

	collections := make([]models.Collection, 0, len(rows))
	for i := range rows {
		collection := &rows[i]
		c := models.Collection{
			ID:           collection.PublicID.String(),
			Name:         collection.Name.String,
			Slug:         collection.Slug.String,
			WorkspaceID:  collection.WorkspaceID.Int32,
			Description:  collection.Description.String,
			AvatarID:     collection.AvatarID.String,
			CreatedAt:    collection.CreatedAt.Time,
			MembersCount: collection.MemberCount,
			EntriesCount: collection.EntryCount,
			InternalID:   collection.ID.Int32,
			OwnerID:      collection.OwnerID.Int32,
		}
		collections = append(collections, c)
	}

	return collections, nil
}

var _ CollectionRepository = (*collectionRepo)(nil)
