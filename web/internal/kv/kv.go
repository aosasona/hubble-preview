package kv

import (
	"context"
	"errors"
	"time"
)

type (
	// Store is the interface for a key-value store.
	Store interface {
		// Get gets the value of a key.
		// This method returns ErrKeyNotFound if the key does not exist.
		Get(KeyContainer) ([]byte, error)
		// GetCtx gets the value of a key with a context.
		// This method returns ErrKeyNotFound if the key does not exist.
		GetContext(context.Context, KeyContainer) ([]byte, error)

		// Set sets the value of a key.
		Set(KeyContainer, []byte) error
		// SetWithTTL sets the value of a key with a time-to-live.
		SetWithTTL(KeyContainer, []byte, time.Duration) error
		// SetCtx sets the value of a key with a context.
		SetContext(context.Context, KeyContainer, []byte) error

		// GetJson gets the value of a key and unmarshals it into the provided interface.
		// This method returns ErrKeyNotFound if the key does not exist.
		GetJson(key KeyContainer, target any) error
		// SetJson sets the value of a key with a JSON value.
		SetJson(key KeyContainer, value any) error
		// SetJsonWithTTL sets the value of a key with a JSON value and a time-to-live.
		SetJsonWithTTL(key KeyContainer, value any, ttl time.Duration) error

		// Exists checks if a key exists.
		Exists(KeyContainer) (bool, error)

		// Delete deletes a key.
		Delete(KeyContainer) error

		// ExpiresAt gets the expiry of a key (i.e. the time left on the lease).
		// If the key does not exist, it should return ErrKeyNotFound.
		// If the key does not have an expiry, it should return ErrNoExpiry.
		// If there is an error, it should return the error.
		// Otherwise, it should return the expiry time.
		ExpiresAt(KeyContainer) (expiresAt time.Time, err error)

		// Close closes the store and/or its underlying client.
		Close() error
	}
)

type Driver string

const (
	DriverEtcd     Driver = "etcd"
	DriverBadgerDb Driver = "badgerdb"
)

var (
	ErrInvalidKey  = errors.New("not a valid Key")
	ErrKeyNotFound = errors.New("key not found")
	ErrNoExpiry    = errors.New("key does not have an expiry")
)

var (
	ErrEtcdEndpointsNotConfigured = errors.New(
		"etcd has been selected as the KV store driver, you need to set the ETCD endpoints",
	)
	ErrInvalidKvDriver     = errors.New("invalid kv driver")
	ErrInvalidBadgerDbPath = errors.New("invalid badger DB path")
)

type StoreConfig struct {
	EtcdStoreConfig     *EtcdStoreConfig
	BadgerDbStoreConfig *BadgerDbStoreConfig
}

func NewStore(driver Driver, config *StoreConfig) (Store, error) {
	switch driver {
	case DriverEtcd:
		return NewEtcdStore(config.EtcdStoreConfig)

	case DriverBadgerDb:
		return NewBadgerDbStore(config.BadgerDbStoreConfig)

	default:
		return nil, ErrInvalidKvDriver
	}
}
