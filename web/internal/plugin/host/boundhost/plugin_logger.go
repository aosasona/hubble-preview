package boundhost

import (
	"context"
	"os"

	zerolog "github.com/rs/zerolog"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
)

type pluginLogger struct {
	boundHost  *BoundHost
	identifier string
	logger     zerolog.Logger
}

func (b *BoundHost) PluginLogger() *pluginLogger {
	if b.logger == nil {
		l := zerolog.New(os.Stderr).
			With().
			Str("source", "plugin").
			Str("identifier", b.pluginIdentifier).
			Logger()

		b.logger = &pluginLogger{
			identifier: b.pluginIdentifier,
			logger:     l,
			boundHost:  b,
		}
	}

	return b.logger
}

/*
warn writes a warning message from the plugin to the host's stderr.

Signature: fn(String) -> Void

Exported as: "log_warn"
*/
func (l *pluginLogger) warn(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	hostLogger := l.boundHost.HostFnLogger(spec.PermLogDebug)
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		hostLogger.error(ErrFailedToReadMemory)
		return 0
	}

	l.logger.Warn().Str("message", string(buf)).Send()
	return 0
}

/*
debug writes a debug message from the plugin to the host's stderr.

Signature: fn(String) -> Void

Exported as: "log_debug"
*/
func (l *pluginLogger) debug(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	hostLogger := l.boundHost.HostFnLogger(spec.PermLogDebug)
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		hostLogger.error(ErrFailedToReadMemory)
		return 0
	}

	l.logger.Debug().Str("message", string(buf)).Send()
	return 0
}

/*
error writes an error message from the plugin to the host's stderr.

Signature: fn(String) -> Void

Exported as: "log_error"
*/
func (l *pluginLogger) error(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	hostLogger := l.boundHost.HostFnLogger(spec.PermLogError)
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		hostLogger.error(ErrFailedToReadMemory)
		return 0
	}

	l.logger.Error().Str("message", string(buf)).Send()
	return 0
}

// NOTE: unexported for now
func (l *pluginLogger) fatal(_ context.Context, m api.Module, offset, byteCount uint32) uint64 {
	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		l.boundHost.HostFnLogger(spec.PermLogError).error(ErrFailedToReadMemory)
		return 0
	}

	l.logger.Fatal().Str("message", string(buf)).Send()
	return 0
}
