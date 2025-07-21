package boundhost

import (
	"context"
	"crypto/rand"
	"errors"

	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
)

// rand generates a cryptographically secure random text with 128 bits of randomness.
func (b *BoundHost) rand(
	ctx context.Context,
	m api.Module,
	desiredLen, _ uint32, // WARNING: we are using offset as the size of the buffer
) uint64 {
	logger := b.HostFnLogger(spec.PermCryptoRand)

	if desiredLen == 0 {
		logger.error(errors.New("desired length is zero"))
		return 0
	}

	if desiredLen > 1024 {
		logger.error(errors.New("desired length is too large"))
		return 0
	}

	randBytes := make([]byte, desiredLen)
	if _, err := rand.Read(randBytes); err != nil {
		logger.error(err, "failed to read random bytes")
		return 0
	}

	encoded, err := alloc.WriteBufferToMemory(ctx, m, randBytes)
	if err != nil {
		logger.error(err, "failed to write random bytes to memory")
		return 0
	}

	return encoded
}
