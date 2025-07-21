package api

import (
	"io"

	"github.com/adelowo/gulter"
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/ograph"
	"go.trulyao.dev/robin"
)

type (
	EntryHandler interface {
		// Import an entry
		Import(ctx *robin.Context, body io.ReadCloser) (ImportEntryResponse, error)

		// GetLinkMetadata returns the parsed OpenGraph metadata for a given link
		GetLinkMetadata(ctx *robin.Context, link string) (ograph.Metadata, error)

		// FindWorkspaceEntries returns all entries in a workspace with pagination
		FindWorkspaceEntries(
			ctx *robin.Context,
			request FindWorkspaceEntriesRequest,
		) (FindEntriesResponse, error)

		// FindCollectionEntries returns all entries in a collection with pagination
		FindCollectionEntries(
			ctx *robin.Context,
			request FindCollectionEntriesRequest,
		) (FindEntriesResponse, error)

		// Delete multiple entries
		Delete(ctx *robin.Context, request DeleteEntriesRequest) (DeleteEntriesResponse, error)

		// Requeue multiple entries
		Requeue(ctx *robin.Context, request RequeueEntriesRequest) (RequeueEntriesResponse, error)

		// Find an entry by ID
		Find(ctx *robin.Context, request FindEntryRequest) (FindEntryResponse, error)

		// Search using full-text and/or vectors
		Search(ctx *robin.Context, request SearchRequest) (SearchResponse, error)
	}

	DeleteEntriesRequest struct {
		WorkspaceSlug string   `json:"workspace_slug" validate:"required,slug"`
		Entries       []string `json:"entry_ids"      validate:"required,min=1,dive,uuid"`
	}

	DeleteEntriesResponse struct {
		DeletedEntries []pgtype.UUID `json:"deleted_entries" mirror:"type:Array<string>"`
		Message        string        `json:"message"`
	}

	ImportEntryPayload struct {
		CollectionID pgtype.UUID   `json:"collection_id" mirror:"type:string"`
		WorkspaceID  pgtype.UUID   `json:"workspace_id"  mirror:"type:string"`
		Links        []string      `json:"links"`
		Files        []gulter.File `json:"files"         mirror:"type:Array<File>"`
	}

	ImportLinksPayload struct {
		Links        []string `json:"links"`
		CollectionID int32    `json:"collection_id"`
		UserID       int32    `json:"user_id"`
	}

	ImportFilesPayload struct {
		Files        []gulter.File `json:"files"         mirror:"type:Array<File>"`
		CollectionID int32         `json:"collection_id"`
		UserID       int32         `json:"user_id"`
	}

	ImportEntryResponse struct {
		WorkspaceID  pgtype.UUID           `json:"workspace_id"  mirror:"type:string"`
		CollectionID pgtype.UUID           `json:"collection_id" mirror:"type:string"`
		Entries      []models.CreatedEntry `json:"entries"`
	}

	FindWorkspaceEntriesRequest struct {
		Pagination    repository.PaginationParams `json:"pagination"`
		WorkspaceSlug string                      `json:"workspace_slug" validate:"required,slug"`
	}

	FindCollectionEntriesRequest struct {
		Pagination     repository.PaginationParams `json:"pagination"`
		CollectionSlug string                      `json:"collection_slug" validate:"required,slug"`
		WorkspaceSlug  string                      `json:"workspace_slug"  validate:"required,slug"`
	}

	FindEntriesResponse struct {
		Entries       []models.Entry             `json:"entries"`
		WorkspaceSlug string                     `json:"workspace_slug"`
		Pagination    repository.PaginationState `json:"pagination"`
	}

	RequeueEntriesRequest struct {
		WorkspaceID string   `json:"workspace_id" validate:"required,uuid"`
		EntryIDs    []string `json:"entry_ids"    validate:"required,min=1,dive,uuid"`
	}

	RequeueEntriesResponse struct {
		WorkspaceSlug string `json:"workspace_slug"`
		Count         int    `json:"count"`
	}

	FindEntryRequest struct {
		EntryID        string `json:"entry_id"        validate:"required,uuid"`
		WorkspaceSlug  string `json:"workspace_slug"  validate:"required,slug"`
		CollectionSlug string `json:"collection_slug" validate:"required,slug"`
	}

	FindEntryResponse struct {
		Entry models.Entry `json:"entry"`
	}

	SearchRequest struct {
		WorkspaceID   string `json:"workspace_id"   validate:"required_without=WorkspaceSlug" mirror:"optional:true"`
		WorkspaceSlug string `json:"workspace_slug" validate:"optional_slug"                  mirror:"optional:true"`
		Query         string `json:"query"          validate:"required,ascii,min=2"`
	}

	SearchResponse struct {
		Results   models.HybridSearchResults `json:"results"`
		Query     string                     `json:"query"`
		TimeTaken int64                      `json:"time_taken_ms"`
	}
)
