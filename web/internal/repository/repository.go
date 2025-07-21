package repository

import (
	"sync"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/pgvector/pgvector-go"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/otp"
)

type baseRepo struct {
	queries    *queries.Queries
	store      kv.Store
	pool       *pgxpool.Pool
	otpManager otp.Manager

	// Singletons
	userRepo        UserRepository
	authRepo        AuthRepository
	mfaRepo         MfaRepository
	totpRepo        TOTPRepository
	workspaceRepo   WorkspaceRepository
	collectionRepo  CollectionRepository
	entryRepo       EntryRepository
	pluginRepo      PluginRepository
	pluginStoreRepo PluginStoreRepository

	// Mutex for thread safety
	mu sync.Mutex
}

type PublicIdOrSlug struct {
	PublicID pgtype.UUID
	Slug     string
}

type Repository interface {
	UserRepository() UserRepository
	AuthRepository() AuthRepository
	MfaRepository() MfaRepository
	TOTPRepository() TOTPRepository
	WorkspaceRepository() WorkspaceRepository
	CollectionRepository() CollectionRepository
	EntryRepository() EntryRepository
	PluginRepository() PluginRepository
	PluginStoreRepository() PluginStoreRepository
}

func New(pool *pgxpool.Pool, store kv.Store, otpManager otp.Manager) Repository {
	queries := queries.New(pool)
	return &baseRepo{
		queries:    queries,
		store:      store,
		pool:       pool,
		otpManager: otpManager,
		mu:         sync.Mutex{},
	}
}

func (r *baseRepo) withLock(f func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f()
}

func (r *baseRepo) UserRepository() UserRepository {
	r.withLock(func() {
		if r.userRepo == nil {
			r.userRepo = &userRepo{baseRepo: r}
		}
	})

	return r.userRepo
}

func (r *baseRepo) AuthRepository() AuthRepository {
	r.withLock(func() {
		if r.authRepo == nil {
			r.authRepo = &authRepo{baseRepo: r}
		}
	})

	return r.authRepo
}

func (r *baseRepo) MfaRepository() MfaRepository {
	r.withLock(func() {
		if r.mfaRepo == nil {
			r.mfaRepo = &mfaRepo{baseRepo: r}
		}
	})

	return r.mfaRepo
}

func (r *baseRepo) TOTPRepository() TOTPRepository {
	r.withLock(func() {
		if r.totpRepo == nil {
			r.totpRepo = &totpRepo{baseRepo: r}
		}
	})

	return r.totpRepo
}

func (r *baseRepo) WorkspaceRepository() WorkspaceRepository {
	r.withLock(func() {
		if r.workspaceRepo == nil {
			r.workspaceRepo = &workspaceRepo{baseRepo: r}
		}
	})

	return r.workspaceRepo
}

func (r *baseRepo) CollectionRepository() CollectionRepository {
	r.withLock(func() {
		if r.collectionRepo == nil {
			r.collectionRepo = &collectionRepo{baseRepo: r}
		}
	})

	return r.collectionRepo
}

func (r *baseRepo) EntryRepository() EntryRepository {
	r.withLock(func() {
		if r.entryRepo == nil {
			r.entryRepo = &entryRepo{baseRepo: r}
		}
	})

	return r.entryRepo
}

func (r *baseRepo) PluginRepository() PluginRepository {
	r.withLock(func() {
		if r.pluginRepo == nil {
			r.pluginRepo = &pluginRepo{baseRepo: r}
		}
	})

	return r.pluginRepo
}

func (r *baseRepo) PluginStoreRepository() PluginStoreRepository {
	r.withLock(func() {
		if r.pluginStoreRepo == nil {
			r.pluginStoreRepo = &pluginStoreRepo{baseRepo: r}
		}
	})

	return r.pluginStoreRepo
}

var _ Repository = (*baseRepo)(nil)
