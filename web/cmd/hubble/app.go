package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/api"
	"go.trulyao.dev/hubble/web/api/middleware"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/database"
	"go.trulyao.dev/hubble/web/internal/database/migrations"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/mail"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/otp"
	"go.trulyao.dev/hubble/web/internal/plugin"
	"go.trulyao.dev/hubble/web/internal/plugin/host"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/procedure"
	"go.trulyao.dev/hubble/web/internal/queue"
	"go.trulyao.dev/hubble/web/internal/ratelimit"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/llm"
	llmcron "go.trulyao.dev/hubble/web/pkg/llm/cron"
	"go.trulyao.dev/hubble/web/pkg/secrets"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
)

type App struct {
	config         *config.Config
	store          kv.Store
	mailer         mail.Mailer
	pool           *pgxpool.Pool
	repository     repository.Repository
	rateLimiter    ratelimit.RateLimiter
	otpManager     otp.Manager
	secretsManager *secrets.Manager
	pluginManager  spec.Manager
	objectsStore   *objectstore.Store
	queue          *queue.Queue
	wasmRuntime    *host.Runtime
	llm            *llm.LLM
	llmCron        *llmcron.Cron

	robin      *robin.Robin
	handler    api.Handler
	middleware middleware.Middleware
}

func NewApp() *App {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	config := must(config.Load())
	seer.SetCollectStackTrace(config.Debug())

	if config.Debug() {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Any("config", config).Msg("loaded config")
	}

	// TODO: remove config.Debug negation
	pool := must(database.InitializePool(config.PostgresDSN, !config.Debug()))

	var mailer mail.Mailer
	mailer = mail.NewNoopMailer()
	if config.Flags.Email {
		mailer = must(mail.NewDefaultMailer(&mail.DefaultMailerParams{SmtpConfig: &config.Smtp}))
	}

	//nolint:exhaustruct
	app := &App{
		config: &config,
		pool:   pool,
		mailer: mailer,
	}

	if err := app.createStore(); err != nil {
		log.Fatal().Err(err).Msg("failed to create store")
	}

	if err := app.createObjectStore(); err != nil {
		log.Fatal().Err(err).Msg("failed to create object store")
	}

	if err := app.initDeps(); err != nil {
		log.Fatal().Err(err).Msg("failed to initialize dependencies")
	}

	return app
}

func (a *App) initDeps() error {
	// TODO: refactor to set deps at once
	a.otpManager = otp.NewManager(a.store)
	a.repository = repository.New(a.pool, a.store, a.otpManager)
	a.rateLimiter = ratelimit.NewDefaultRateLimiter(ratelimit.Deps{
		Store:  a.store,
		Limits: procedure.RateLimits,
		Scope:  "a",
	})

	a.middleware = middleware.New(&middleware.Deps{
		Repository:  a.repository,
		Config:      a.config,
		RateLimiter: a.rateLimiter,
	})

	a.secretsManager = secrets.NewManager(a.config, a.repository.TOTPRepository())
	pluginManager, err := plugin.NewManagerV1(a.config, a.repository)
	if err != nil {
		return seer.Wrap("create_plugin_manager", err)
	}
	a.pluginManager = pluginManager

	wasmRuntime, err := host.NewRuntime(&host.RuntimeOptions{
		ContextType: host.ContextTypeBackground,
		Config:      a.config,
		Repos:       a.repository,
	})
	if err != nil {
		return seer.Wrap("create_wasm_runtime", err)
	}
	a.wasmRuntime = wasmRuntime

	llm, err := llm.NewService(a.config, a.repository)
	if err != nil {
		return seer.Wrap("create_llm_service", err)
	}
	a.llm = llm

	a.queue = queue.New(a.config, a.repository, a.objectsStore, a.wasmRuntime, a.llm)
	a.wasmRuntime.SetQueueFn(a.queue.Add) // set queue function to wasm runtime

	// Only enable CRON if LLM is enabled
	llmCron, err := llmcron.NewCron(a.config, a.repository, a.queue)
	if err != nil {
		return seer.Wrap("create_llm_cron", err)
	}
	a.llmCron = llmCron

	a.handler = api.New(&api.Deps{
		Repo:          a.repository,
		Config:        a.config,
		Mailer:        a.mailer,
		OtpManager:    a.otpManager,
		ObjectStore:   a.objectsStore,
		PluginManager: a.pluginManager,
		Queue:         a.queue,
		LLM:           a.llm,
	})

	return nil
}

func (a *App) createObjectStore() error {
	minio, err := objectstore.NewMinioStore(objectstore.MinioOptions{
		Endpoint:  a.config.Minio.Endpoint,
		AccessKey: a.config.Minio.AccessKey,
		SecretKey: a.config.Minio.SecretKey,
		IsDev:     a.config.InDevelopment(),
	})
	if err != nil {
		return seer.Wrap("create_object_store", err)
	}

	a.objectsStore = minio
	return nil
}

func (a *App) createStore() error {
	storeConfig := &kv.StoreConfig{
		EtcdStoreConfig: &kv.EtcdStoreConfig{
			Endpoints: a.config.EtcdEndpoints,
		},
		BadgerDbStoreConfig: &kv.BadgerDbStoreConfig{
			Path: a.config.BadgerDbPath,
		},
	}

	store, err := kv.NewStore(kv.Driver(a.config.Drivers.KV), storeConfig)
	if err != nil {
		return seer.Wrap("create_store", err)
	}

	a.store = store
	return nil
}

func (a *App) close() {
	// Plugin manager
	log.Info().Msg("closing plugin manager")
	if err := a.pluginManager.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close plugin manager")
	}

	// WASM runtime
	log.Info().Msg("shutting down WASM runtime")
	if err := a.wasmRuntime.Close(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown wasm runtime")
	}

	// Ratelimit manager
	log.Info().Msg("closing rate limit manager")
	if err := a.rateLimiter.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close rate limit manager")
	}

	// KV store
	log.Info().Msg("closing kv store")
	if err := a.store.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close etcd client")
	}

	// Object store
	log.Info().Msg("closing object store")
	if err := a.objectsStore.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close object store")
	}

	// Queue
	log.Info().Msg("shutting down queue")
	if err := a.queue.Close(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown queue")
	}

	// embeddings cron manager
	log.Info().Msg("shutting down embeddings cron manager")
	if err := a.llmCron.Stop(); err != nil {
		log.Error().Err(err).Msg("failed to shutdown embeddings cron manager")
	}

	log.Info().Msg("closing database connection pool")
	a.pool.Close()
}

func (a *App) migrate() error {
	return migrations.Migrate(a.config.PostgresDSN, migrations.Up)
}

func (a *App) handleInterrupt() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c

	log.Info().Msg("received interrupt signal")

	a.close()

	log.Info().Msg("exiting")
	os.Exit(0)

	return nil
}
