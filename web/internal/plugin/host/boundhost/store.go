package boundhost

import (
	"bytes"
	"context"
	"fmt"

	"capnproto.org/go/capnp/v3"
	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/hubble/web/internal/plugin/host/alloc"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	pluginstore "go.trulyao.dev/hubble/web/internal/plugin/store"
	"go.trulyao.dev/hubble/web/schema"
)

type boundStore struct {
	store     pluginstore.Store
	boundHost *BoundHost
}

/*
Set a value for a given key (scoped to a plugin identifier) in the plugin store.

Signature: fn(request: Capnp::StoreSetRequest) -> String (new value)

Exported as `store_set`

NOTE: this function returns the new value in the shared memory
*/
func (s *boundStore) Set(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	logger := s.boundHost.HostFnLogger(spec.PermStoreSet)

	buf, ok := m.Memory().Read(offset, byteCount)
	if !ok {
		logger.error(ErrFailedToReadMemory)
		return 0
	}

	msg, err := capnp.Unmarshal(buf)
	if err != nil {
		logger.error(err, "failed to unmarshal capnp message")
		return 0
	}

	params, err := schema.ReadRootStoreSetRequest(msg)
	if err != nil {
		logger.errorf("failed to read store set request")
		return 0
	}

	key, err := params.Key()
	if err != nil || !params.HasKey() {
		logger.errorf("no key present in store_set request")
		return 0
	}

	value, err := params.Value()
	if err != nil || !params.HasValue() {
		logger.errorf("no value present in store_set request")
		return 0
	}

	newValue, err := s.store.Set(ctx, s.boundHost.pluginIdentifier, key, []byte(value))
	if err != nil {
		logger.errorf("failed to set value in store: %v", err)
		return 0
	}

	encodedPtr, err := alloc.WriteBufferToMemory(ctx, m, newValue)
	if err != nil {
		logger.error(err, "failed to write new value to memory")
		return 0
	}

	return encodedPtr
}

const NotFoundValue = "__NOT_FOUND_0x0000__"

/*
Get retrieves a value for a given key (scoped to a plugin identifier) in the plugin store. An empty string is returned if the key is not found.

Signature: fn(key: String) -> String (value)

Exported as `store_get`

NOTE: this function writes either the value or `__NOT_FOUND_0x0000__` to the shared memory
*/
func (s *boundStore) Get(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	logger := s.boundHost.HostFnLogger(spec.PermStoreGet)

	buf, err := alloc.ReadBufferFromMemory(ctx, m, offset, byteCount)
	if err != nil {
		logger.error(err, "failed to read memory")
		return 0
	}

	key := string(buf)
	value, exists, err := s.store.Get(ctx, s.boundHost.pluginIdentifier, key)
	if err != nil {
		logger.errorf("failed to get value from store: %v", err)
		return 0
	}

	if !exists {
		encodedPtr, err := alloc.WriteBufferToMemory(ctx, m, []byte(NotFoundValue))
		if err != nil {
			logger.error(err, "failed to write not found value to memory")
			return 0
		}
		return encodedPtr
	}

	encodedPtr, err := alloc.WriteBufferToMemory(ctx, m, value)
	if err != nil {
		logger.error(err, "failed to write value to memory")
		return 0
	}

	return encodedPtr
}

/*
Delete removes a value for a given key (scoped to a plugin identifier) in the plugin store.

Signature: fn(key: String) -> String (status) where status is either "OK" or "ERR(<error>)"

Exported as `store_delete`

NOTE: this function writes either an `OK` or `ERR(<error>)` to the shared memory
*/
func (s *boundStore) Delete(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	logger := s.boundHost.HostFnLogger(spec.PermStoreDelete)

	buf, err := alloc.ReadBufferFromMemory(ctx, m, offset, byteCount)
	if err != nil {
		logger.error(err, "failed to read memory")
		return 0
	}

	key := string(buf)
	err = s.store.Delete(ctx, s.boundHost.pluginIdentifier, key)
	if err != nil {
		logger.errorf("failed to delete value from store: %v", err)
		errBuf := fmt.Appendf(nil, "ERR(%s)", err)
		encodedPtr, err := alloc.WriteBufferToMemory(ctx, m, errBuf)
		if err != nil {
			logger.error(err, "failed to write error message to memory")
			return 0
		}
		return encodedPtr
	}

	encodedPtr, err := alloc.WriteBufferToMemory(ctx, m, []byte("OK"))
	if err != nil {
		logger.error(err, "failed to write OK to memory")
		return 0
	}

	return encodedPtr
}

/*
All retrieves all key-value pairs for a given plugin identifier in the plugin store.

Signature: fn() -> List[StoreAllResponse]

Exported as `store_all`
*/
func (s *boundStore) All(ctx context.Context, m api.Module, _, _ uint32) uint64 {
	logger := s.boundHost.HostFnLogger(spec.PermStoreAll)

	pairs, err := s.store.All(ctx, s.boundHost.pluginIdentifier)
	if err != nil {
		logger.errorf("failed to get keys from store: %v", err)
		return 0
	}

	arena := capnp.SingleSegment(nil)
	msg, seg, err := capnp.NewMessage(arena)
	if err != nil {
		logger.error(err, "failed to create capnp message")
		return 0
	}

	keys, err := schema.NewRootStoreAllResponse(seg)
	if err != nil {
		logger.error(err, "failed to create capnp message")
		return 0
	}

	pairsList, err := keys.NewPairs(int32(len(pairs)))
	if err != nil {
		logger.error(err, "failed to create capnp message")
		return 0
	}

	i := 0
	for k, v := range pairs {
		pair, err := schema.NewStoreKvPair(seg)
		if err != nil {
			logger.error(err, "failed to create capnp message for kv pair")
			i++
			continue
		}

		_ = pair.SetKey(k)
		_ = pair.SetValue(string(v))
		_ = pairsList.Set(i, pair)

		i++
	}

	var outputBuffer bytes.Buffer
	if err := capnp.NewEncoder(&outputBuffer).Encode(msg); err != nil {
		logger.error(err, "failed to encode capnp message")
		return 0
	}

	ptr, err := alloc.Allocate(ctx, m, uint64(outputBuffer.Len()))
	if err != nil {
		logger.error(err, "failed to allocate memory for capnp message")
		return 0
	}

	if !m.Memory().Write(uint32(ptr), outputBuffer.Bytes()[:outputBuffer.Len()]) {
		logger.error(err, "failed to write capnp message to memory")
		return 0
	}

	return alloc.EncodePtrWithSize(ptr, uint64(outputBuffer.Len()))
}

/*
Clear clears all values for a given plugin identifier in the plugin store.

Signature: fn() -> Void

Exported as `store_clear`
*/
func (s *boundStore) Clear(ctx context.Context, m api.Module, offset, byteCount uint32) uint64 {
	logger := s.boundHost.HostFnLogger(spec.PermStoreClear)

	err := s.store.Clear(ctx, s.boundHost.pluginIdentifier)
	if err != nil {
		logger.errorf("failed to clear store: %v", err)
		return 0
	}

	return 0
}
