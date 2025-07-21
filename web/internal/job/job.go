package job

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(entry,chunk_embedding,entry_chunk_embedding)
type JobType string

type Job interface {
	Type() JobType
}

type QueueFn func(Job) error

type (
	EntryJob struct {
		ID int32 `json:"entry_id"`
	}

	ChunkEmbeddingJob struct {
		ID      int32  `json:"id"`
		Content string `json:"content"`
	}

	EntryChunkEmbeddingJob struct {
		Entries []pgtype.UUID `json:"entries"`
	}
)

func (e *EntryJob) Type() JobType {
	return JobTypeEntry
}

func (e *EntryJob) Bytes() []byte {
	return fmt.Appendf(nil, `{"entry_id":%d}`, e.ID)
}

func (s *ChunkEmbeddingJob) Type() JobType {
	return JobTypeChunkEmbedding
}

func (s *ChunkEmbeddingJob) Bytes() []byte {
	m := make(map[string]any)
	m["id"] = s.ID
	m["content"] = s.Content

	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	return b
}

func (s *EntryChunkEmbeddingJob) Type() JobType {
	return JobTypeEntryChunkEmbedding
}

func (s *EntryChunkEmbeddingJob) Bytes() []byte {
	m := make(map[string]any)
	m["entries"] = s.Entries

	b, err := json.Marshal(m)
	if err != nil {
		return nil
	}

	return b
}
