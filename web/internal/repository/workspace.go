package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/models"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/seer"
	"golang.org/x/sync/errgroup"
)

var (
	ErrWorkspaceNotFound     = apperrors.BadRequest("workspace not found")
	ErrInviteForAnotherEmail = apperrors.BadRequest(
		"this invite was sent to another email address, please sign in with that email to continue",
	)
)

const (
	MinInviteResendInterval = 2 * 24 * time.Hour
)

const (
	WorkspaceInviteTTL = 14 * 24 * time.Hour
)

type (
	workspaceRepo struct {
		*baseRepo
	}

	UpsertWorkspaceInviteParams struct {
		WorkspaceID int32
		Email       string
		Role        rbac.Role
		InitiatorID int32
	}

	MemberLookupArgs struct {
		Email    string
		PublicID pgtype.UUID
		ID       int32
	}

	FindMembersResult struct {
		Members []models.Member
		Total   int64
	}

	FindInviteArgs struct {
		InviteID pgtype.UUID
	}

	UpdateInviteStatusArgs struct {
		Invite      models.WorkspaceInvite
		Status      models.InviteStatus
		InvitedUser models.User
	}

	UpdateMemberRoleArgs struct {
		WorkspaceID int32
		UserID      pgtype.UUID
		Role        rbac.Role
	}

	RemoveWorkspaceMemberArgs struct {
		WorkspaceOwnerID int32
		WorkspaceID      int32
		// TargetUserID is the real `user_id` of the user to be removed (column from the users table) not the member ID
		TargetUserID int32
	}

	UpdateWorkspaceDetailsArgs struct {
		Name        string
		Description string
		Slug        string
		WorkspaceID int32
	}

	FindMemberArg struct {
		WorkspaceID pgtype.UUID
		// UserID is the ID of the user in the users table
		UserID int32
		// MemberID is the ID of the member in the workspace_members table
		MemberID int32
		// Email is the email of the user
		Email string
	}

	WorkspaceRepository interface {
		// Create a new workspace (and the default member i.e the owner)
		Create(workspace *models.Workspace) (*models.Workspace, error)

		// Update a workspace
		UpdateDetails(args UpdateWorkspaceDetailsArgs) (models.Workspace, error)

		// FindByID finds a workspace by its public ID
		FindByID(workspaceID pgtype.UUID) (*models.Workspace, error)

		// FindByInternalID finds a workspace by its internal ID
		FindByInternalID(workspaceID int32) (*models.Workspace, error)

		// Find all workspaces a user is a member of
		FindAllByUserID(userID int32) ([]models.Workspace, error)

		// Find a workspace and the user's status (role, is_member) in the workspace
		FindWithMembershipStatus(
			params PublicIdOrSlug,
			userID int32,
		) (*models.WorkspaceWithMembershipStatus, error)

		// Find a workspace's members
		FindMembers(
			workspaceID pgtype.UUID,
			pagination PaginationParams,
		) (FindMembersResult, error)

		// Find a workspace member by ID
		FindMember(args FindMemberArg) (models.Member, error)

		// Delete a workspace (mark it as deleted and all its related data: members, collections, entries)
		Delete(workspaceID pgtype.UUID) error

		// Find a workspace's public ID by its slug (cached)
		SlugToID(slug string) (pgtype.UUID, error)

		// Exists checks if a collection exists in a workspace
		NameExists(name string, ownerId int32) (bool, error)

		// SlugExists checks if a collection exists in a workspace
		SlugExists(slug string) (bool, error)

		// MemberExists checks if a member exists in a workspace
		MemberExists(workspaceID int32, args MemberLookupArgs) (bool, error)

		// UpdateMemberRole updates a member's role in a workspace
		UpdateMemberRole(args UpdateMemberRoleArgs) (err error)

		// CollectionExists checks if a collection exists in a workspace
		CollectionExists(workspaceID pgtype.UUID, collectionID pgtype.UUID) (bool, error)

		// FindInvite finds an invite by ID and the current user
		FindInvite(args FindInviteArgs) (models.WorkspaceInvite, error)

		// UpsertWorkspaceInvite creates or updates a workspace invite
		UpsertWorkspaceInvite(args UpsertWorkspaceInviteParams) (pgtype.UUID, error)

		// TrackInviteSent caches the invite sent to an email for a workspace (for rate limiting)
		TrackInviteSent(workspaceID int32, email string, inviteID pgtype.UUID) error

		// UntrackInvite removes the invite from the cache
		UntrackInvite(workspaceID int32, email string, inviteID pgtype.UUID) error

		// ValidateInvite checks if an invite is valid
		ValidateInvite(invite models.WorkspaceInvite) error

		// UpdateInviteStatus updates the status of a workspace invite (optional: create a new member if the status is accepted)
		UpdateInviteStatus(args UpdateInviteStatusArgs) error

		// RemoveMember removes a member from a workspace and all its collections
		RemoveMember(args RemoveWorkspaceMemberArgs) error
	}
)

// FindByInternalID implements WorkspaceRepository.
func (w *workspaceRepo) FindByInternalID(workspaceID int32) (*models.Workspace, error) {
	row, err := w.queries.FindWorkspaceByID(context.TODO(), workspaceID)
	if err != nil {
		return nil, err
	}

	workspace := new(models.Workspace)
	workspace.From(&row.Workspace)
	return workspace, nil
}

// Delete implements WorkspaceRepository.
func (w *workspaceRepo) Delete(workspaceID pgtype.UUID) error {
	result, err := w.queries.MarkWorkspaceAsDeleted(context.TODO(), workspaceID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("workspace not found or already deleted")
	}

	return nil
}

// UpdateDetails implements WorkspaceRepository.
func (w *workspaceRepo) UpdateDetails(
	args UpdateWorkspaceDetailsArgs,
) (models.Workspace, error) {
	result, err := w.queries.UpdateWorkspaceDetails(
		context.TODO(),
		queries.UpdateWorkspaceDetailsParams{
			Name:        lib.PgText(args.Name),
			Description: lib.PgText(args.Description),
			Slug:        lib.PgText(args.Slug),
			WorkspaceID: args.WorkspaceID,
		},
	)
	if err != nil {
		return models.Workspace{}, err
	}

	workspace := models.Workspace{} //nolint:exhaustruct
	workspace.From(&result)
	return workspace, nil
}

func (w *workspaceRepo) RemoveMember(args RemoveWorkspaceMemberArgs) error {
	tx, err := w.pool.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(context.TODO()) //nolint:errcheck

	queriesWithTx := w.queries.WithTx(tx)

	// mark workspce member as deleted
	result, err := queriesWithTx.MarkWorkspaceMemberAsDeleted(
		context.TODO(),
		queries.MarkWorkspaceMemberAsDeletedParams{
			WorkspaceID: args.WorkspaceID,
			UserID:      args.TargetUserID,
		},
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return seer.New("mark_workspace_member_deleted", "", http.StatusBadRequest)
	}

	// Attempt to reassign the ownership to the workspace owner if they are in the collection
	collectionsWithWkOwner, err := queriesWithTx.FindUserCollectionsWithWorkspaceOwner(
		context.TODO(),
		queries.FindUserCollectionsWithWorkspaceOwnerParams{
			WorkspaceID: args.WorkspaceID,
			UserID:      args.TargetUserID,
		},
	)
	if err != nil {
		return err
	}
	if err := queriesWithTx.ReassignCollectionOwnership(context.TODO(), queries.ReassignCollectionOwnershipParams{
		WorkspaceID:   args.WorkspaceID,
		NewOwnerID:    args.WorkspaceOwnerID,
		OldOwnerID:    args.TargetUserID,
		OwnerRole:     rbac.RoleOwner,
		CollectionIds: collectionsWithWkOwner,
	}); err != nil {
		return seer.Wrap("reassign_collection_ownership", err)
	}

	// Otherwise, just mark it as deleted
	if err := queriesWithTx.MarkUsersCollectionsAsDeleted(context.TODO(), queries.MarkUsersCollectionsAsDeletedParams{
		UserID:      args.TargetUserID,
		WorkspaceID: args.WorkspaceID,
	}); err != nil {
		return seer.Wrap("mark_users_collections_as_deleted", err)
	}

	// Commit the transaction
	if err := tx.Commit(context.TODO()); err != nil {
		return seer.Wrap("commit_tx", err)
	}

	return nil
}

// UpdateMemberRole implements WorkspaceRepository.
func (w *workspaceRepo) UpdateMemberRole(args UpdateMemberRoleArgs) (err error) {
	tx, err := w.pool.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return seer.Wrap("begin_tx", err)
	}
	defer tx.Rollback(context.TODO()) //nolint:errcheck

	queriesWithTx := w.queries.WithTx(tx)

	// Update collections roles for the user where the user is not the owner
	err = queriesWithTx.UpdateAllCollectionMemberRole(
		context.TODO(),
		queries.UpdateAllCollectionMemberRoleParams{
			Role:      args.Role,
			UserID:    args.UserID,
			OwnerRole: rbac.RoleOwner,
		},
	)
	if err != nil {
		return seer.Wrap("update_all_collection_member_role", err)
	}

	// Update the member role
	err = queriesWithTx.UpdateWorkspaceMemberRole(
		context.TODO(),
		queries.UpdateWorkspaceMemberRoleParams{
			Role:        args.Role,
			WorkspaceID: args.WorkspaceID,
			UserID:      args.UserID,
		},
	)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(context.TODO()); err != nil {
		return seer.Wrap("commit_tx", err)
	}

	return nil
}

// UpdateInviteStatus implements WorkspaceRepository.
func (w *workspaceRepo) UpdateInviteStatus(args UpdateInviteStatusArgs) error {
	if args.Status != models.InviteStatusAccepted &&
		args.Status != models.InviteStatusDeclined &&
		args.Status != models.InviteStatusRevoked {
		return apperrors.New(
			"Invalid invite status",
			http.StatusBadRequest,
		)
	}

	_, err := w.queries.UpdateWorkspaceInviteStatus(
		context.TODO(),
		queries.UpdateWorkspaceInviteStatusParams{
			Status:   string(args.Status),
			InviteID: args.Invite.InviteID,
		},
	)
	if err != nil {
		return err
	}

	// Create new workspace member if invite is accepted
	if args.Status == models.InviteStatusAccepted {
		_, err = w.queries.CreateWorkspaceMember(
			context.TODO(),
			queries.CreateWorkspaceMemberParams{
				WorkspaceID: args.Invite.Workspace.ID,
				UserID:      args.InvitedUser.ID,
				Role:        args.Invite.Role,
			},
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateInvite implements WorkspaceRepository.
func (w *workspaceRepo) ValidateInvite(invite models.WorkspaceInvite) error {
	switch {
	case invite.Status == models.InviteStatusAccepted:
		return apperrors.New(
			"This invite has already been accepted",
			http.StatusConflict,
		)

	case invite.Status == models.InviteStatusDeclined:
		return apperrors.New(
			"This invite has already been declined, please request a new invite",
			http.StatusConflict,
		)

	case invite.Status == models.InviteStatusRevoked:
		return apperrors.New(
			"This invite has been revoked, please request a new invite",
			http.StatusConflict,
		)

	case invite.Status == models.InviteStatusExpired:
		return apperrors.New(
			"This invite has expired, please request a new invite",
			http.StatusConflict,
		)

	// Check that the invite has not expired (7 days)
	case invite.InvitedAt.Add(WorkspaceInviteTTL).Before(time.Now()):
		return apperrors.New(
			"Invite has expired, please request a new invite",
			http.StatusConflict,
		)
	}

	return nil
}

// FindInvite implements WorkspaceRepository.
func (w *workspaceRepo) FindInvite(args FindInviteArgs) (models.WorkspaceInvite, error) {
	row, err := w.queries.FindInviteById(context.TODO(), args.InviteID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.WorkspaceInvite{}, apperrors.BadRequest(
				"this invite either does not exist or was not sent to you, please check the invite link and try again.",
			)
		}
	}

	//nolint:exhaustruct
	inviter := models.User{}
	inviter.From(&row.User)
	invite := models.WorkspaceInvite{
		Workspace: models.InviteWorkspaceDetails{
			ID:       row.WorkspaceID,
			PublicID: row.WorkspacePublicID,
			Name:     row.WorkspaceName,
			Slug:     row.WorkspaceSlug.String,
		},
		Inviter:   inviter,
		Status:    models.InviteStatus(row.Status),
		Role:      row.Role,
		InvitedAt: row.InvitedAt.Time,
		ID:        row.ID,
		InviteID:  row.InviteID,
		Invited: models.InvitedDetails{
			UserID: row.InvitedUserID.Int32,
			Email:  row.InvitedEmail,
			Exists: row.InvitedUserExists,
		},
	}

	return invite, nil
}

// FindMembers implements WorkspaceRepository.
func (w *workspaceRepo) FindMembers(
	workspaceID pgtype.UUID,
	pagination PaginationParams,
) (FindMembersResult, error) {
	rows, err := w.queries.FindWorkspaceMembers(context.TODO(), queries.FindWorkspaceMembersParams{
		Limit:       pagination.Limit(),
		Offset:      pagination.Offset(),
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return FindMembersResult{}, err
	}

	var total int64
	members := make([]models.Member, 0)
	for i := range len(rows) {
		row := &rows[i]
		member := models.Member{
			ID:        row.ID,
			InviteID:  row.InviteID,
			FirstName: row.FirstName.String,
			LastName:  row.LastName.String,
			Email:     row.Email.String,
			Role:      rbac.Role(row.Role),
			Status:    models.InviteStatus(row.Status),
			CreatedAt: row.CreatedAt.Time,
			User: models.MemberUser{
				ID:       row.UserID.Int32,
				PublicID: row.PublicUserID,
			},
		}

		if total == 0 {
			total = row.TotalCount
		}

		members = append(members, member)
	}

	return FindMembersResult{
		Members: lib.WithMaxSize(members, pagination.PerPage),
		Total:   total,
	}, nil
}

// FindMember implements WorkspaceRepository.
func (w *workspaceRepo) FindMember(args FindMemberArg) (models.Member, error) {
	row, err := w.queries.FindWorkspaceMemeber(context.TODO(), queries.FindWorkspaceMemeberParams{
		WorkspaceID: args.WorkspaceID,
		Email:       lib.PgText(args.Email),
		UserID:      lib.PgInt4(args.UserID),
		MemberID:    lib.PgInt4(args.MemberID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Member{}, apperrors.BadRequest("member not found")
		}

		return models.Member{}, err
	}

	return models.Member{
		ID:        row.ID,
		InviteID:  row.InviteID.String(),
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Email:     row.Email,
		Role:      row.Role,
		Status:    models.InviteStatus(row.Status),
		CreatedAt: row.CreatedAt.Time,
		User: models.MemberUser{
			ID:       row.UserID,
			PublicID: row.PublicUserID,
		},
	}, nil
}

// TrackInviteSent implements WorkspaceRepository.
func (w *workspaceRepo) TrackInviteSent(
	workspaceID int32,
	email string,
	inviteID pgtype.UUID,
) error {
	storeKey := kv.KeyWorkspaceInvite(workspaceID, email)

	err := w.store.SetWithTTL(storeKey, []byte(inviteID.String()), MinInviteResendInterval)
	if err != nil {
		log.Error().
			Err(err).
			Str("storeKey", storeKey.String()).
			Msg("failed to cache workspace invite")
	}

	return err
}

func (w *workspaceRepo) UntrackInvite(
	workspaceID int32,
	email string,
	inviteID pgtype.UUID,
) error {
	storeKey := kv.KeyWorkspaceInvite(workspaceID, email)

	exists, err := w.store.Exists(storeKey)
	if err != nil {
		return seer.Wrap("key_workspace_invite_exists", err)
	}

	if !exists {
		return nil
	}

	err = w.store.Delete(storeKey)
	return err
}

// MemberExists implements WorkspaceRepository.
func (w *workspaceRepo) MemberExists(workspaceID int32, args MemberLookupArgs) (bool, error) {
	exists, err := w.queries.WorkspaceMemberExists(
		context.TODO(),
		queries.WorkspaceMemberExistsParams{
			WorkspaceID:  workspaceID,
			Email:        lib.PgText(args.Email),
			UserID:       lib.PgInt4(args.ID),
			PublicUserID: args.PublicID,
		},
	)
	return exists, err
}

// UpsertWorkspaceInvite implements WorkspaceRepository.
func (w *workspaceRepo) UpsertWorkspaceInvite(
	args UpsertWorkspaceInviteParams,
) (pgtype.UUID, error) {
	if args.Role == rbac.RoleOwner {
		return pgtype.UUID{}, apperrors.Forbidden("cannot invite an owner to a workspace")
	}

	storeKey := kv.KeyWorkspaceInvite(args.WorkspaceID, args.Email)
	// Check if an invite was sent recently to the same email for the same workspace
	exists, err := w.store.Exists(storeKey)
	if err != nil {
		return pgtype.UUID{}, seer.Wrap("can_send_workspace_invite", err)
	}

	if exists {
		return pgtype.UUID{}, apperrors.Forbidden(
			"an invite was sent to this email recently, please wait before sending another",
		)
	}

	// Upsert the invite
	invite, err := w.queries.UpsertWorkspaceInvite(
		context.TODO(),
		queries.UpsertWorkspaceInviteParams{
			WorkspaceID: args.WorkspaceID,
			Email:       args.Email,
			Role:        args.Role,
			InvitedBy:   args.InitiatorID,
		},
	)
	if err != nil {
		return pgtype.UUID{}, err
	}

	return invite.InviteID, nil
}

// CollectionExists implements WorkspaceRepository.
func (w *workspaceRepo) CollectionExists(
	workspaceID pgtype.UUID,
	collectionID pgtype.UUID,
) (bool, error) {
	return w.queries.CollectionExistsInWorkspace(
		context.TODO(),
		queries.CollectionExistsInWorkspaceParams{
			WorkspaceID:  workspaceID,
			CollectionID: collectionID,
		},
	)
}

func (w *workspaceRepo) NameExists(name string, ownerId int32) (bool, error) {
	return w.queries.WorkspaceNameExists(context.TODO(), queries.WorkspaceNameExistsParams{
		DisplayName: strings.TrimSpace(name),
		OwnerID:     ownerId,
	})
}

func (w *workspaceRepo) SlugExists(slug string) (bool, error) {
	return w.queries.WorkspaceSlugExists(
		context.TODO(),
		pgtype.Text{String: strings.TrimSpace(slug), Valid: true},
	)
}

func (w *workspaceRepo) SlugToID(slug string) (pgtype.UUID, error) {
	key := kv.KeyWorkspaceSlugId(slug)
	id, err := w.store.Get(key)
	if err != nil && !errors.Is(err, kv.ErrKeyNotFound) {
		return pgtype.UUID{}, seer.Wrap("cached_slug_get", err)
	}

	// Return cached result if found
	if !errors.Is(err, kv.ErrKeyNotFound) && id != nil && len(id) > 0 {
		return lib.UUIDFromString(string(id))
	}

	// Find workspace ID by slug from the database
	workspaceId, err := w.queries.FindWorkspaceIdBySlug(
		context.TODO(),
		pgtype.Text{String: slug, Valid: true},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, ErrWorkspaceNotFound
		}

		return pgtype.UUID{}, err
	}

	// Cache the result
	go func() {
		if err := w.store.Set(key, []byte(workspaceId.String())); err != nil {
			log.Error().Err(err).Str("workspaceSlug", slug).Msg("failed to cache workspace slug")
		}
	}()

	return workspaceId, nil
}

func (w *workspaceRepo) Create(workspace *models.Workspace) (*models.Workspace, error) {
	workspace.Name = strings.TrimSpace(workspace.Name) // Normalize workspace name

	// Check if workspace already exists
	exists, err := w.NameExists(workspace.Name, workspace.OwnerID)
	if err != nil {
		return &models.Workspace{}, err
	}
	if exists {
		return &models.Workspace{}, apperrors.NewValidationError(
			apperrors.ErrorMap{"name": {"A workspace with this name already exists"}},
		)
	}

	// Check slug
	slug := workspace.Slug
	if slug == "" {
		slug = lib.Slugify(workspace.Name)
	}

	// Check if slug already exists
	exists, err = w.SlugExists(slug)
	if err != nil {
		return &models.Workspace{}, err
	}

	if exists {
		// If the slug was provided by the user, return an error
		if workspace.Slug != "" {
			return &models.Workspace{}, apperrors.NewValidationError(
				apperrors.ErrorMap{"slug": {"A workspace with this slug already exists"}},
			)
		}

		// Otherwise, append a random slug
		slug = fmt.Sprintf("%s-%d", slug, lib.RandomInt(1000, 9999))
	}

	description := new(pgtype.Text)
	if err := description.Scan(strings.TrimSpace(workspace.Description)); err != nil {
		return &models.Workspace{}, seer.Wrap("scan_description", err)
	}

	// Create workspace
	createdWorkspace, err := w.queries.CreateWorkspace(
		context.TODO(),
		queries.CreateWorkspaceParams{
			Name:        workspace.Name,
			OwnerID:     workspace.OwnerID,
			Slug:        pgtype.Text{String: slug, Valid: true},
			Description: *description,
		},
	)
	if err != nil {
		return &models.Workspace{}, err
	}

	// Create default member
	_, err = w.queries.CreateWorkspaceMember(
		context.TODO(),
		queries.CreateWorkspaceMemberParams{
			WorkspaceID: createdWorkspace.ID,
			UserID:      createdWorkspace.OwnerID,
			Role:        rbac.RoleOwner,
		},
	)
	if err != nil {
		return &models.Workspace{}, err
	}

	workspace.From(&createdWorkspace)
	return workspace, nil
}

func (w *workspaceRepo) FindByID(workspaceID pgtype.UUID) (*models.Workspace, error) {
	data, err := w.queries.FindWorkspaceByPublicID(context.TODO(), workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}

		return nil, err
	}

	workspace := new(models.Workspace)
	workspace.From(&data)
	return workspace, nil
}

func (w *workspaceRepo) FindAllByUserID(userID int32) ([]models.Workspace, error) {
	workspaces := make([]models.Workspace, 0)

	userWorkspaces, err := w.queries.FindAllWorkspacesByUserID(context.TODO(), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return workspaces, nil
		}

		return nil, err
	}

	eg := new(errgroup.Group)
	eg.SetLimit(runtime.NumCPU())
	for i := range userWorkspaces {
		userWorkspace := &userWorkspaces[i]

		eg.Go(func() error {
			collections, err := w.baseRepo.CollectionRepository().
				FindByWorkspaceAndUser(userWorkspace.ID, userID)
			if err != nil {
				return err
			}

			workspace := new(models.Workspace)
			workspace.From(userWorkspace)
			workspace.Collections = collections
			workspaces = append(workspaces, *workspace)

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, seer.Wrap("find_all_workspaces_by_user_id", err)
	}

	return workspaces, nil
}

func (w *workspaceRepo) FindWithMembershipStatus(
	params PublicIdOrSlug,
	userID int32,
) (*models.WorkspaceWithMembershipStatus, error) {
	data, err := w.queries.FindWorkspaceWithMembershipStatus(
		context.TODO(),
		queries.FindWorkspaceWithMembershipStatusParams{
			UserID:   userID,
			PublicID: params.PublicID,
			Slug:     lib.PgText(params.Slug),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}

		return nil, err
	}

	result := models.WorkspaceWithMembershipStatus{
		ID:       data.ID,
		PublicID: data.PublicID,
		Workspace: &models.Workspace{
			ID:                   data.PublicID.String(),
			Name:                 data.DisplayName,
			Slug:                 data.Slug.String,
			OwnerID:              data.OwnerID,
			Description:          data.Description.String,
			AvatarID:             data.AvatarID.String,
			EnablePublicIndexing: data.EnablePublicIndexing,
			InviteOnly:           data.InviteOnly,
			CreatedAt:            data.CreatedAt.Time,
			InternalID:           data.ID,
		},

		MembershipStatus: &models.MembershipStatus{
			IsMember: data.IsMember,
			Role:     rbac.Role(data.Role),
			MemberID: data.MemberID.Int32,
			UserID:   data.MemberUserID.Int32,
		},
	}

	return &result, nil
}

var _ WorkspaceRepository = (*workspaceRepo)(nil)
