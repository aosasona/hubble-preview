package repository

import (
	"context"
	"encoding/json"
	"errors"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/pkg/document"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/hubble/web/pkg/ograph"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/seer"
)

const (
	DefaultFullTextWeight = 0.35
	DefaultSemanticWeight = 0.65
)

const KeywordScore = 0.075 // When a keyword is matched, it will add this score to the final score

type (
	EnqueueEntryParams struct {
		ID      int32                 `json:"id"`
		Payload document.QueuePayload `json:"payload"`
	}

	FindEntriesArgs struct {
		Workspace  PublicIdOrSlug
		Collection PublicIdOrSlug
		UserID     int32
	}

	FindEntriesByWorkspaceResult struct {
		TotalCount int64
		Entries    []models.Entry
	}

	DeleteEntriesArgs struct {
		Entries []pgtype.UUID `json:"entries"`
	}

	EntryOwnership struct {
		EntryID  pgtype.UUID
		OwnerID  int32
		IsOwner  bool
		UserRole rbac.Role
	}

	NewChunkArgs struct {
		EntryID        pgtype.UUID
		Index          int32
		MinimumVersion int32
		Content        string
		Language       string
	}

	UpdateQueueArgs struct {
		Status  queries.EntryStatus
		EntryID int32
	}

	UpdateEntryArgs struct {
		Context         context.Context
		PublicID        pgtype.UUID
		Name            string
		MarkdownContent string
		// TextContent is the plain text content of the entry, used for search indexing
		TextContent string
		Checksum    string
	}

	CreateChunksArgs struct {
		Context context.Context
		Chunks  []NewChunkArgs
	}

	RequeueEntriesArgs struct {
		WorkspaceID    pgtype.UUID
		EntryPublicIDs []pgtype.UUID
	}

	FindbyIdArgs struct {
		InternalID int32
		PublicID   pgtype.UUID
		Workspace  PublicIdOrSlug
		Collection PublicIdOrSlug
	}

	UnindexedChunk struct {
		ID      int32
		Content string
	}

	UpdateChunkSemanticVectorArgs struct {
		ChunkID int32
		Vector  []float32
		Status  queries.EntryChunkEmbeddingStatus
		Error   error
	}

	HybridSearchArgs struct {
		Context        context.Context
		TextQuery      string
		SemanticVector []float32

		UserID     int32
		Workspace  PublicIdOrSlug
		Pagination PaginationParams
	}

	// Cache entry
	EntryRepository interface {
		// SaveLinkMetadata caches metadata for a given URL for 24 hours
		SaveLinkMetadata(url string, metadata *ograph.Metadata) error

		// GetLinkMetadata retrieves cached metadata for a given URL if it exists
		GetLinkMetadata(url string) (metadata ograph.Metadata, exists bool, err error)

		// CreateLinkEntry creates a new link entry
		CreateLinkEntry(entry *models.LinkEntry) (models.CreatedEntry, error)

		// CreateFileEntry creates a new file entry
		CreateFileEntry(entry *models.FileEntry) (models.CreatedEntry, error)

		// EnqueueEntry enqueues an entry for processing
		EnqueueEntries(entries []EnqueueEntryParams) error

		// DequeueEntries dequeues entries
		DequeueEntries(entries []pgtype.UUID) ([]int32, error)

		// FindByID finds an entry by its ID
		FindByID(args *FindbyIdArgs) (models.Entry, error)

		// FindQueuedIDs returns the IDs of all valid entries in the queue
		FindQueuedIDs() ([]int32, error)

		// Find all entries in all collections in a workspace (and optionally a collection) with pagination
		FindAllWithPagination(
			args *FindEntriesArgs,
			pagination PaginationParams,
		) (FindEntriesByWorkspaceResult, error)

		// UpdateEntry updates an existing entry
		UpdateEntry(args *UpdateEntryArgs) error

		// Delete multiple entries
		Delete(args DeleteEntriesArgs) ([]pgtype.UUID, error)

		// GetOwnership returns the ownership status of entries for a given user
		GetOwnerships(userID int32, entryIDs []pgtype.UUID) ([]EntryOwnership, error)

		// CreateChunk creates a new chunk for an entry
		CreateChunk(args *NewChunkArgs) (*models.Chunk, error)

		// CreateChunks creates multiple chunks for an entry
		CreateChunks(args *CreateChunksArgs) (int64, error)

		// UpdateChunk updates a chunk
		UpdateChunk(chunk *models.Chunk) (*models.Chunk, error)

		// UpdateQueue updates the queue record for a given entry
		UpdateQueue(args *UpdateQueueArgs) error

		// RequeueEntries deletes all existing chunks matching the minimum version for the given entries and resets the queue status
		RequeueEntries(args *RequeueEntriesArgs) ([]int32, error)

		// FindUnindexedChunks returns all entries that are not indexed yet
		FindUnindexedChunks() ([]UnindexedChunk, error)

		// FindUnindexedEntryChunks returns all chunks for a given entry that are not indexed yet
		FindUnindexedEntryChunks(entryId pgtype.UUID) ([]UnindexedChunk, error)

		// UpdateSemanticVectorState updates the state of a chunk
		UpdateSemanticVectorState(args *UpdateChunkSemanticVectorArgs) error

		// CanEmbedChunk checks if a chunk can be processed for semantic embedding
		CanEmbedChunk(id int32) (bool, error)

		// QueryWithHybridSearch searches for a query using hybrid search
		QueryWithHybridSearch(args *HybridSearchArgs) (*models.HybridSearchResults, error)
	}

	entryRepo struct {
		*baseRepo
	}
)

/*
QueryWithHybridSearch searches for a query using hybrid search

We do this in the following way:

  - First we search from the database using the embedding vector (if present) and full text search (concurrently)

  - The we normalize the scores of the results to be between 0 and 1 (this is important for the hybrid score we will calculate later on). This is done using the `Normalize` function.

  - We calculate the `hybrid score` for each result using the following formula:

    hybrid_score = (semantic_score * DefaultSemanticWeight) + (text_score * DefaultFullTextWeight) + KeywordScore

  - We merge the results from both searches into a single list

  - We sort the results by the hybrid score

  - Finally we return the results

NOTE: while fuzzy search *could* increase the results, it is not used here because it is not very performant and somewhat unnecessary at the moment
*/
func (e *entryRepo) QueryWithHybridSearch(
	args *HybridSearchArgs,
) (*models.HybridSearchResults, error) {
	ctx := args.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	searchResult := make([]models.SearchResult, 0)

	// Search with semantic vector
	rows, err := e.queries.QueryWithHybridSearch(
		ctx,
		queries.QueryWithHybridSearchParams{
			Embedding:         pgvector.NewVector(args.SemanticVector),
			WorkspacePublicID: args.Workspace.PublicID,
			WorkspaceSlug:     lib.PgText(args.Workspace.Slug),
			UserID:            args.UserID,
			Query:             args.TextQuery,
			Limit:             args.Pagination.Limit(),
			Offset:            args.Pagination.Offset(),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &models.HybridSearchResults{}, nil
		}
		return nil, seer.Wrap("query_with_semantic_vectors", err)
	}

	for i := range rows {
		row := &rows[i]
		meta, _ := models.UnmarshalEntryMetadata(row.Meta, document.EntryType(row.Type))
		searchResult = append(searchResult, models.SearchResult{
			ID:      row.PublicID,
			Rank:    float32(row.Rank),
			Name:    row.Title,
			Preview: row.ChunkContent.String,
			Type:    document.EntryType(row.Type),
			Status:  row.Status,
			Source:  models.SearchTypeSemantic,
			Chunk: models.SearchResultChunkMetadata{
				ID:    row.ChunkID,
				Index: row.ChunkIndex,
			},
			Collection: models.EntryRelation{
				ID:   row.CollectionID,
				Name: row.CollectionName,
				Slug: row.CollectionSlug.String,
			},
			Workspace: models.EntryRelation{
				ID:   row.WorkspaceID,
				Name: row.WorkspaceName,
				Slug: row.WorkspaceSlug.String,
			},
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			ArchivedAt:    row.ArchivedAt.Time,
			HybridScore:   row.Score,
			TextScore:     row.TextScore.Float64,
			SemanticScore: row.SemanticScore,
			MatchedBy:     []models.SearchSource{},
			Metadata:      meta,
			FileID:        row.FileID.String,
			FilesizeBytes: row.FilesizeBytes,
		})
	}

	return e.RerankResults(args.TextQuery, searchResult), nil
}

func (e *entryRepo) RerankResults(
	query string,
	results []models.SearchResult,
) *models.HybridSearchResults {
	dedupMap := map[int32]*models.SearchResult{}

	addMatchedBy := func(result *models.SearchResult, searchType models.SearchSource) {
		if !slices.Contains(result.MatchedBy, searchType) {
			result.MatchedBy = append(result.MatchedBy, searchType)
		}
	}

	containskeyword := func(result *models.SearchResult, query string) bool {
		return strings.Contains(strings.ToLower(result.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(result.Preview), strings.ToLower(query))
	}

	// Merge vector results into the dedup map
	for i := range results {
		result := &results[i]

		if result.TextScore != 0.0 {
			addMatchedBy(result, models.SearchTypeFullText)
		}

		if result.SemanticScore != 0.0 {
			addMatchedBy(result, models.SearchTypeSemantic)
		}

		if containskeyword(result, query) {
			result.HybridScore += KeywordScore
			addMatchedBy(result, models.SearchTypeKeyword)
		}

		dedupMap[result.Chunk.ID] = result
	}

	// Group by entry ID
	resultsMap, minHybridScore, maxHybridScore := groupByEntry(dedupMap)

	// Collect into a slice
	sliceResult := collectIntoSortedSlice(resultsMap)

	// Calculate relevance percents
	calculateRelevancePercentile(sliceResult)

	return &models.HybridSearchResults{
		CollapsedResults: sliceResult,
		MinHybridScore:   minHybridScore,
		MaxHybridScore:   maxHybridScore,
	}
}

// FindUnindexedEntryChunks implements EntryRepository.
func (e *entryRepo) FindUnindexedEntryChunks(entryId pgtype.UUID) ([]UnindexedChunk, error) {
	rows, err := e.queries.FindEntryChunks(context.TODO(), entryId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []UnindexedChunk{}, nil
		}
		return nil, err
	}

	unindexed := make([]UnindexedChunk, 0, len(rows))
	for _, row := range rows {
		unindexed = append(unindexed, UnindexedChunk{ID: row.ID, Content: row.Content.String})
	}

	return unindexed, nil
}

// CanEmbedChunk implements EntryRepository.
func (e *entryRepo) CanEmbedChunk(id int32) (bool, error) {
	return e.queries.ChunkCanBeProcessed(context.TODO(), id)
}

// FindUnindexedChunks implements EntryRepository.
func (e *entryRepo) FindUnindexedChunks() ([]UnindexedChunk, error) {
	rows, err := e.queries.FindUnindexedChunks(context.TODO())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []UnindexedChunk{}, nil
		}
		return nil, err
	}

	unindexed := make([]UnindexedChunk, 0, len(rows))
	for _, row := range rows {
		unindexed = append(unindexed, UnindexedChunk{ID: row.ID, Content: row.Content.String})
	}

	return unindexed, nil
}

// UpdateSemanticVectorState implements EntryRepository.
func (e *entryRepo) UpdateSemanticVectorState(args *UpdateChunkSemanticVectorArgs) error {
	var errorMessage string
	ctx := context.TODO()

	switch args.Status {
	case queries.EntryChunkEmbeddingStatusDone:
		if len(args.Vector) == 0 {
			return seer.New("update_semantic_vector", "empty vector")
		}

		return e.queries.UpdateChunkSemanticVector(ctx, queries.UpdateChunkSemanticVectorParams{
			SemanticVector: pgvector.NewVector(args.Vector),
			ChunkID:        args.ChunkID,
		})

	case queries.EntryChunkEmbeddingStatusFailed:
		errorMessage = "an untracked error occurred"
		if args.Error != nil {
			errorMessage = args.Error.Error()
		}
		fallthrough

	default:

		return e.queries.UpdateChunkEmbeddingStatus(ctx, queries.UpdateChunkEmbeddingStatusParams{
			EmbeddingStatus:    args.Status,
			LastEmbeddingError: lib.PgText(errorMessage),
			ChunkID:            args.ChunkID,
		})
	}
}

// RequeueEntries implements EntryRepository.
func (e *entryRepo) RequeueEntries(args *RequeueEntriesArgs) ([]int32, error) {
	tx, err := e.pool.BeginTx(context.TODO(), pgx.TxOptions{}) //nolint:all
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.TODO()) //nolint:errcheck

	queriesWithTx := e.queries.WithTx(tx)

	// Delete all chunks for the given entries
	ids, err := queriesWithTx.DeleteEntryChunksByPublicId(
		context.TODO(),
		queries.DeleteEntryChunksByPublicIdParams{
			EntryPublicIds: args.EntryPublicIDs,
			WorkspaceID:    args.WorkspaceID,
		},
	)
	if err != nil {
		return nil, seer.Wrap("delete_entry_chunks", err)
	}

	// If ids is empty, there is a good chance that they do not have existing chunks
	if len(ids) != len(args.EntryPublicIDs) {
		resolved, err := queriesWithTx.ResolveEntryIdsWithoutChunks(
			context.TODO(),
			args.EntryPublicIDs,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, seer.Wrap("resolve_entry_ids_without_chunks", err)
		}
		ids = make([]int32, 0, len(resolved))
		for _, id := range resolved {
			ids = append(ids, id.ID)
		}
	}

	// Reset the queue status for the given entries
	updated, err := queriesWithTx.RequeueEntries(context.TODO(), ids)
	if err != nil {
		return nil, seer.Wrap("requeue_entries", err)
	}

	if err := tx.Commit(context.TODO()); err != nil {
		return nil, seer.Wrap("commit_tx", err)
	}

	return updated, nil
}

// UpdateEntry implements EntryRepository.
func (e *entryRepo) UpdateEntry(args *UpdateEntryArgs) error {
	ctx := args.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	_, err := e.queries.UpdateEntry(ctx, queries.UpdateEntryParams{
		Name:        lib.PgText(args.Name),
		Content:     lib.PgText(args.MarkdownContent),
		TextContent: lib.PgText(args.TextContent),
		Checksum:    lib.PgText(args.Checksum),
		PublicID:    args.PublicID,
	})
	return err
}

// FindByID implements EntryRepository.
func (e *entryRepo) FindByID(args *FindbyIdArgs) (models.Entry, error) {
	row, err := e.queries.FindEntryById(context.TODO(), queries.FindEntryByIdParams{
		EntryPublicID:      args.PublicID,
		EntryID:            lib.PgInt4(args.InternalID),
		WorkspaceSlug:      lib.PgText(args.Workspace.Slug),
		WorkspacePublicID:  args.Workspace.PublicID,
		CollectionSlug:     lib.PgText(args.Collection.Slug),
		CollectionPublicID: args.Collection.PublicID,
	})
	if err != nil {
		return models.Entry{}, err
	}

	meta, _ := models.UnmarshalEntryMetadata(row.Meta, row.Type)

	return models.Entry{
		ID:            row.ID,
		PublicID:      row.PublicID,
		OriginID:      row.Origin,
		Name:          row.Name,
		Content:       row.Content,
		TextContent:   row.TextContent,
		Version:       row.Version,
		Type:          row.Type,
		ParentID:      row.ParentID,
		FileID:        row.FileID.String,
		FilesizeBytes: row.FilesizeBytes,
		Status:        row.Status,
		QueuedAt:      row.QueuedAt.Time,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
		ArchivedAt:    row.ArchivedAt.Time,
		AddedBy: models.EntryAddedBy{
			FirstName: row.AddedByFirstName,
			LastName:  row.AddedByLastName,
			Username:  row.AddedByUsername,
		},
		Collection: models.EntryRelation{
			ID:   row.CollectionID,
			Name: row.CollectionName,
			Slug: row.CollectionSlug.String,
		},
		Workspace: models.EntryRelation{
			ID:   row.WorkspaceID,
			Name: row.WorkspaceName,
			Slug: row.WorkspaceSlug.String,
		},
		Metadata: meta,
	}, nil
}

// UpdateQueue implements EntryRepository.
func (e *entryRepo) UpdateQueue(args *UpdateQueueArgs) error {
	err := e.queries.UpdateEntryStatus(context.TODO(), queries.UpdateEntryStatusParams{
		Status:  args.Status,
		EntryID: args.EntryID,
	})
	if err != nil {
		return seer.Wrap("update_entry_status", err)
	}

	return nil
}

// FindQueuedIDs implements EntryRepository.
func (e *entryRepo) FindQueuedIDs() ([]int32, error) {
	rows, err := e.queries.FindAllQueuedEntries(context.TODO())
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []int32{}, nil
		}
		return nil, err
	}

	return rows, nil
}

// CreateChunk implements EntryRepository.
func (e *entryRepo) CreateChunk(args *NewChunkArgs) (*models.Chunk, error) {
	row, err := e.queries.InsertChunk(context.TODO(), queries.InsertChunkParams{
		EntryPublicID: args.EntryID,
		Index:         args.Index,
		MinVersion:    args.MinimumVersion,
		Content:       lib.PgText(args.Content),
		Language:      lib.PgText(args.Language),
	})
	if err != nil {
		return &models.Chunk{}, err
	}

	chunk := new(models.Chunk)
	chunk.From(args.EntryID, &row)

	return chunk, nil
}

func (e *entryRepo) CreateChunks(args *CreateChunksArgs) (int64, error) {
	ctx := args.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	tx, err := e.pool.BeginTx(ctx, pgx.TxOptions{}) //nolint:all
	if err != nil {
		return 0, seer.Wrap("begin_tx", err)
	}
	defer tx.Rollback(context.TODO()) //nolint:errcheck

	queriesWithTx := e.queries.WithTx(tx)

	idsMap := map[pgtype.UUID]int32{}
	ids := make([]pgtype.UUID, 0, len(idsMap))
	for _, arg := range args.Chunks {
		_, ok := idsMap[arg.EntryID]
		if ok {
			continue
		}

		idsMap[arg.EntryID] = 0
		ids = append(ids, arg.EntryID)
	}

	resolvedIds, err := queriesWithTx.ResolveEntryIds(ctx, ids)
	if err != nil {
		log.Error().Err(err).Msg("failed to resolve entry ids")
		return 0, seer.Wrap("resolve_entry_ids", err)
	}

	for _, resolved := range resolvedIds {
		if resolved.PublicID.Valid {
			idsMap[resolved.PublicID] = resolved.ID
		}
	}

	params := make([]queries.InsertChunksParams, 0)
	for _, arg := range args.Chunks {
		id := idsMap[arg.EntryID]
		minVersion := arg.MinimumVersion
		if minVersion == 0 {
			minVersion = 1
		}
		params = append(params, queries.InsertChunksParams{
			EntryID:    lib.PgInt4(id),
			Index:      arg.Index,
			MinVersion: minVersion,
			Content:    lib.PgText(arg.Content),
			Language:   lib.PgText(arg.Language),
		})
	}

	count, err := queriesWithTx.InsertChunks(ctx, params)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert chunks")
		return 0, seer.Wrap("insert_chunks", err)
	}

	if err := tx.Commit(context.TODO()); err != nil {
		return 0, seer.Wrap("commit_tx", err)
	}

	return count, nil
}

func (e *entryRepo) UpdateChunk(chunk *models.Chunk) (*models.Chunk, error) {
	row, err := e.queries.UpdateChunk(context.TODO(), queries.UpdateChunkParams{
		Content:  lib.PgText(chunk.Content),
		Language: lib.PgText(chunk.Language),
		EntryID:  lib.PgInt4(chunk.EntryID),
		Index:    chunk.Index,
		Version:  chunk.MinVersion,
	})
	if err != nil {
		return &models.Chunk{}, err
	}

	chunk.From(chunk.EntryPublicID, &row)
	return chunk, nil
}

// GetOwnership implements EntryRepository.
func (e *entryRepo) GetOwnerships(userID int32, entryIDs []pgtype.UUID) ([]EntryOwnership, error) {
	rows, err := e.queries.GetEntriesOwnership(context.TODO(), queries.GetEntriesOwnershipParams{
		UserID:         userID,
		EntryPublicIds: entryIDs,
	})
	if err != nil {
		return nil, err
	}

	ownership := make([]EntryOwnership, 0)
	for _, row := range rows {
		ownership = append(ownership, EntryOwnership{
			EntryID:  row.PublicID,
			OwnerID:  row.AddedBy,
			IsOwner:  row.IsOwner,
			UserRole: row.UserRole,
		})
	}

	return ownership, nil
}

// Delete implements EntryRepository.
func (e *entryRepo) Delete(args DeleteEntriesArgs) ([]pgtype.UUID, error) {
	tx, err := e.pool.BeginTx(context.TODO(), pgx.TxOptions{}) //nolint:all
	if err != nil {
		return nil, seer.Wrap("begin_tx", err)
	}

	q := e.queries.WithTx(tx)

	// Remove all queued entries first
	if _, err = q.DequeueEntries(context.TODO(), args.Entries); err != nil {
		return nil, seer.Wrap("dequeue_entries", tx.Rollback(context.TODO()))
	}

	// Delete entries
	deletedIds, err := q.DeleteEntries(context.TODO(), args.Entries)
	if err != nil {
		return nil, seer.Wrap("delete_entries", tx.Rollback(context.TODO()))
	}

	if err := tx.Commit(context.TODO()); err != nil {
		return nil, seer.Wrap("commit_tx", err)
	}

	return deletedIds, nil
}

// FindByWorkspaceID implements EntryRepository.
func (e *entryRepo) FindAllWithPagination(
	args *FindEntriesArgs,
	pagination PaginationParams,
) (FindEntriesByWorkspaceResult, error) {
	result := FindEntriesByWorkspaceResult{
		TotalCount: 0,
		Entries:    make([]models.Entry, 0),
	}

	rows, err := e.queries.FindEntries(context.TODO(), queries.FindEntriesParams{
		Limit:              pagination.Limit(),
		Offset:             pagination.Offset(),
		WorkspacePublicID:  args.Workspace.PublicID,
		WorkspaceSlug:      lib.PgText(args.Workspace.Slug),
		CollectionPublicID: args.Collection.PublicID,
		CollectionSlug:     lib.PgText(args.Collection.Slug),
		UserID:             args.UserID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return result, nil
		}

		return result, err
	}

	for i := range rows {
		row := &rows[i]
		if result.TotalCount == 0 {
			result.TotalCount = row.TotalEntries
		}

		meta, _ := models.UnmarshalEntryMetadata(row.Meta, document.EntryType(row.Type))

		result.Entries = append(result.Entries, models.Entry{
			ID:            row.ID,
			PublicID:      row.PublicID,
			OriginID:      row.Origin,
			Name:          row.Name,
			Content:       row.Content,
			TextContent:   row.TextContent,
			Version:       row.Version,
			Type:          document.EntryType(row.Type),
			ParentID:      row.ParentID,
			FileID:        row.FileID.String,
			FilesizeBytes: row.FilesizeBytes,
			Status:        row.Status,
			QueuedAt:      row.QueuedAt.Time,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
			ArchivedAt:    row.ArchivedAt.Time,
			Metadata:      meta,
			AddedBy: models.EntryAddedBy{
				FirstName: row.AddedByFirstName,
				LastName:  row.AddedByLastName,
				Username:  row.AddedByUsername,
			},
			Collection: models.EntryRelation{
				ID:   row.CollectionID,
				Name: row.CollectionName,
				Slug: row.CollectionSlug.String,
			},
			Workspace: models.EntryRelation{
				ID:   row.WorkspaceID,
				Name: row.WorkspaceName,
				Slug: row.WorkspaceSlug.String,
			},
		})
	}

	result.Entries = lib.WithMaxSize(result.Entries, pagination.PerPage)

	return result, nil
}

func (e *entryRepo) EnqueueEntries(entries []EnqueueEntryParams) error {
	items := make([]queries.EnqueueEntriesParams, 0, len(entries))
	for _, entry := range entries {
		items = append(items, queries.EnqueueEntriesParams{
			Payload: entry.Payload,
			EntryID: entry.ID,
		})
	}

	_, err := e.queries.EnqueueEntries(context.TODO(), items)
	return err
}

func (e *entryRepo) DequeueEntries(entries []pgtype.UUID) ([]int32, error) {
	deleted, err := e.queries.DequeueEntries(context.TODO(), entries)
	return deleted, err
}

// CreateLinkEntry implements EntryRepository.
func (e *entryRepo) CreateLinkEntry(entry *models.LinkEntry) (models.CreatedEntry, error) {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return models.CreatedEntry{}, seer.Wrap("marshal_link_metadata", err)
	}

	created, err := e.queries.CreateLinkEntry(context.TODO(), queries.CreateLinkEntryParams{
		Name:         strings.TrimSpace(entry.Title),
		Meta:         meta,
		CollectionID: entry.CollectionID,
		UserID:       entry.UserID,
	})
	if err != nil {
		return models.CreatedEntry{}, err
	}

	return models.CreatedEntry{
		ID:         created.PublicID,
		InternalID: created.ID,
		Name:       created.Name,
		Type:       created.EntryType,
	}, nil
}

func (e *entryRepo) CreateFileEntry(entry *models.FileEntry) (models.CreatedEntry, error) {
	meta, err := json.Marshal(models.FileMetadata{
		Extension:        path.Ext(entry.OriginalName),
		OriginalFilename: entry.OriginalName,
		MimeType:         entry.MimeType,
		ExtraMetadata:    json.RawMessage{},
	})
	if err != nil {
		return models.CreatedEntry{}, seer.Wrap("marshal_file_metadata", err)
	}

	created, err := e.queries.CreateFileEntry(context.TODO(), queries.CreateFileEntryParams{
		Name:          lib.StripExtension(strings.TrimSpace(entry.OriginalName)),
		Meta:          meta,
		FileID:        pgtype.Text{String: entry.FileID, Valid: true},
		EntryType:     entry.Type,
		CollectionID:  entry.CollectionID,
		AddedBy:       entry.UserID,
		FilesizeBytes: entry.Filesize,
		Content:       pgtype.Text{}, //nolint:exhaustruct
		Checksum:      pgtype.Text{}, //nolint:exhaustruct
	})
	if err != nil {
		return models.CreatedEntry{}, err
	}

	return models.CreatedEntry{
		ID:         created.PublicID,
		InternalID: created.ID,
		Name:       created.Name,
		Type:       created.EntryType,
	}, nil
}

// SaveLinkMetadata implements EntryRepository.
func (e *entryRepo) SaveLinkMetadata(url string, metadata *ograph.Metadata) error {
	return e.store.SetJsonWithTTL(kv.KeyLinkMetadata(url), *metadata, time.Hour*24)
}

// GetLinkMetadata implements EntryRepository.
func (e *entryRepo) GetLinkMetadata(url string) (ograph.Metadata, bool, error) {
	exists, err := e.store.Exists(kv.KeyLinkMetadata(url))
	if err != nil {
		return ograph.Metadata{}, false, err
	}

	if !exists {
		//nolint:exhaustruct
		return ograph.Metadata{}, false, nil
	}

	var metadata ograph.Metadata
	if err := e.store.GetJson(kv.KeyLinkMetadata(url), &metadata); err != nil {
		// Delete the key if it fails to unmarshal
		if err := e.store.Delete(kv.KeyLinkMetadata(url)); err != nil {
			return ograph.Metadata{}, false, err
		}

		return ograph.Metadata{}, false, err
	}

	return metadata, true, nil
}

var _ EntryRepository = (*entryRepo)(nil)
