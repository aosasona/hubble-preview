package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/api"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/plugin"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/ograph"
	"go.trulyao.dev/hubble/web/pkg/secrets"
	"go.trulyao.dev/hubble/web/ui"
	"go.trulyao.dev/mirror/v2"
	"go.trulyao.dev/mirror/v2/config"
	"go.trulyao.dev/mirror/v2/generator/typescript"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
	"golang.org/x/sync/errgroup"
)

func ping(ctx *robin.Context, _ robin.Void) (string, error) {
	return "pong", nil
}

func (a *App) buildInstance() (*robin.Instance, error) {
	var err error

	//nolint:exhaustruct
	if a.robin, err = robin.New(robin.Options{
		//nolint:exhaustruct
		CodegenOptions: robin.CodegenOptions{
			Path:             "./ui/src/lib/server",
			GenerateBindings: a.config.InDevelopment(),
			ThrowOnError:     true,
		},
		ErrorHandler:    apperrors.ErrorHandler,
		EnableDebugMode: a.config.Debug(),
	}); err != nil {
		return nil, err
	}

	// Setup the handlers
	a.attachProcedures()

	instance, err := a.robin.Build()
	if err != nil {
		return nil, err
	}

	if err := instance.Export(); err != nil {
		return nil, err
	}

	return instance, nil
}

func (a *App) generateAdditionalTypes() error {
	//nolint:exhaustruct
	m := mirror.New(config.Config{
		Enabled:              a.config.InDevelopment(),
		EnableParserCache:    true,
		FlattenEmbeddedTypes: false,
	})

	//nolint:exhaustruct
	m.AddSources(
		models.User{},
		models.Collection{},
		models.Workspace{},
		models.MemberWithUserID{},
		models.EntryAddedBy{},
		models.EntryRelation{},
		models.Entry{},
		models.PluginSource{},
		ograph.Metadata{},
		models.FileMetadata{},
		models.Member{},
		models.MembershipStatus{},
		models.MemberUser{},
		models.InstalledPlugin{},
		models.SearchResultChunkMetadata{},
		models.SearchResult{},
		models.MatchedChunk{},
		models.CollapsedSearchResult{},
		models.HybridSearchResults{},
		api.PluginListItemSource{},
		api.PluginListItem{},
		spec.Privilege{},
		plugin.PluginV1{},
		plugin.SourceV1{},
		spec.SourceWithPlugins{},
	)

	m.AddTarget(
		typescript.DefaultConfig().
			SetOutputPath("./ui/src/lib/server").
			SetFileName("types.ts").
			SetIncludeSemiColon(true).
			SetPreferUnknown(true).
			SetIndentationType(config.IndentSpace).
			SetIndentationCount(2),
	)

	return m.GenerateAndSaveAll()
}

func (a *App) Run() error {
	// Migration
	if err := a.migrate(); err != nil && err.Error() != migrate.ErrNoChange.Error() {
		return err
	}

	// Build and serve instance
	instance, err := a.buildInstance()
	if err != nil {
		return err
	}

	// Check if it is using `go run ...` in development mode
	isDev := strings.HasPrefix(os.Args[0], os.TempDir()) || strings.Contains(os.Args[0], "tmp")
	mux := http.NewServeMux()

	//nolint:exhaustruct
	corsOpts := &robin.CorsOptions{
		Origins:          []string{"http://localhost:5173"},
		AllowCredentials: true,
		Methods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		Headers:          []string{"Content-Type", "Authorization", "Origin"},
	}

	// Robin handler
	handle := instance.Handler()
	mux.HandleFunc("POST /_", func(w http.ResponseWriter, r *http.Request) {
		if isDev {
			robin.CorsHandler(w, corsOpts)
		}
		handle(w, r)
	})

	// SPA and CORS preflight
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && isDev {
			robin.PreflightHandler(w, corsOpts)

			return
		}

		ui.ServeSPA(w, r)
	})

	// API endpoints
	instance.AttachRestEndpoints(mux, &robin.RestApiOptions{
		Enable:                 true,
		Prefix:                 "/api/v1",
		DisableNotFoundHandler: true,
	})

	// Run the gorutines in parallel and wait for them to finish; collect any errors that may have occurred
	g := new(errgroup.Group)

	// Handle graceful shutdown
	g.Go(func() error { return a.handleInterrupt() })

	// Generate types
	g.Go(func() error { return a.generateAdditionalTypes() })

	// Handle secrets rotation
	g.Go(func() error {
		if err := a.secretsManager.HandleSecretRotation(secrets.ScopeTotp); err != nil {
			log.Error().Err(err).Msg("failed to rotate TOTP secrets")
			return err
		}

		return nil
	})

	// Start the queue
	g.Go(func() error {
		if err := a.queue.Load(); err != nil {
			log.Error().Err(err).Msg("failed to load queue")
			return err
		}

		if err := a.queue.Start(); err != nil {
			log.Fatal().Err(err).Msg("failed to start queue")
		}

		return nil
	})

	// Start the embeddings cron manager
	g.Go(func() error {
		if err := a.llmCron.Start(); err != nil {
			log.Error().Err(err).Msg("failed to start embeddings cron manager")
			return err
		}

		return nil
	})

	// Start the HTTP server
	g.Go(func() error {
		defer func() {
			if err := recover(); err != nil && a.config.InDevelopment() {
				log.Error().Any("error", err).Bool("panic", true).Msg("recovered from panic")
			}
		}()

		// Only collect stack traces in development mode
		seer.SetCollectStackTrace(
			a.config.InDevelopment() || a.config.InStaging() || a.config.Debug(),
		)

		log.Info().Int("port", a.config.Port).Msg("Server started")
		return http.ListenAndServe(fmt.Sprintf(":%d", a.config.Port), mux)
	})

	return g.Wait()
}

// protect is a helper function that wraps a procedure with the WithAuth middleware
func (a *App) protect(procedure robin.Procedure) robin.Procedure {
	procedure.WithMiddleware(a.middleware.WithAuth)
	return procedure
}

// protectAll is a helper function that wraps multiple procedures with the WithAuth middleware
func (a *App) protectAll(procedure ...robin.Procedure) {
	for _, p := range procedure {
		a.protect(p)
	}
}

func query[In, Out any](
	r *robin.Robin,
	name string,
	fn func(*robin.Context, In) (Out, error),
	alias ...string,
) robin.Procedure {
	q := robin.Q(name, fn)

	if len(alias) > 0 && alias[0] != "" {
		q.WithAlias(alias[0])
	}

	r.Add(q)

	return q
}

func mutation[In, Out any](
	r *robin.Robin,
	name string,
	fn func(*robin.Context, In) (Out, error),
	alias ...string,
) robin.Procedure {
	mut := robin.M(name, fn)

	if len(alias) > 0 && alias[0] != "" {
		mut.WithAlias(alias[0])
	}

	r.Add(mut)

	return mut
}
