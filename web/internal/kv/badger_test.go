package kv_test

import (
	"testing"
	"time"

	"go.trulyao.dev/hubble/web/internal/kv"
)

func mockStore(t *testing.T) kv.Store {
	store, err := kv.NewBadgerDbStore(&kv.BadgerDbStoreConfig{
		Path: "/tmp/kv.db",
	})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	return store
}

func Test_Set(t *testing.T) {
	t.Run("set key", func(t *testing.T) {
		store := mockStore(t)
		defer store.Close()

		err := store.Set(kv.KeyEmailVerificationToken(1), []byte("token"))
		if err != nil {
			t.Fatalf("failed to set key: %v", err)
		}

		value, err := store.Get(kv.KeyEmailVerificationToken(1))
		if err != nil {
			t.Fatalf("failed to get key: %v", err)
		}

		if string(value) != "token" {
			t.Fatalf("expected value to be 'token', got %s", value)
		}
	})
}

func Test_ExpiredAt(t *testing.T) {
	t.Run("set key with expiry (should be non-existent)", func(t *testing.T) {
		store := mockStore(t)
		defer store.Close()

		key := kv.KeyEmailVerificationToken(34)

		err := store.SetWithTTL(key, []byte("token"), time.Second*2)
		if err != nil {
			t.Fatalf("failed to set key with expiry: %v", err)
		}

		expiresAt, err := store.ExpiresAt(key)
		if err != nil {
			t.Fatalf("failed to get expiry: %v", err)
		}

		if expiresAt.IsZero() {
			t.Fatalf("expected expiry time to be non-zero")
		}

		time.Sleep(time.Second * 2)

		_, err = store.ExpiresAt(key)
		if err == nil {
			t.Fatalf("expected key to be expired")
		}
	})

	t.Run("set key with expiry (should have a valid ttl left)", func(t *testing.T) {
		store := mockStore(t)
		defer store.Close()

		key := kv.KeyEmailVerificationToken(34)
		// Ensure the guy has the expected TTL
		if err := store.SetWithTTL(key, []byte("token"), time.Second*10); err != nil {
			t.Fatalf("failed to set key with expiry: %v", err)
		}

		expiresAt, err := store.ExpiresAt(key)
		if err != nil {
			t.Fatalf("failed to get expiry: %v", err)
		}

		if expiresAt.IsZero() {
			t.Fatalf("expected expiry time to be non-zero")
		}

		time.Sleep(time.Second * 5)

		expiresAt, err = store.ExpiresAt(key)
		if err != nil {
			t.Fatalf("failed to get expiry: %v", err)
		}

		if expiresAt.IsZero() {
			t.Fatalf("expected expiry time to be non-zero")
		}

		if time.Until(expiresAt) < time.Second*3 {
			t.Fatalf(
				"expected expiry time to be at least 15 seconds, got %s",
				time.Until(expiresAt),
			)
		}
	})
}

func Test_JSONGetSet(t *testing.T) {
	type Foo struct {
		Bar string `json:"bar"`
	}

	t.Run("set and retrieve json value", func(t *testing.T) {
		store := mockStore(t)
		defer store.Close()

		foo := Foo{Bar: "baz"}

		err := store.SetJson(kv.KeyEmailVerificationToken(1), foo)
		if err != nil {
			t.Fatalf("failed to set json value: %s", err.Error())
		}

		retrieved := new(Foo)
		err = store.GetJson(kv.KeyEmailVerificationToken(1), retrieved)
		if err != nil {
			t.Fatalf("failed to get json value: %s", err.Error())
		}

		if retrieved.Bar != "baz" {
			t.Fatalf("expected bar to be 'baz', got %s", retrieved.Bar)
		}
	})
}
