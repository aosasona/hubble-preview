package boundhost

import (
	"context"
	"strings"

	"capnproto.org/go/capnp/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/hubble/web/schema"
)

/*
createEntryChunks creates new chunks for a given entry ID.

Signature: fn(request: Capnp::CreateChunksRequest) -> u64 (created_count)

Exported as: `entry_create_chunk`
*/
func (b *BoundHost) createEntryChunks(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	logger := b.HostFnLogger(spec.PermChunksCreate)

	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		logger.error(ErrFailedToReadMemory)
		return 0
	}
	defer b.dealloc(ctx, m, logger, offset, byteCount) // DEALLOCATE MEMORY

	msg, err := capnp.Unmarshal(buf)
	if err != nil {
		logger.error(ErrFailedToUnmarshalCapnp)
		return 0
	}

	request, err := schema.ReadRootCreateChunksRequest(msg)
	if err != nil {
		logger.errorf("failed to read entry chunk data")
		return 0
	}

	if !request.HasChunks() {
		logger.warnf("no chunks found")
		return 0
	}

	chunks, err := request.Chunks()
	if err != nil {
		logger.errorf("failed to read entry chunks")
		return 0
	}

	chunksToInsert := make([]repository.NewChunkArgs, 0)
	entryIds := make([]pgtype.UUID, 0)
	for i := range chunks.Len() {
		chunk := chunks.At(i)

		var (
			entryId  pgtype.UUID
			content  string
			language string
		)

		id, _ := chunk.EntryId()
		entryId = lib.PgUUIDString(id)

		content, _ = chunk.Content()
		language, _ = chunk.Language()

		chunksToInsert = append(chunksToInsert, repository.NewChunkArgs{
			EntryID:        entryId,
			Index:          chunk.Index(),
			MinimumVersion: chunk.MinimumVersion(),
			Content:        content,
			Language:       lib.NormalizePgLanguage(strings.ToLower(language)),
		})

		entryIds = append(entryIds, entryId)
	}

	createdCount, err := b.repository.EntryRepository().CreateChunks(
		&repository.CreateChunksArgs{Context: ctx, Chunks: chunksToInsert},
	)
	if err != nil {
		logger.errorf("failed to create entry chunks")
		return 0
	}

	// Queue the affected chunks for embedding
	go func(ids []pgtype.UUID) {
		if err := b.queueFn(&job.EntryChunkEmbeddingJob{Entries: ids}); err != nil {
			logger.errorf("failed to queue entry chunk embedding job: %v", err)
		}
	}(entryIds)

	return uint64(createdCount)
}

/*
updateEntry updates an entry with the given ID.

Signature: fn(request: Capnp::UpdateEntryRequest) -> Void

Exported as: `entry_update`
*/
func (b *BoundHost) updateEntry(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	logger := b.HostFnLogger(spec.PermEntriesUpdate)

	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		logger.error(ErrFailedToReadMemory)
		return 0
	}
	defer b.dealloc(ctx, m, logger, offset, byteCount) // DEALLOCATE MEMORY

	msg, err := capnp.Unmarshal(buf)
	if err != nil {
		logger.error(ErrFailedToUnmarshalCapnp)
		return 0
	}

	request, err := schema.ReadRootUpdateEntryRequest(msg)
	if err != nil {
		logger.error(ErrFailedToReadCapnp)
		return 0
	}

	var id, name, content, plainTextContent, checksum string

	if request.HasId() {
		id, err = request.Id()
		if err != nil {
			logger.errorf("failed to read entry id")
			return 0
		}
	}
	entryID := lib.PgUUIDString(id)

	if request.HasName() {
		name, _ = request.Name()
		name = strings.TrimSpace(name)
	}

	if request.HasMarkdownContent() {
		content, _ = request.MarkdownContent()
	}

	if request.HasPlainTextContent() {
		plainTextContent, _ = request.PlainTextContent()
	}

	if request.HasChecksum() {
		checksum, _ = request.Checksum()
	}

	if err := b.repository.EntryRepository().UpdateEntry(&repository.UpdateEntryArgs{
		Context:         ctx,
		PublicID:        entryID,
		Name:            name,
		MarkdownContent: content,
		TextContent:     plainTextContent,
		Checksum:        checksum,
	}); err != nil {
		logger.error(err, "failed to update entry in database")
	}

	return 0
}
