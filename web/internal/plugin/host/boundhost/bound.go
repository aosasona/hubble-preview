// A bound host is one that has been bound to a particular plugin via its identifier
package boundhost

import (
	"context"

	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	pluginstore "go.trulyao.dev/hubble/web/internal/plugin/store"
	"go.trulyao.dev/hubble/web/internal/repository"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(log_debug,log_error,log_warn,transform_url_to_markdown,transform_html_to_markdown,chunk_with_overlap,chunk_by_sentence,network_request,entry_update,entry_create_chunks,store_set,store_get,store_all,store_delete,store_clear,crypto_rand)
type Fn string

type (
	PrivilegedFn struct {
		Identifier spec.Perm
		Fn         HostFn
	}

	HostFn func(ctx context.Context, m api.Module, ptr, size uint32) uint64
)

type BoundHost struct {
	repository       repository.Repository
	context          context.Context
	pluginIdentifier string
	privileges       spec.Privileges
	store            pluginstore.Store
	boundStore       *boundStore
	logger           *pluginLogger
	queueFn          job.QueueFn
}

func New(
	context context.Context,
	repository repository.Repository,
	pluginIdentifier string,
	privileges spec.Privileges,
	queueFn job.QueueFn,
) *BoundHost {
	b := BoundHost{
		repository:       repository,
		context:          context,
		pluginIdentifier: pluginIdentifier,
		privileges:       privileges,
		store:            pluginstore.NewDBStore(repository),
		logger:           nil,
		queueFn:          queueFn,
	}

	return &b
}

func (b *BoundHost) BoundStore() *boundStore {
	if b.boundStore == nil {
		b.boundStore = &boundStore{
			store:     b.store,
			boundHost: b,
		}
	}

	return b.boundStore
}

func (b *BoundHost) GetFunctions() map[Fn]PrivilegedFn {
	functions := map[Fn]PrivilegedFn{}

	// Logger
	functions[FnLogDebug] = PrivilegedFn{Identifier: spec.PermLogDebug, Fn: b.PluginLogger().debug}
	functions[FnLogError] = PrivilegedFn{Identifier: spec.PermLogError, Fn: b.PluginLogger().error}
	functions[FnLogWarn] = PrivilegedFn{Identifier: spec.PermLogWarn, Fn: b.PluginLogger().warn}

	// Network
	functions[FnNetworkRequest] = PrivilegedFn{
		Identifier: spec.PermNetworkRequest,
		Fn:         b.makeRequest,
	}

	// Transform
	functions[FnChunkWithOverlap] = PrivilegedFn{
		Identifier: spec.PermTransformChunkWithOverlap,
		Fn:         b.chunkWithOverlap,
	}
	functions[FnChunkBySentence] = PrivilegedFn{
		Identifier: spec.PermTransformChunkBySentence,
		Fn:         b.chunkBySentence,
	}

	functions[FnTransformUrlToMarkdown] = PrivilegedFn{
		Identifier: spec.PermTransformUrlToMarkdown,
		Fn:         b.transformUrlToMarkdown,
	}
	functions[FnTransformHtmlToMarkdown] = PrivilegedFn{
		Identifier: spec.PermTransformHtmlToMarkdown,
		Fn:         b.transformHtmlToMarkdown,
	}

	// Entries
	functions[FnEntryUpdate] = PrivilegedFn{
		Identifier: spec.PermEntriesUpdate,
		Fn:         b.updateEntry,
	}
	functions[FnEntryCreateChunks] = PrivilegedFn{
		Identifier: spec.PermChunksCreate,
		Fn:         b.createEntryChunks,
	}

	// Store
	store := b.BoundStore()
	functions[FnStoreSet] = PrivilegedFn{Identifier: spec.PermStoreSet, Fn: store.Set}
	functions[FnStoreGet] = PrivilegedFn{Identifier: spec.PermStoreGet, Fn: store.Get}
	functions[FnStoreAll] = PrivilegedFn{Identifier: spec.PermStoreAll, Fn: store.All}
	functions[FnStoreDelete] = PrivilegedFn{Identifier: spec.PermStoreDelete, Fn: store.Delete}
	functions[FnStoreClear] = PrivilegedFn{Identifier: spec.PermStoreClear, Fn: store.Clear}

	// Crypto
	functions[FnCryptoRand] = PrivilegedFn{Identifier: spec.PermCryptoRand, Fn: b.rand}

	// Filter out functions that are not in the privileges list
	for name, fn := range functions {
		if !b.privileges.Has(fn.Identifier) {
			functions[name] = PrivilegedFn{
				Identifier: spec.PermNoop,
				Fn:         b.permissionDenined(fn.Identifier),
			}
		}
	}

	return functions
}

func (b *BoundHost) permissionDenined(perm spec.Perm) HostFn {
	return func(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
		logger := b.HostFnLogger(perm)
		logger.errorf("attempt to call a privileged function without permission")
		return 0
	}
}

func (b *BoundHost) dealloc(
	ctx context.Context,
	m api.Module,
	logger *hostFnLogger,
	offset, byteCount uint32,
) {
	if err := alloc.Deallocate(ctx, m, uint64(offset), uint64(byteCount)); err != nil {
		logger.error(err, "failed to deallocate memory")
	}
}
