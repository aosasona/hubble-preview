package boundhost

import (
	"errors"
	"fmt"
	"os"

	zerolog "github.com/rs/zerolog"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
)

var (
	ErrFailedToReadMemory     = errors.New("failed to read shared memory from arguments")
	ErrFailedToUnmarshalCapnp = errors.New("failed to unmarshal capnp message")
	ErrFailedToReadCapnp      = errors.New("failed to read capnp message")
)

type hostFnLogger struct {
	logger zerolog.Logger
}

func (b *BoundHost) HostFnLogger(perm spec.Perm) *hostFnLogger {
	logger := zerolog.New(os.Stderr).
		With().
		Str("source", "host_function").
		Str("fn", perm.String()).
		Logger()

	return &hostFnLogger{logger: logger}
}

func (h *hostFnLogger) error(err error, msg ...string) {
	l := h.logger.Error().Stack().Err(err)
	if len(msg) > 0 {
		l.Str("message", msg[0])
	}
	l.Send()
}

func (h *hostFnLogger) errorf(format string, args ...interface{}) {
	err := fmt.Errorf(format, args...)
	h.logger.Error().Stack().Err(err).Send()
}

func (h *hostFnLogger) info(msg string) {
	h.logger.Info().Msg(msg)
}

func (h *hostFnLogger) debug(msg string) {
	h.logger.Debug().Msg(msg)
}

func (h *hostFnLogger) warn(msg string) {
	h.logger.Warn().Msg(msg)
}

func (h *hostFnLogger) warnf(format string, args ...interface{}) {
	h.logger.Warn().Stack().Msg(fmt.Sprintf(format, args...))
}
