package boundhost

import (
	"bytes"
	"context"

	"capnproto.org/go/capnp/v3"
	"github.com/jonathanhecl/chunker"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/schema"
)

const (
	DefaultChunkSize    = 1_000
	DefaultChunkOverlap = 0.075 * DefaultChunkSize // 5% overlap
)

type ChunkMethod int

const (
	ChunkMethodSentence ChunkMethod = iota
	ChunkMethodOverlap
)

/*
Chunk the text based on the DefaultChunkSize and DefaultChunkOverlap.

The overlap A.K.A. a sliding window is used to ensure that chunks carry over some context from previous chunks.

Signature: fn(text: String) -> Capnp::ChunkResult

Exported as: `chunk_with_overlap`
*/
func (b *BoundHost) chunkWithOverlap(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	return b.chunk(ChunkMethodOverlap, ctx, m, offset, byteCount)
}

/*
Chunk the text by sentence.

Signature: fn(text: String) -> Capnp::ChunkResult

Exported as: `chunk_by_sentence`
*/
func (b *BoundHost) chunkBySentence(
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	return b.chunk(ChunkMethodSentence, ctx, m, offset, byteCount)
}

func (b *BoundHost) chunk(
	method ChunkMethod,
	ctx context.Context,
	m api.Module,
	offset, byteCount uint32,
) uint64 {
	fn := spec.PermTransformChunkBySentence
	if method == ChunkMethodOverlap {
		fn = spec.PermTransformChunkWithOverlap
	}
	logger := b.HostFnLogger(fn)

	buf, err := alloc.ReadBufferFromMemory(ctx, m, offset, byteCount)
	if err != nil {
		logger.error(err, "failed to read memory")
		return 0
	}

	text := string(buf)

	var chunks []string
	switch method {
	case ChunkMethodSentence:
		chunks = chunker.ChunkSentences(text)

	case ChunkMethodOverlap:
		c := chunker.NewChunker(
			DefaultChunkSize,
			DefaultChunkOverlap,
			chunker.DefaultSeparators,
			false,
			false,
		)
		chunks = c.Chunk(text)
	default:
		logger.errorf("unknown chunk method: %d", method)
		return 0
	}

	arena := capnp.SingleSegment(nil)
	msg, seg, err := capnp.NewMessage(arena)
	if err != nil {
		logger.error(err, "failed to create capnp message")
		return 0
	}

	chunkResult, err := schema.NewRootChunkResult(seg)
	if err != nil {
		logger.error(err, "failed to create chunk result")
		return 0
	}

	chunkList, err := chunkResult.NewChunks(int32(len(chunks)))
	if err != nil {
		logger.error(err, "failed to create chunk list")
		return 0
	}

	for i, chunk := range chunks {
		if err := chunkList.Set(i, chunk); err != nil {
			logger.error(err, "failed to set chunk")
			return 0
		}
	}

	if err := chunkResult.SetChunks(chunkList); err != nil {
		logger.error(err, "failed to set chunk result")
		return 0
	}

	// Serialize the message to a byte slice
	var output bytes.Buffer
	if err := capnp.NewEncoder(&output).Encode(msg); err != nil {
		logger.error(err, "failed to encode capnp message")
		return 0
	}

	// Allocate memory for the message in the module
	ptr, err := alloc.Allocate(ctx, m, uint64(output.Len()))
	if err != nil {
		logger.error(err, "failed to allocate memory for chunk result")
		return 0
	}

	// Write the message to the allocated memory
	if !m.Memory().Write(uint32(ptr), output.Bytes()[:output.Len()]) {
		logger.error(err, "failed to write chunk result to memory")
		return 0
	}

	// Return the pointer + size of the message
	return alloc.EncodePtrWithSize(ptr, uint64(output.Len()))
}
