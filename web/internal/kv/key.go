package kv

import (
	"fmt"
	"strings"
)

type (
	//{namespace, collection, record_specific_id}
	//
	// e.g. user:foo:100
	//
	// WARNING: never manually construct a KeyContainer, always use the Key function
	KeyContainer [3]string

	namespace  string
	collection string
)

func NoopKey() KeyContainer {
	return KeyContainer{}
}

// Key creates a new key from a namespace, collection, and meta
//
// The key will be in the format "namespace:collection:meta" (e.g. "user:verification_token:100")
func Key(namespace namespace, col collection, meta string) KeyContainer {
	return KeyContainer{string(namespace), string(col), meta}
}

// UnmarshalKey unmarshals a key from a string
//
// The key should be in the format "namespace:collection:meta" (e.g. "user:verification_token:100") or "namespace:collection" (e.g. "config:foo")
func UnmarshalKey(k string) (KeyContainer, error) {
	fields := strings.Split(k, ":")
	if len(fields) < 2 {
		return KeyContainer{}, ErrInvalidKey
	}

	switch len(fields) {
	case 2:
		return KeyContainer{fields[0], fields[1], ""}, nil
	case 3:
		return KeyContainer{fields[0], fields[1], fields[2]}, nil
	}

	return KeyContainer{}, ErrInvalidKey
}

// String returns the string representation of the key (implements the Stringer interface)
func (k *KeyContainer) String() string {
	return fmt.Sprintf("%s:%s:%s", k[0], k[1], k[2])
}

// Byte returns the byte representation of the key
func (k *KeyContainer) Byte() []byte {
	return fmt.Appendf(nil, "%s:%s:%s", k[0], k[1], k[2])
}

// Namespace returns the namespace of the key
func (k *KeyContainer) Namespace() namespace {
	return namespace(k[0])
}

// Collection returns the collection the key belongs to
func (k *KeyContainer) Collection() collection {
	return collection(k[1])
}

// Meta returns the meta of the key
func (k *KeyContainer) Meta() string {
	return k[2]
}

var _ fmt.Stringer = (*KeyContainer)(nil)
