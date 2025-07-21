package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/adelowo/gulter"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/document"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/hubble/web/pkg/ograph"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
	"golang.org/x/sync/errgroup"
)

type entryHandler struct {
	*baseHandler
}

// Search implements EntryHandler.
func (e *entryHandler) Search(ctx *robin.Context, request SearchRequest) (SearchResponse, error) {
	var response SearchResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}
	request.Query = strings.TrimSpace(request.Query)

	workspaceID := lib.PgUUIDString(request.WorkspaceID)
	status, err := e.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{
			PublicID: workspaceID,
			Slug:     request.WorkspaceSlug,
		}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !status.MembershipStatus.Role.Can(rbac.PermSearchEntry) {
		return response, rbac.ErrPermissionDenied
	}

	pagination := repository.PaginationParams{
		Page:    1,
		PerPage: 75, // For now, 75
	}

	// We need a context to control the timeout
	parentCtx := context.Background()
	timedCtx, cancel := context.WithTimeout(parentCtx, time.Minute*1)
	defer cancel()

	// Generate vector if enabled
	var vector []float32
	if e.config.LLM.EnabledEmbeddings() {
		vector, err = e.llm.GenerateEmbedding(timedCtx, request.Query)
		if err != nil {
			return response, err
		}
	}

	start := time.Now() // we are not interested in the time taken for the vector generation, we don't control this
	results, err := e.repos.EntryRepository().QueryWithHybridSearch(&repository.HybridSearchArgs{
		Context:        timedCtx,
		TextQuery:      request.Query,
		SemanticVector: vector,
		UserID:         auth.UserID,
		Pagination:     pagination,
		Workspace: repository.PublicIdOrSlug{
			PublicID: workspaceID,
			Slug:     request.WorkspaceSlug,
		},
	})
	if err != nil {
		return response, err
	}

	// Apply specificity if set
	if e.config.Search.Mode == config.SearchModeThreshold {
		results.CollapsedResults = repository.ApplyMinThreshold(
			results.CollapsedResults,
			e.config.Search.Threshold,
		)
	}

	return SearchResponse{
		Results:   *results,
		Query:     request.Query,
		TimeTaken: time.Since(start).Milliseconds(),
	}, nil
}

// Find implements EntryHandler.
func (e *entryHandler) Find(
	ctx *robin.Context,
	request FindEntryRequest,
) (FindEntryResponse, error) {
	var response FindEntryResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	entry, err := e.repos.EntryRepository().FindByID(&repository.FindbyIdArgs{
		InternalID: 0,
		PublicID:   lib.PgUUIDString(request.EntryID),
		Collection: repository.PublicIdOrSlug{Slug: request.CollectionSlug}, //nolint:exhaustruct
		Workspace:  repository.PublicIdOrSlug{Slug: request.WorkspaceSlug},  //nolint:exhaustruct
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return response, apperrors.BadRequest("entry not found or has been deleted")
		}
		return response, err
	}

	// NOTE: We don't really care about the workspace here since entries are largely independent of their workspace (for now anyway)

	// Check if the user has access to the entry in this collection's context
	perm, err := e.repos.CollectionRepository().
		FindWithMembershipStatus(&repository.FindWithMembershipStatusArgs{
			UserID:         auth.UserID,
			WorkspaceID:    entry.Workspace.ID,
			CollectionID:   entry.Collection.ID,
			WorkspaceSlug:  "",
			CollectionSlug: "",
		})
	if err != nil {
		return response, err
	}

	if !perm.MembershipStatus.Role.Can(rbac.PermReadEntry) {
		return response, rbac.ErrPermissionDenied
	}

	return FindEntryResponse{Entry: entry}, nil
}

// Requeue implements EntryHandler.
func (e *entryHandler) Requeue(
	ctx *robin.Context,
	request RequeueEntriesRequest,
) (RequeueEntriesResponse, error) {
	var response RequeueEntriesResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return response, err
	}

	workspaceId := lib.PgUUIDString(request.WorkspaceID)
	result, err := e.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermRequeueEntry) {
		return response, rbac.ErrPermissionDenied
	}

	// Filter duplicates and format the entries
	uniqueEntries := lib.UniqueSlice(request.EntryIDs)
	ids := make([]pgtype.UUID, 0, len(uniqueEntries))
	for _, entry := range uniqueEntries {
		id, err := lib.UUIDFromString(entry)
		if err != nil {
			return response, apperrors.BadRequest("invalid entry ID: " + entry)
		}
		ids = append(ids, id)
	}

	queued, err := e.repos.EntryRepository().RequeueEntries(&repository.RequeueEntriesArgs{
		WorkspaceID:    workspaceId,
		EntryPublicIDs: ids,
	})
	if err != nil {
		return response, err
	}

	if len(queued) == 0 {
		return response, apperrors.BadRequest("unable to requeue entries")
	}

	// Send the entries to the queue
	go func(entries []int32) {
		for _, id := range entries {
			if err := e.queue.Add(&job.EntryJob{ID: id}); err != nil {
				log.Error().Err(err).Msg("failed to enqueue entry")
			}
		}
		log.Debug().Int("count", len(entries)).Msg("entries re-queued successfully")
	}(queued)

	return RequeueEntriesResponse{
		WorkspaceSlug: result.Workspace.Slug,
		Count:         len(queued),
	}, nil
}

// Delete implements EntryHandler.
func (e *entryHandler) Delete(
	ctx *robin.Context,
	request DeleteEntriesRequest,
) (DeleteEntriesResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return DeleteEntriesResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return DeleteEntriesResponse{}, err
	}

	// Filter duplicates
	uniqueEntries := lib.UniqueSlice(request.Entries)
	entries := make([]pgtype.UUID, 0)

	for _, entry := range uniqueEntries {
		id, err := lib.UUIDFromString(entry)
		if err != nil {
			return DeleteEntriesResponse{}, apperrors.BadRequest("invalid entry ID: " + entry)
		}

		entries = append(entries, id)
	}

	ownerships, err := e.repos.EntryRepository().GetOwnerships(auth.UserID, entries)
	if err != nil {
		return DeleteEntriesResponse{}, seer.Wrap(
			"get_entries_ownerships",
			err,
			"oops, we ran into an issue while trying to process your request",
		)
	}

	// Filter out entries that the user doesn't have access to delete
	entriesToDelete := make([]pgtype.UUID, 0)
	for _, ownership := range ownerships {
		if ownership.IsOwner || ownership.UserRole.Can(rbac.PermDeleteEntry) {
			entriesToDelete = append(entriesToDelete, ownership.EntryID)
		}
	}

	if len(entriesToDelete) == 0 {
		return DeleteEntriesResponse{}, apperrors.Forbidden("no entries with delete permission")
	}

	// Delete entries
	deletedEntries, err := e.repos.EntryRepository().Delete(repository.DeleteEntriesArgs{
		Entries: entries,
	})
	if err != nil {
		return DeleteEntriesResponse{}, err
	}

	suffix := "ies"
	if len(deletedEntries) == 1 {
		suffix = "y"
	}
	message := fmt.Sprintf("successfully deleted %d entr%s", len(deletedEntries), suffix)
	if len(entriesToDelete) != len(entries) {
		message = "some entries could not be deleted, this may be due to insufficient permissions"
	}

	return DeleteEntriesResponse{
		DeletedEntries: deletedEntries,
		Message:        lib.UppercaseFirst(message),
	}, nil
}

// FindCollectionEntries implements EntryHandler.
func (e *entryHandler) FindCollectionEntries(
	ctx *robin.Context,
	request FindCollectionEntriesRequest,
) (FindEntriesResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return FindEntriesResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return FindEntriesResponse{}, err
	}

	member, err := e.repos.CollectionRepository().FindMember(
		repository.PublicIdOrSlug{Slug: request.WorkspaceSlug},  //nolint:exhaustruct
		repository.PublicIdOrSlug{Slug: request.CollectionSlug}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return FindEntriesResponse{}, err
	}

	if !member.Role.Can(rbac.PermListCollectionEntries) {
		return FindEntriesResponse{}, apperrors.Forbidden(
			"permission denied",
		)
	}

	data, err := e.repos.EntryRepository().FindAllWithPagination(
		//nolint:exhaustruct
		&repository.FindEntriesArgs{
			Workspace: repository.PublicIdOrSlug{
				Slug: request.WorkspaceSlug,
			},
			Collection: repository.PublicIdOrSlug{
				Slug: request.CollectionSlug,
			},
			UserID: auth.UserID,
		},
		request.Pagination,
	)
	if err != nil {
		return FindEntriesResponse{}, err
	}

	return FindEntriesResponse{
		Entries:       data.Entries,
		WorkspaceSlug: request.WorkspaceSlug,
		Pagination: request.Pagination.ToState(repository.PageStateArgs{
			CurrentCount: len(data.Entries),
			TotalCount:   data.TotalCount,
		}),
	}, nil
}

// FindWorkspaceEntries implements EntryHandler.
func (e *entryHandler) FindWorkspaceEntries(
	ctx *robin.Context,
	request FindWorkspaceEntriesRequest,
) (FindEntriesResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return FindEntriesResponse{}, err
	}

	if err := lib.ValidateStruct(&request); err != nil {
		return FindEntriesResponse{}, err
	}

	workspace, err := e.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{Slug: request.WorkspaceSlug}, //nolint:exhaustruct
		auth.UserID,
	)
	if err != nil {
		return FindEntriesResponse{}, err
	}
	canListEntries := workspace.MembershipStatus.Role.Can(rbac.PermListWorkspaceEntries)
	if !canListEntries {
		return FindEntriesResponse{}, apperrors.Forbidden("permission denied")
	}
	data, err := e.repos.EntryRepository().FindAllWithPagination(
		//nolint:exhaustruct
		&repository.FindEntriesArgs{
			Workspace: repository.PublicIdOrSlug{PublicID: workspace.PublicID},
			UserID:    auth.UserID,
		}, request.Pagination,
	)
	if err != nil {
		return FindEntriesResponse{}, err
	}

	return FindEntriesResponse{
		Entries:       data.Entries,
		WorkspaceSlug: workspace.Workspace.Slug,
		Pagination: request.Pagination.ToState(repository.PageStateArgs{
			CurrentCount: len(data.Entries),
			TotalCount:   data.TotalCount,
		}),
	}, nil
}

// LookupUrl implements EntryHandler.
func (e *entryHandler) GetLinkMetadata(ctx *robin.Context, link string) (ograph.Metadata, error) {
	if link == "" {
		return ograph.Metadata{}, apperrors.BadRequest("url is required")
	}

	parsedUrl, err := url.Parse(link)
	if err != nil {
		return ograph.Metadata{}, apperrors.BadRequest("invalid url")
	}

	// Check if metadata is cached
	cachedMetadata, exists, err := e.repos.EntryRepository().GetLinkMetadata(parsedUrl.String())
	if err != nil {
		return ograph.Metadata{}, err
	}

	// Return cached metadata if it exists
	if exists {
		return cachedMetadata, nil
	}

	meta, err := ograph.Parse(parsedUrl.String())
	if err != nil {
		return ograph.Metadata{}, err
	}

	// Cache metadata
	go func() {
		if err := e.repos.EntryRepository().SaveLinkMetadata(parsedUrl.String(), meta); err != nil {
			log.Error().Err(err).Msg("failed to cache metadata")
		}
	}()

	return *meta, nil
}

// Create implements EntryHandler.
func (e *entryHandler) Import(ctx *robin.Context, body io.ReadCloser) (ImportEntryResponse, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return ImportEntryResponse{}, err
	}

	payload, err := extractImportPayload(ctx)
	if err != nil {
		return ImportEntryResponse{}, err
	}

	// Ensure collection exists in workspace
	collectionExists, err := e.repos.WorkspaceRepository().
		CollectionExists(payload.WorkspaceID, payload.CollectionID)
	if err != nil {
		return ImportEntryResponse{}, err
	}

	if !collectionExists {
		return ImportEntryResponse{}, apperrors.BadRequest(
			"collection does not exist in this workspace",
		)
	}

	// Check collection permissions
	workspaceData, err := e.repos.CollectionRepository().FindWithMembershipStatus(
		//nolint:exhaustruct
		&repository.FindWithMembershipStatusArgs{
			UserID:       auth.UserID,
			WorkspaceID:  payload.WorkspaceID,
			CollectionID: payload.CollectionID,
		},
	)
	if err != nil {
		return ImportEntryResponse{}, err
	}
	if !workspaceData.MembershipStatus.Role.Can(rbac.PermCreateEntry) {
		return ImportEntryResponse{}, apperrors.Forbidden("permission denied")
	}

	// Look up the actual internal ID so we don't have to do a bunch of joins for all the (potentially) bulk inserts
	collectionID, err := e.repos.CollectionRepository().
		GetInternalID(repository.GetInternalIDParams{
			WorkspaceID:  payload.WorkspaceID,
			CollectionID: payload.CollectionID,
		})
	if err != nil {
		return ImportEntryResponse{}, seer.Wrap("lookup_internal_collection_id", err)
	}

	createdEntries := make([]models.CreatedEntry, 0, len(payload.Links)+len(payload.Files))

	// Save link entries first
	if len(payload.Links) > 0 {
		links := e.ImportLinks(ctx, &ImportLinksPayload{
			Links:        payload.Links,
			CollectionID: collectionID,
			UserID:       auth.UserID,
		})

		createdEntries = append(createdEntries, links...)
	}

	// Save file entries
	if len(payload.Files) > 0 {
		files := e.ImportFiles(ctx, &ImportFilesPayload{
			Files:        payload.Files,
			CollectionID: collectionID,
			UserID:       auth.UserID,
		})

		createdEntries = append(createdEntries, files...)
	}

	// Queue all entries for processing
	entries := make([]repository.EnqueueEntryParams, 0, len(createdEntries))
	for _, entry := range createdEntries {
		// Send them all to the queue
		entries = append(entries, repository.EnqueueEntryParams{
			ID:      entry.InternalID,
			Payload: document.QueuePayload{Type: entry.Type},
		})
	}

	if err = e.repos.EntryRepository().EnqueueEntries(entries); err != nil {
		return ImportEntryResponse{}, seer.Wrap("enqueue_entries_in_handler", err)
	}

	// Send the entries to the queue
	go func(entries []repository.EnqueueEntryParams) {
		for _, entry := range entries {
			if err := e.queue.Add(&job.EntryJob{ID: entry.ID}); err != nil {
				log.Error().Err(err).Msg("failed to enqueue entry")
			}
		}
		log.Debug().Msg("entries enqueued successfully")
	}(entries)

	return ImportEntryResponse{
		WorkspaceID:  payload.WorkspaceID,
		CollectionID: payload.CollectionID,
		Entries:      createdEntries,
	}, nil
}

func (e *entryHandler) ImportFiles(
	ctx *robin.Context,
	payload *ImportFilesPayload,
) []models.CreatedEntry {
	entries := make([]models.CreatedEntry, 0, len(payload.Files))
	succeeded := 0

	eg := new(errgroup.Group)
	for _, file := range payload.Files {
		file := file
		eg.Go(func() error {
			e, err := e.repos.EntryRepository().CreateFileEntry(&models.FileEntry{
				FileID:       file.StorageKey,
				OriginalName: file.OriginalName,
				SavedName:    file.UploadedFileName,
				Filesize:     file.Size,
				MimeType:     file.MimeType,
				Type: document.InferType(
					document.InferExtension(file.OriginalName),
					file.MimeType,
				),
				CollectionID: payload.CollectionID,
				UserID:       payload.UserID,
			})
			if err != nil {
				return err
			}

			entries = append(entries, e)
			succeeded++
			return nil
		})
	}

	// We don't need to handle errors here, as we're just counting the number of successful imports
	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("failed to import files")
	}

	return entries
}

func (e *entryHandler) ImportLinks(
	ctx *robin.Context,
	payload *ImportLinksPayload,
) []models.CreatedEntry {
	entries := make([]models.CreatedEntry, 0, len(payload.Links))
	succeeded := 0

	eg := new(errgroup.Group)
	for _, link := range payload.Links {
		link := link
		eg.Go(func() error {
			meta, err := e.GetLinkMetadata(ctx, link)
			if err != nil {
				return err
			}

			e, err := e.repos.EntryRepository().CreateLinkEntry(&models.LinkEntry{
				Title:        meta.Title,
				Link:         link,
				CollectionID: payload.CollectionID,
				UserID:       payload.UserID,
				Metadata:     meta,
			})
			if err != nil {
				return err
			}

			entries = append(entries, e)
			succeeded++
			return nil
		})
	}

	// We don't need to handle errors here, as we're just counting the number of successful imports
	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("failed to import links")
	}

	return entries
}

func extractImportPayload(ctx *robin.Context) (*ImportEntryPayload, error) {
	payload := new(ImportEntryPayload)

	err := ctx.Request().ParseMultipartForm(25 << 20) // 10 MB
	if err != nil {
		return &ImportEntryPayload{}, apperrors.BadRequest("failed to parse form")
	}

	ctxFiles, ok := ctx.Get("files").(gulter.Files)
	if !ok {
		// Totally fine to not have files
		ctxFiles = gulter.Files{}
	}

	files := []gulter.File{}
	if f, ok := ctxFiles["files"]; ok {
		files = f
	}

	links := ctx.Request().MultipartForm.Value["links"]
	if len(links) == 0 && len(files) == 0 {
		return &ImportEntryPayload{}, apperrors.BadRequest("no links or files provided")
	}

	collectionID := ctx.Request().MultipartForm.Value["collection_id"]
	if len(collectionID) == 0 {
		return &ImportEntryPayload{}, apperrors.BadRequest("collection ID is required")
	}

	workspaceID := ctx.Request().MultipartForm.Value["workspace_id"]
	if len(workspaceID) == 0 {
		return &ImportEntryPayload{}, apperrors.BadRequest("workspace ID is required")
	}

	workspaceUUID, err := lib.UUIDFromString(workspaceID[0])
	if err != nil {
		return &ImportEntryPayload{}, apperrors.BadRequest("invalid workspace ID")
	}

	collectionUUID, err := lib.UUIDFromString(collectionID[0])
	if err != nil {
		return &ImportEntryPayload{}, apperrors.BadRequest("invalid collection ID")
	}

	payload.Links = links
	payload.Files = files
	payload.CollectionID = collectionUUID
	payload.WorkspaceID = workspaceUUID

	return payload, nil
}

var _ EntryHandler = (*entryHandler)(nil)
