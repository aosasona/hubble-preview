package alloc

import (
	"context"
	"errors"

	"github.com/tetratelabs/wazero/api"
	"go.trulyao.dev/seer"
)

// Allocate allocates memory of the given size in the module's memory space.
func Allocate(ctx context.Context, module api.Module, size uint64) (ptr uint64, err error) {
	allocateFn := module.ExportedFunction("allocate")
	if allocateFn == nil {
		return 0, seer.New("get_allocate_exported_fn", "allocate function not found in module")
	}

	// Call the allocate function
	result, err := allocateFn.Call(ctx, size)
	if err != nil {
		return 0, seer.Wrap("call_allocate_fn", err)
	}

	if len(result) != 1 {
		return 0, seer.New("call_allocate_fn", "unexpected number of results")
	}

	// Get the pointer from the result
	ptr = result[0]
	return ptr, nil
}

// Deallocate frees the memory at the given pointer and size in the module's memory space.
func Deallocate(ctx context.Context, module api.Module, ptr, size uint64) error {
	deallocateFn := module.ExportedFunction("deallocate")
	if deallocateFn == nil {
		return seer.New("get_deallocate_exported_fn", "deallocate function not found in module")
	}

	// Call the deallocate function
	_, err := deallocateFn.Call(ctx, ptr, size)
	if err != nil {
		return seer.Wrap("call_deallocate_fn", err)
	}

	return nil
}

// EncodePtrWithSize encodes a pointer and size into a single value.
// This is useful for returning both a pointer and size from a function as WASM functions can only return one value.
func EncodePtrWithSize(ptr, size uint64) uint64 {
	// Encode the pointer and size into a single value
	return (ptr << 32) | size
}

// DecodePtrWithSize decodes a pointer and size from an encoded value.
func DecodePtrWithSize(encoded uint64) (ptr, size uint64) {
	// Decode the pointer and size from the encoded value
	ptr = encoded >> 32
	size = encoded & 0xFFFFFFFF
	return ptr, size
}

// WriteBufferToMemory writes a byte slice to the module's memory and returns the encoded pointer and size in a single value.
func WriteBufferToMemory(ctx context.Context, m api.Module, buf []byte) (uint64, error) {
	size := uint64(len(buf))
	ptr, err := Allocate(ctx, m, size)
	if err != nil {
		return 0, errors.New("failed to allocate memory for buffer")
	}

	if !m.Memory().Write(uint32(ptr), buf) {
		return 0, errors.New("failed to write buffer to memory")
	}

	return EncodePtrWithSize(ptr, size), nil
}

func ReadBufferFromMemory(
	ctx context.Context,
	module api.Module,
	ptr, size uint32,
) ([]byte, error) {
	buf, ok := module.Memory().Read(ptr, size)
	if !ok {
		return nil, errors.New("failed to read memory")
	}

	return buf, nil
}
