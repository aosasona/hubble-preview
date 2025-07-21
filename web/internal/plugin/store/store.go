package pluginstore

import (
	"context"
	"errors"

	"go.trulyao.dev/hubble/web/internal/repository"
)

var (
	ErrInvalidIdentifier = errors.New("invalid identifier")
	ErrInvalidKey        = errors.New("invalid key")
)

type Store interface {
	// Set sets a value for a value for a given plugin identifier
	Set(ctx context.Context, identifier string, key string, value []byte) ([]byte, error)

	// Get retrieves a value for a given plugin identifier
	Get(ctx context.Context, identifier string, key string) ([]byte, bool, error)

	// Delete deletes a value for a given plugin identifier
	Delete(ctx context.Context, identifier string, key string) error

	// All retrieves all keys for a given plugin identifier
	All(ctx context.Context, identifier string) (map[string][]byte, error)

	// Clear clears all values for a given plugin identifier
	Clear(ctx context.Context, identifier string) error
}

func NewDBStore(repo repository.Repository) Store {
	return &dbStore{repo: repo}
}
