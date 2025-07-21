package kv

import (
	"context"
	"time"

	json "github.com/goccy/go-json"
	etcd "go.etcd.io/etcd/client/v3"
	"go.trulyao.dev/seer"
)

// An ETCD adapter for the kV store
type EtcdStore struct {
	client *etcd.Client
}

type EtcdStoreConfig struct {
	Endpoints []string
}

func NewEtcdStore(config *EtcdStoreConfig) (*EtcdStore, error) {
	if len(config.Endpoints) == 0 {
		return nil, ErrEtcdEndpointsNotConfigured
	}

	etcdConfig := etcd.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := etcd.New(etcdConfig)
	if err != nil {
		return nil, seer.Wrap("make_etcd_client", err)
	}

	return &EtcdStore{client}, nil
}

// Close the store and/or its underlying client.
func (e *EtcdStore) Close() error {
	return e.client.Close()
}

// Exists checks if a key exists.
func (e *EtcdStore) Exists(key KeyContainer) (bool, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()
	resp, err := e.client.Get(ctx, key.String(), etcd.WithCountOnly())
	if err != nil {
		return false, err
	}

	return resp.Count > 0, nil
}

// Get gets the value of a key.
func (e *EtcdStore) Get(key KeyContainer) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	return e.GetContext(ctx, key)
}

// GetCtx gets the value of a key with a context.
func (e *EtcdStore) GetContext(ctx context.Context, key KeyContainer) ([]byte, error) {
	resp, err := e.client.Get(ctx, key.String())
	if err != nil {
		return []byte{}, err
	}

	if len(resp.Kvs) == 0 {
		return []byte{}, ErrKeyNotFound
	}

	return resp.Kvs[len(resp.Kvs)-1].Value, nil
}

// Set sets the value of a key.
func (e *EtcdStore) Set(key KeyContainer, value []byte) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	return e.SetContext(ctx, key, value)
}

// SetContext sets the value of a key with a context.
func (e *EtcdStore) SetContext(ctx context.Context, key KeyContainer, value []byte) error {
	_, err := e.client.Put(ctx, key.String(), string(value))
	return err
}

// SetWithTTL sets the value of a key with a time-to-live.
func (e *EtcdStore) SetWithTTL(key KeyContainer, value []byte, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 5*time.Second)
	defer cancel()

	lease, err := e.client.Grant(ctx, int64(ttl.Seconds()))
	if err != nil {
		return err
	}

	_, err = e.client.Put(ctx, key.String(), string(value), etcd.WithLease(lease.ID))
	return err
}

// GetJson gets the value of a key and unmarshals it into a JSON object.
func (e *EtcdStore) GetJson(key KeyContainer, target any) error {
	value, err := e.Get(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(value, target); err != nil {
		return seer.Wrap("unmarshal_json_etcd", err)
	}

	return nil
}

// SetJson sets the value of a key with a JSON value.
func (e *EtcdStore) SetJson(key KeyContainer, value any) error {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := e.Set(key, marshalledValue); err != nil {
		return seer.Wrap("set_json_etcd", err)
	}

	return nil
}

func (e *EtcdStore) SetJsonWithTTL(key KeyContainer, value any, ttl time.Duration) error {
	marshalledValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if err := e.SetWithTTL(key, marshalledValue, ttl); err != nil {
		return seer.Wrap("set_json_with_ttl_etcd", err)
	}

	return nil
}

// Delete deletes a key.
func (e *EtcdStore) Delete(key KeyContainer) error {
	_, err := e.client.Delete(context.TODO(), key.String())
	return err
}

/*
ExpiresAt gets the expiry of a key i.e. the time the lease will expire.

If the key does not exist, it should return ErrKeyNotFound.
A valid time should be returned if the key has an expiry time and still exists.
*/
func (e *EtcdStore) ExpiresAt(key KeyContainer) (time.Time, error) {
	resp, err := e.client.Get(context.TODO(), key.String())
	if err != nil {
		return time.Time{}, err
	}

	if len(resp.Kvs) == 0 {
		return time.Time{}, ErrKeyNotFound
	}

	leaseID := resp.Kvs[len(resp.Kvs)-1].Lease

	// Get the lease info
	lease, err := e.client.Lease.TimeToLive(context.TODO(), etcd.LeaseID(leaseID))
	if err != nil {
		return time.Time{}, err
	}

	// If the original lease has no expiry
	if lease.GrantedTTL == -1 {
		return time.Time{}, ErrNoExpiry
	}

	// If the lease has expired, return a time in the past
	if lease.TTL == -1 {
		return time.Now().Add(-time.Second), nil
	}

	return time.Now().Add(time.Duration(lease.TTL) * time.Second), nil
}

var _ Store = (*EtcdStore)(nil)
