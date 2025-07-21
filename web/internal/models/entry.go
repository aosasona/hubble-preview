package models

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/pkg/document"
	"go.trulyao.dev/hubble/web/pkg/ograph"
)

type SearchSource string

const (
	SearchTypeFullText SearchSource = "full_text"
	SearchTypeSemantic SearchSource = "semantic"
	SearchTypeKeyword  SearchSource = "keyword"
	SearchTypeFuzzy    SearchSource = "fuzzy"
)

type (
	CreatedEntry struct {
		ID         pgtype.UUID        `json:"id"`
		InternalID int32              `json:"-"`
		Name       string             `json:"name"`
		Type       document.EntryType `json:"type"`
	}

	LinkEntry struct {
		Title        string `json:"title"`
		Link         string `json:"link"`
		CollectionID int32  `json:"-"`
		UserID       int32  `json:"-"`
		Metadata     ograph.Metadata
	}

	FileEntry struct {
		FileID       string             `json:"file_id"`
		OriginalName string             `json:"name"`
		SavedName    string             `json:"saved_filename"`
		Filesize     int64              `json:"filesize_bytes"`
		MimeType     string             `json:"mime_type"`
		Type         document.EntryType `json:"type"`
		CollectionID int32              `json:"-"`
		UserID       int32              `json:"-"`
	}

	FileMetadata struct {
		OriginalFilename string `json:"original_filename"`
		MimeType         string `json:"mime_type"`
		Extension        string `json:"extension"`
		// ExtraMetadata is a JSON string that can be used to store additional metadata by plugins and other things that need to store unstructured data
		ExtraMetadata json.RawMessage `json:"extra_metadata,omitempty"`
	}

	EntryAddedBy struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
	}

	EntryRelation struct {
		ID   pgtype.UUID `json:"id"   mirror:"type:string"`
		Name string      `json:"name"`
		Slug string      `json:"slug"`
	}

	Entry struct {
		ID            int32              `json:"-"`
		PublicID      pgtype.UUID        `json:"id"             mirror:"type:string"`
		OriginID      pgtype.UUID        `json:"origin"         mirror:"type:string"`
		Name          string             `json:"name"`
		Content       pgtype.Text        `json:"content"        mirror:"type:string"`
		TextContent   pgtype.Text        `json:"text_content"   mirror:"type:string"`
		Version       int32              `json:"version"`
		Type          document.EntryType `json:"type"           mirror:"type:'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other'"`
		ParentID      pgtype.Int4        `json:"parent_id"      mirror:"type:number"`
		FileID        string             `json:"file_id"        mirror:"optional:true"`
		FilesizeBytes int64              `json:"filesize_bytes"`

		Status   queries.EntryStatus `json:"status"    mirror:"type:'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused'"`
		QueuedAt time.Time           `json:"queued_at"`

		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
		ArchivedAt time.Time `json:"archived_at"`

		AddedBy EntryAddedBy `json:"added_by"`

		Collection EntryRelation `json:"collection"`
		Workspace  EntryRelation `json:"workspace"`

		Metadata any `json:"metadata" mirror:"type:import('./types').FileMetadata | import('./types').Metadata"`
	}

	Chunk struct {
		ID            int32       `json:"id"`
		EntryID       int32       `json:"entry_id"`
		EntryPublicID pgtype.UUID `json:"entry_public_id" mirror:"type:string,optional:true"`
		Index         int32       `json:"index"`
		MinVersion    int32       `json:"min_version"`
		Content       string      `json:"content"`
		CreatedAt     time.Time   `json:"created_at"`
		UpdatedAt     time.Time   `json:"updated_at"`
		DeletedAt     time.Time   `json:"deleted_at"`
		Language      string      `json:"language"`
		// The full-text vector of the chunk
		TextVector any `json:"-"`
		// The semantic vector of the chunk using an embedding model
		SemanticVector pgvector.Vector `json:"-"`
	}

	SearchResultChunkMetadata struct {
		ID    int32 `json:"id"`
		Index int32 `json:"index"`
	}

	SearchResult struct {
		ID            pgtype.UUID               `json:"id"             mirror:"type:string"`
		Rank          float32                   `json:"rank"`
		Name          string                    `json:"name"`
		Preview       string                    `json:"preview"`
		Type          document.EntryType        `json:"type"           mirror:"type:'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other'"`
		Source        SearchSource              `json:"search_type"    mirror:"type:'full_text' | 'semantic'"`
		Status        queries.EntryStatus       `json:"status"         mirror:"type:'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused'"`
		MatchedBy     []SearchSource            `json:"matched_by"     mirror:"type:Array<'full_text' | 'semantic' | 'keyword'>"`
		TextScore     float64                   `json:"text_score"`
		SemanticScore float64                   `json:"semantic_score"`
		HybridScore   float64                   `json:"hybrid_score"`
		Chunk         SearchResultChunkMetadata `json:"chunk"`
		Metadata      any                       `json:"metadata"       mirror:"type:import('./types').FileMetadata | import('./types').Metadata"`
		FileID        string                    `json:"file_id"        mirror:"optional:true"`
		FilesizeBytes int64                     `json:"filesize_bytes"`

		Collection EntryRelation `json:"collection"`
		Workspace  EntryRelation `json:"workspace"`
		CreatedAt  time.Time     `json:"created_at"`
		UpdatedAt  time.Time     `json:"updated_at"`
		ArchivedAt time.Time     `json:"archived_at"`
	}

	MatchedChunk struct {
		ID            int32   `json:"id"`
		Index         int32   `json:"index"`
		Text          string  `json:"text"`
		Rank          float32 `json:"rank"`
		TextScore     float64 `json:"text_score"`
		SemanticScore float64 `json:"semantic_score"`
		HybridScore   float64 `json:"hybrid_score"`
	}

	CollapsedSearchResult struct {
		ID               pgtype.UUID         `json:"id"                mirror:"type:string"`
		Name             string              `json:"name"`
		Type             document.EntryType  `json:"type"              mirror:"type:'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other'"`
		Matches          []MatchedChunk      `json:"matches"`
		Status           queries.EntryStatus `json:"status"            mirror:"type:'queued' | 'processing' | 'completed' | 'failed' | 'canceled' | 'paused'"`
		FileID           string              `json:"file_id"           mirror:"optional:true"`
		FilesizeBytes    int64               `json:"filesize_bytes"`
		RelevancePercent float64             `json:"relevance_percent"`
		Collection       EntryRelation       `json:"collection"`
		Workspace        EntryRelation       `json:"workspace"`
		CreatedAt        time.Time           `json:"created_at"`
		UpdatedAt        time.Time           `json:"updated_at"`
		ArchivedAt       time.Time           `json:"archived_at"`
		Metadata         any                 `json:"metadata"          mirror:"type:import('./types').FileMetadata | import('./types').Metadata"`
	}

	HybridSearchResults struct {
		CollapsedResults []CollapsedSearchResult `json:"results"`
		MinHybridScore   float64                 `json:"min_hybrid_score"`
		MaxHybridScore   float64                 `json:"max_hybrid_score"`
	}
)

func (c *Chunk) From(entryPublicId pgtype.UUID, chunk *queries.EntryChunk) {
	*c = Chunk{
		ID:             chunk.ID,
		EntryID:        chunk.EntryID.Int32,
		Index:          chunk.ChunkIndex,
		MinVersion:     chunk.MinVersion,
		Content:        chunk.Content.String,
		CreatedAt:      chunk.CreatedAt.Time,
		UpdatedAt:      chunk.UpdatedAt.Time,
		DeletedAt:      chunk.DeletedAt.Time,
		Language:       chunk.Language.String,
		TextVector:     chunk.TextVector,
		SemanticVector: chunk.SemanticVector,
		EntryPublicID:  entryPublicId,
	}
}

func UnmarshalEntryMetadata(data []byte, t document.EntryType) (any, error) {
	switch t {
	case document.EntryTypeLink:
		var meta ograph.Metadata
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil, err
		}
		return meta, nil
	default:
		var meta FileMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil, err
		}
		return meta, nil
	}
}
