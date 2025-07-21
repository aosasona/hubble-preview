package host

import (
	"context"
	"errors"
	"os"
	"path"
	"time"

	"github.com/tetratelabs/wazero"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/plugin"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/seer"
)

type ContextType int

const (
	ContextTypeBackground ContextType = iota
	ContextTypeTimeRestricted
)

const DefaultTimeout = time.Second * 30

var ErrNilOptions = errors.New("options cannot be nil")

type Runtime struct {
	config  *config.Config
	runtime wazero.Runtime
	repos   repository.Repository
	host    *Host
	queueFn job.QueueFn
}

type RuntimeOptions struct {
	// ContextType is the type of context to use for the runtime.
	// It can be either ContextTypeBackground or ContextTypeTimeRestricted.
	ContextType ContextType

	// Config is the application configuration.
	Config *config.Config

	// Repos is the repository interface.
	Repos repository.Repository
}

func NewRuntime(options *RuntimeOptions) (*Runtime, error) {
	if options == nil {
		return nil, ErrNilOptions
	}

	if options.Repos == nil {
		return nil, seer.Wrap("nil_repos_in_runtime", errors.New("repos cannot be nil"))
	}

	ctx := context.Background()
	if options.ContextType == ContextTypeTimeRestricted {
		parentCtx := context.Background()
		cancellableCtx, cancel := context.WithTimeout(parentCtx, DefaultTimeout)
		defer cancel()
		ctx = cancellableCtx
	}

	rt := wazero.NewRuntime(ctx)

	return &Runtime{
		runtime: rt,
		config:  options.Config,
		repos:   options.Repos,
		host:    nil,
		queueFn: nil,
	}, nil
}

func (r *Runtime) Host() *Host {
	if r.queueFn == nil {
		panic("queueFn is not set")
	}

	if r.host == nil {
		r.host = NewHost(r.repos, r.queueFn)
	}
	return r.host
}

func (r *Runtime) SetQueueFn(fn job.QueueFn) {
	r.queueFn = fn
}

func (r *Runtime) Close() error {
	return r.runtime.Close(context.Background())
}

// LoadPlugin loads and pre-compiles a plugin for the given identifier.
func (r *Runtime) LoadPlugin(
	ctx context.Context,
	target *models.InstalledPlugin,
) (*Instance, error) {
	instance, err := r.CompilePlugin(ctx, target)
	if err != nil {
		return nil, seer.Wrap("compile_plugin_in_load_plugin", err)
	}

	return instance, nil
}

func (r *Runtime) CompilePlugin(
	ctx context.Context,
	target *models.InstalledPlugin,
) (*Instance, error) {
	var (
		pluginPath = path.Join(
			r.config.Plugins.Directory,
			plugin.DirInstalledPlugins,
			target.PluginIdentifier,
		)

		wasmPath  = path.Join(pluginPath, plugin.OutputWasmFile)
		cachePath = path.Join(pluginPath, plugin.DirCompilationCache)
	)

	// Ensure the plugin exists
	if _, err := os.Stat(wasmPath); err != nil {
		return nil, seer.Wrap("find_plugin_wasm_in_load_plugin", err)
	}

	// Create the cache directory if it doesn't exist
	if err := os.MkdirAll(cachePath, os.ModePerm); err != nil {
		return nil, seer.Wrap("create_plugin_cache_dir", err)
	}

	// Read the binary from the plugin path
	binary, err := os.ReadFile(wasmPath)
	if err != nil {
		return nil, seer.Wrap("read_plugin_wasm_in_load_plugin", err)
	}

	// Create the cache config (so that we can re-use the compilation)
	compilationCache, err := wazero.NewCompilationCacheWithDir(cachePath)
	if err != nil {
		return nil, seer.Wrap("create_plugin_cache_config", err)
	}

	rtConfig := wazero.NewRuntimeConfig().
		WithCloseOnContextDone(true).
		WithCompilationCache(compilationCache)

	rt := wazero.NewRuntimeWithConfig(ctx, rtConfig)

	// Read the binary from the plugin path
	compiled, err := rt.CompileModule(ctx, binary)
	if err != nil {
		compilationCache.Close(ctx) //nolint:errcheck
		rt.Close(ctx)               //nolint:errcheck

		return nil, seer.Wrap("compile_plugin_in_load_plugin", err)
	}

	// cleanupFn fn
	cleanupFn := func() error {
		if err := compilationCache.Close(ctx); err != nil {
			return seer.Wrap("close_plugin_cache", err)
		}

		if err := rt.Close(ctx); err != nil {
			return seer.Wrap("close_plugin_runtime", err)
		}

		return nil
	}

	// Load the host functions
	hostModule := rt.NewHostModuleBuilder("env")
	boundHost := r.Host().Bind(ctx, target.PluginIdentifier, target.Privileges())
	for name, fn := range boundHost.GetFunctions() {
		hostModule.NewFunctionBuilder().WithFunc(fn.Fn).Export(name.String())
	}

	if _, err = hostModule.Instantiate(ctx); err != nil {
		cleanupFn() //nolint:errcheck
		return nil, seer.Wrap("instantiate_plugin_log_function", err)
	}

	moduleConfig := wazero.NewModuleConfig().WithStdout(os.Stdout).WithStderr(os.Stderr)
	module, err := rt.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		cleanupFn() //nolint:errcheck
		return nil, seer.Wrap("instantiate_plugin_in_load_plugin", err)
	}

	return &Instance{
		cleanup: cleanupFn,
		module:  module,
	}, nil
}
