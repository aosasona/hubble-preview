package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.trulyao.dev/hubble/web/internal/database/queries"
)

type (
	KvLookupArgs struct {
		Identifier string
		Key        string
	}

	KvSetArgs struct {
		Identifier string
		Key        string
		Value      []byte
	}
)

type PluginStoreRepository interface {
	// StoreGet retrieves a value for a given plugin identifier
	Get(ctx context.Context, args KvLookupArgs) (value []byte, exists bool, err error)

	// Set sets a value for a value for a given plugin identifier
	Set(ctx context.Context, args KvSetArgs) ([]byte, error)

	// Delete deletes a value for a given plugin identifier
	Delete(ctx context.Context, args KvLookupArgs) error

	// All retrieves all key-value pairs for a given plugin identifier
	All(ctx context.Context, identifier string) (map[string][]byte, error)

	// Clear clears all key-value pairs for a given plugin identifier
	Clear(ctx context.Context, identifier string) error
}

type pluginStoreRepo struct {
	*baseRepo
}

// All implements PluginStoreRepository.
func (p *pluginStoreRepo) All(ctx context.Context, identifier string) (map[string][]byte, error) {
	values := make(map[string][]byte)

	rows, err := p.queries.PluginKvGetByPluginID(ctx, identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return values, nil
		}
		return nil, err
	}

	for i := range rows {
		row := &rows[i]
		values[row.Key] = []byte(row.Value)
	}

	return values, nil
}

// Clear implements PluginStoreRepository.
func (p *pluginStoreRepo) Clear(ctx context.Context, identifier string) error {
	return p.queries.PluginKvDeleteByPluginID(ctx, identifier)
}

// Delete implements PluginStoreRepository.
func (p *pluginStoreRepo) Delete(ctx context.Context, args KvLookupArgs) error {
	return p.queries.PluginKvDelete(ctx, queries.PluginKvDeleteParams{
		Identifier: args.Identifier,
		Key:        args.Key,
	})
}

// Get implements PluginStoreRepository.
func (p *pluginStoreRepo) Get(
	ctx context.Context,
	args KvLookupArgs,
) ([]byte, bool, error) {
	v, err := p.queries.PluginKvGet(ctx, queries.PluginKvGetParams{
		Key:        args.Key,
		Identifier: args.Identifier,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if strings.TrimSpace(v) == "" {
		return nil, false, nil
	}

	return []byte(v), true, nil
}

// Set implements PluginStoreRepository.
func (p *pluginStoreRepo) Set(ctx context.Context, args KvSetArgs) ([]byte, error) {
	row, err := p.queries.PluginKvSet(ctx, queries.PluginKvSetParams{
		Key:        args.Key,
		Identifier: args.Identifier,
		Value:      string(args.Value),
	})
	if err != nil {
		return nil, err
	}

	return []byte(row.Value), nil
}

var _ PluginStoreRepository = (*pluginStoreRepo)(nil)
