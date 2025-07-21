package pluginstore

import (
	"context"
	"regexp"

	"go.trulyao.dev/hubble/web/internal/repository"
)

type dbStore struct {
	repo repository.Repository
}

// All implements Store.
func (d *dbStore) All(ctx context.Context, identifier string) (map[string][]byte, error) {
	if err := d.validateIdentifier(identifier); err != nil {
		return map[string][]byte{}, err
	}

	return d.repo.PluginStoreRepository().All(ctx, identifier)
}

// Clear implements Store.
func (d *dbStore) Clear(ctx context.Context, identifier string) error {
	if err := d.validateIdentifier(identifier); err != nil {
		return err
	}

	return d.repo.PluginStoreRepository().Clear(ctx, identifier)
}

// Delete implements Store.
func (d *dbStore) Delete(ctx context.Context, identifier string, key string) error {
	if err := d.validate(key, identifier); err != nil {
		return err
	}

	return d.repo.PluginStoreRepository().Delete(ctx, repository.KvLookupArgs{
		Identifier: identifier,
		Key:        key,
	})
}

// Get implements Store.
func (d *dbStore) Get(ctx context.Context, identifier string, key string) ([]byte, bool, error) {
	if err := d.validate(key, identifier); err != nil {
		return nil, false, err
	}

	return d.repo.PluginStoreRepository().Get(ctx, repository.KvLookupArgs{
		Identifier: identifier,
		Key:        key,
	})
}

// Set implements Store.
func (d *dbStore) Set(
	ctx context.Context,
	identifier string,
	key string,
	value []byte,
) ([]byte, error) {
	if err := d.validate(key, identifier); err != nil {
		return nil, err
	}

	newValue, err := d.repo.PluginStoreRepository().Set(ctx, repository.KvSetArgs{
		Identifier: identifier,
		Key:        key,
		Value:      value,
	})
	return newValue, err
}

var (
	keyRegex        = regexp.MustCompile(`^[a-z0-9_\.]+$`)
	identifierRegex = regexp.MustCompile(`^[a-f0-9]{32}$`)
)

func (d *dbStore) validateKey(key string) error {
	if !keyRegex.MatchString(key) {
		return ErrInvalidKey
	}
	return nil
}

func (d *dbStore) validateIdentifier(identifier string) error {
	if !identifierRegex.MatchString(identifier) {
		return ErrInvalidIdentifier
	}
	return nil
}

func (d *dbStore) validate(key, identifier string) error {
	if err := d.validateKey(key); err != nil {
		return err
	}
	if err := d.validateIdentifier(identifier); err != nil {
		return err
	}
	return nil
}

var _ Store = (*dbStore)(nil)
