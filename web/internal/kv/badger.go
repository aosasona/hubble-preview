package kv

import (
	"context"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/goccy/go-json"
	"go.trulyao.dev/seer"
)

type (
	BadgerDbStoreConfig struct {
		Path string
	}

	// A badger adapter for the KV store
	BadgerDbStore struct {
		db *badger.DB
	}
)

func NewBadgerDbStore(config *BadgerDbStoreConfig) (*BadgerDbStore, error) {
	if config.Path == "" {
		return nil, seer.Wrap("new_badger_db", ErrInvalidBadgerDbPath)
	}

	db, err := badger.Open(badger.DefaultOptions(config.Path))
	if err != nil {
		return nil, seer.Wrap("open_badger_db", err)
	}

	return &BadgerDbStore{db}, nil
}

// Close the store and/or its underlying client.
func (b *BadgerDbStore) Close() error {
	return b.db.Close()
}

// Get gets the value of a key.
func (b *BadgerDbStore) Get(key KeyContainer) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	return b.GetContext(ctx, key)
}

// GetCtx gets the value of a key with a context.
func (b *BadgerDbStore) GetContext(ctx context.Context, key KeyContainer) ([]byte, error) {
	value := []byte{}

	err := b.db.View(func(tx *badger.Txn) error {
		val, err := tx.Get(key.Byte())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrKeyNotFound
			}

			return err
		}

		value, err = val.ValueCopy(nil)
		if err != nil {
			return err
		}

		return nil
	})

	return value, err
}

// Set sets the value of a key.
func (b *BadgerDbStore) Set(key KeyContainer, value []byte) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	return b.SetContext(ctx, key, value)
}

// SetCtx sets the value of a key with a context.
func (b *BadgerDbStore) SetContext(ctx context.Context, key KeyContainer, value []byte) error {
	return b.db.Update(func(tx *badger.Txn) error {
		return tx.Set(key.Byte(), value)
	})
}

// SetWithTTL sets the value of a key with a time-to-live.
func (b *BadgerDbStore) SetWithTTL(key KeyContainer, value []byte, ttl time.Duration) error {
	return b.db.Update(func(tx *badger.Txn) error {
		entry := badger.NewEntry(key.Byte(), value).WithTTL(ttl)
		return tx.SetEntry(entry)
	})
}

// GetJson gets the value of a key and unmarshals it into the provided interface.
func (b *BadgerDbStore) GetJson(key KeyContainer, target any) error {
	value, err := b.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(value, target); err != nil {
		return seer.Wrap("unmarshal_json_badger", err)
	}

	return nil
}

// SetJson sets the value of a key with a JSON value.
func (b *BadgerDbStore) SetJson(key KeyContainer, value any) error {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := b.Set(key, marshalledValue); err != nil {
		return seer.Wrap("set_json_badger", err)
	}

	return nil
}

// SetJsonWithTTL sets the value of a key with a JSON value and a time-to-live.
func (b *BadgerDbStore) SetJsonWithTTL(key KeyContainer, value any, ttl time.Duration) error {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := b.SetWithTTL(key, marshalledValue, ttl); err != nil {
		return seer.Wrap("set_json_with_ttl_badger", err)
	}

	return nil
}

// Delete removes a value from the store.
func (b *BadgerDbStore) Delete(key KeyContainer) error {
	return b.db.Update(func(tx *badger.Txn) error {
		return tx.Delete(key.Byte())
	})
}

// Exists checks if a key exists.
func (b *BadgerDbStore) Exists(key KeyContainer) (bool, error) {
	var exists bool

	err := b.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key.Byte())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
		}

		exists = item != nil
		return nil
	})

	return exists, err
}

/*
ExpiresAt gets the expiry of a key i.e. the time the lease will expire.

If the key does not exist, it should return ErrKeyNotFound.
A valid time should be returned if the key has an expiry time and still exists.
*/
func (b *BadgerDbStore) ExpiresAt(key KeyContainer) (time.Time, error) {
	var expiresAt time.Time

	err := b.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get(key.Byte())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrKeyNotFound
			}

			return err
		}

		if item.IsDeletedOrExpired() {
			return ErrNoExpiry
		}

		expiresAt = time.Unix(int64(item.ExpiresAt()), 0)

		return nil
	})

	return expiresAt, err
}

var _ Store = (*BadgerDbStore)(nil)
