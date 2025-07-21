package plugin

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"slices"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/document"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
)

const (
	DirTmp    = "tmp"
	DirOutput = "output"
	// Plugins are installed in "<plugin_dir>/installed/<identifier>"
	DirInstalledPlugins = "installed"
	DirCompilationCache = "cache"
)

const (
	PluginConfigFileName = "plugin.toml"
	OutputWasmFile       = "plugin.wasm"
	OutputSHA256File     = "plugin.wasm.sha256"
)

var CoreSources = []string{
	"https://github.com/keystroke-tools/hub.git",
}

// NOTE: New sources are cloned into `<plugin_dir>/tmp/<source_name>`
// When they are processed by the plugin manager, the plugin themselves end up in `<plugin_dir>/<workspace_id>/<source_name>`

var (
	ErrInvalidConfig     = errors.New("invalid config argument")
	ErrInvalidRepository = errors.New("invalid repository argument")
	ErrNotADirectory     = errors.New("not a directory")
)

func NewManagerV1(config *config.Config, repository repository.Repository) (spec.Manager, error) {
	if config == nil {
		return nil, ErrInvalidConfig
	}

	if repository == nil {
		return nil, ErrInvalidRepository
	}

	stat, err := os.Stat(config.Plugins.Directory)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.MkdirAll(config.Plugins.Directory, os.ModePerm); err != nil {
				return nil, fmt.Errorf("failed to create plugins directory: %w", err)
			}
		}

		return nil, fmt.Errorf("failed to stat plugins directory: %w", err)
	}

	if !stat.IsDir() {
		return nil, ErrNotADirectory
	}

	handle, err := os.Open(config.Plugins.Directory)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugins directory: %w", err)
	}

	return &ManagerV1{
		config:     config,
		repository: repository,
		handle:     handle,
	}, nil
}

type ManagerV1 struct {
	config     *config.Config
	repository repository.Repository
	// A file handle to the plugins directory (for reading and writing its subdirectories and files)
	handle *os.File
}

// InstallCorePlugins implements spec.Manager.
func (m *ManagerV1) InstallCorePlugins(workspaceID int32) error {
	for _, sourceStr := range CoreSources {
		remote, err := spec.ParseRemoteSource(sourceStr)
		if err != nil {
			return seer.Wrap("parse_remote_source", err)
		}

		// Check if the source already exists
		exists, err := m.repository.PluginRepository().SourceExists(&repository.FindSourceByURLArgs{
			WorkspaceID: workspaceID,
			Source:      &remote,
		})
		if err != nil {
			return seer.Wrap("source_exists", err)
		}

		if exists {
			continue
		}

		// Create source
		createdSource, err := m.AddRemoteSource(spec.AddRemoteSourceArgs{
			WorkspaceID: workspaceID,
			Remote:      &remote,
		})
		if err != nil {
			log.Error().Str("source", sourceStr).Err(err).Msg("failed to create core source")
			return seer.Wrap("create_core_source", err)
		}

		// Install all plugins from the core source
		for _, plugin := range createdSource.Plugins {
			err = m.InstallPlugin(&spec.InstallPluginArgs{
				WorkspaceID: workspaceID,
				PluginName:  plugin.Name(),
				Source:      &remote,
				PullLatest:  false,
			})
			if err != nil {
				log.Error().
					Str("plugin", plugin.Name()).
					Err(err).
					Msg("failed to install core plugin")
				return seer.Wrap("install_core_plugin", err)
			}
		}
	}

	return nil
}

// BelongsToCore implements spec.Manager.
func (m *ManagerV1) BelongsToCore(source *spec.RemoteSource) bool {
	return slices.Contains(CoreSources, source.RawURL())
}

// RemovePlugin implements spec.Manager.
func (m *ManagerV1) RemovePlugin(args *spec.RemovePluginArgs) error {
	// Remove the plugin from the database
	err := m.repository.PluginRepository().RemoveInstalledPlugin(&repository.RemovePluginArgs{
		WorkspaceID: args.WorkspaceID,
		Identifier:  args.Identifier,
	})
	if err != nil {
		return seer.Wrap("remove_plugin", err)
	}

	// Remove the plugin directory
	pluginDir := m.installedPluginDir(args.Identifier)
	if err := os.RemoveAll(pluginDir); err != nil {
		return seer.Wrap("remove_plugin_dir", err)
	}

	return nil
}

func (m *ManagerV1) installedPluginDir(identifier string) string {
	return path.Join(m.config.Plugins.Directory, DirInstalledPlugins, identifier)
}

// InstallPlugin implements spec.Manager.
func (m *ManagerV1) InstallPlugin(args *spec.InstallPluginArgs) error {
	if args == nil {
		return errors.New("args cannot be nil in ManagerV1.InstallPlugin")
	}

	// Ensure the source exists in this workspace
	source, err := m.repository.PluginRepository().FindSourceByURL(&repository.FindSourceByURLArgs{
		WorkspaceID: args.WorkspaceID,
		Source:      args.Source,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.BadRequest("source has not been added to this workspace")
		}
		return seer.Wrap("source_exists_install_plugin", err)
	}

	loadPlugins := m.GetSourcePluginsFromCache
	if args.PullLatest {
		loadPlugins = m.GetSourcePlugins
	}

	plugins, err := loadPlugins(&source.SourceURL)
	if err != nil {
		return seer.Wrap("get_source_plugins", err)
	}

	var plugin spec.Plugin
	for _, p := range plugins {
		if p.Name() == args.PluginName {
			plugin = p
			break
		}
	}

	if plugin == nil {
		return seer.New("install_plugin", "plugin not available in source")
	}

	stat, err := os.Stat(plugin.CachePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return seer.New(
				"install_plugin",
				"plugin not found, ensure source is valid and up to date",
			)
		}
		return seer.Wrap("stat_plugin_path", err)
	}

	// Copy over the plugin to the workspace directory
	identifier := m.GeneratePluginIdentifier(&spec.GeneratePluginIdentifierArgs{
		WorkspaceID: source.WorkspaceID,
		PluginName:  plugin.Name(),
		SourceURL:   args.Source.RawURL(),
	})

	// Copy the plugin to the installed directory
	if err := m.copyPluginToInstalledDir(&CopyPluginArgs{
		WorkspaceID: args.WorkspaceID,
		CachePath:   plugin.CachePath(),
		Identifier:  identifier,
		SourceURL:   args.Source,
	}); err != nil {
		return err
	}

	targets := make([]document.EntryType, 0, len(plugin.Targets()))
	targets = append(targets, plugin.Targets()...)

	privileges := make([]queries.PluginPrivilege, 0, len(plugin.Privileges()))
	for _, privilege := range plugin.Privileges() {
		privileges = append(privileges, queries.PluginPrivilege{
			Identifier:  privilege.Identifier.String(),
			Description: privilege.Description,
		})
	}

	// Add the plugin to the database
	_, err = m.repository.PluginRepository().UpsertInstalledPlugin(&repository.UpsertPluginArgs{
		Identifier:          identifier,
		WorkspaceID:         source.WorkspaceID,
		SourceID:            source.ID,
		Name:                plugin.Name(),
		Description:         plugin.Description(),
		Modes:               plugin.Modes(),
		Targets:             targets,
		Checksum:            plugin.Checksum(),
		PluginLastUpdatedAt: stat.ModTime(),
		Privileges:          privileges,
	})
	if err != nil {
		return seer.Wrap("upsert_plugin", err)
	}

	return nil
}

type CopyPluginArgs struct {
	WorkspaceID int32
	CachePath   string
	Identifier  string
	SourceURL   *spec.RemoteSource
}

func (m *ManagerV1) copyPluginToInstalledDir(args *CopyPluginArgs) error {
	pluginDir := m.installedPluginDir(args.Identifier)

	// Remove the directory if it already exists
	if err := os.RemoveAll(pluginDir); err != nil {
		return seer.Wrap("remove_plugin_dir", err)
	}

	// Create the directory
	if err := os.MkdirAll(pluginDir, os.ModePerm); err != nil {
		return seer.Wrap("mkdir_plugin_dir", err)
	}

	filesToCopy := []string{
		PluginConfigFileName,
		path.Join(DirOutput, OutputWasmFile),
		path.Join(DirOutput, OutputSHA256File),
	}

	// Copy the plugin config file to the installed directory
	for _, file := range filesToCopy {
		src := path.Join(args.CachePath, file)

		filename := path.Base(file)
		dest := path.Join(pluginDir, filename)

		if err := lib.CopyFile(src, dest); err != nil {
			return seer.Wrap("copy_plugin_file", err)
		}
	}

	// Create the .meta file
	metaFile := path.Join(pluginDir, ".meta")
	meta := fmt.Sprintf(
		`{"workspace_id": %d, "source": %s}`,
		args.WorkspaceID,
		args.SourceURL.RawURL(),
	)

	if err := os.WriteFile(metaFile, []byte(meta), 0644); err != nil {
		return seer.Wrap("write_plugin_meta_file", err)
	}

	// Copy the `output/deps` directory to the installed directory if it exists
	depsDir := path.Join(args.CachePath, DirOutput, "deps")
	if _, err := os.Stat(depsDir); err == nil {
		if err := lib.CopyDir(depsDir, path.Join(pluginDir, "deps")); err != nil {
			return seer.Wrap("copy_plugin_deps_dir", err)
		}
	}

	// Clear the compilation cache directory if it exists
	cachePath := path.Join(pluginDir, DirCompilationCache)
	if err := os.RemoveAll(cachePath); err != nil {
		return seer.Wrap("remove_plugin_cache_dir", err)
	}

	return nil
}

// GeneratePluginIdentifier implements spec.Manager.
func (m *ManagerV1) GeneratePluginIdentifier(args *spec.GeneratePluginIdentifierArgs) string {
	canonical := fmt.Sprintf("%d|%s|%s", args.WorkspaceID, args.SourceURL, args.PluginName)
	hash := sha256.Sum256([]byte(canonical))

	return hex.EncodeToString(hash[:16])
}

// Close implements spec.Manager.
func (m *ManagerV1) Close() error {
	if m.handle != nil {
		if err := m.handle.Close(); err != nil {
			return fmt.Errorf("failed to close plugins directory handle: %w", err)
		}
	}

	m.handle = nil
	return nil
}

// RemoveSource implements spec.Managers
func (m *ManagerV1) RemoveSource(workspaceID int32, source *spec.RemoteSource) error {
	if source == nil {
		return seer.New("remove_source", "source is nil")
	}

	if m.BelongsToCore(source) {
		return apperrors.BadRequest("cannot remove core source")
	}

	exists, err := m.repository.PluginRepository().SourceExists(&repository.FindSourceByURLArgs{
		WorkspaceID: workspaceID,
		Source:      source,
	})
	if err != nil {
		return seer.Wrap("source_exists", err)
	}

	if !exists {
		return apperrors.BadRequest("source does not exist")
	}

	workspace, err := m.repository.WorkspaceRepository().FindByInternalID(workspaceID)
	if err != nil {
		return seer.Wrap("find_workspace", err)
	}

	// Remove all plugins installed for this source
	if err := m.repository.PluginRepository().RemoveSourcePlugins(workspaceID, source); err != nil {
		return seer.Wrap("remove_source_plugins", err)
	}

	// Remove the source from the database
	if err := m.repository.PluginRepository().RemoveSource(&repository.DeleteSourceArgs{
		WorkspaceID: workspaceID,
		Source:      source,
	}); err != nil {
		return seer.Wrap("remove_source", err)
	}

	// Clean out the repository folders
	if err := m.removeSourceDirectories(workspace.ID, source.Name()); err != nil {
		return seer.Wrap("remove_source_directories", err)
	}

	return nil
}

// GetSourcePlugins implements spec.Manager.
func (m *ManagerV1) GetSourcePlugins(remote *spec.RemoteSource) ([]spec.Plugin, error) {
	tempSourceDir := path.Join(m.config.Plugins.Directory, DirTmp, remote.Name())

	if err := os.RemoveAll(tempSourceDir); err != nil {
		return nil, seer.Wrap("remove_tmp_source_dir", err)
	}

	repo, err := m.cloneRepository(remote)
	if err != nil {
		return nil, seer.Wrap("clone_repository_in_get_source_plugins", err)
	}

	// Load and parse the source config
	source, err := m.loadSourceConfig(repo)
	if err != nil {
		return nil, err
	}

	// Load the source's plugins
	plugins, err := m.loadSourcePlugins(repo, source)
	if err != nil {
		return nil, seer.Wrap("load_source_plugins", err)
	}

	return plugins, nil
}

// getSourcePluginsFromTmp loads plugins from an existing temporary source directory
func (m *ManagerV1) GetSourcePluginsFromCache(remote *spec.RemoteSource) ([]spec.Plugin, error) {
	tempSourceDir := path.Join(m.config.Plugins.Directory, DirTmp, remote.Name())

	repo, err := git.PlainOpen(tempSourceDir)
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			return nil, seer.New("get_source_plugins_from_tmp", "source does not exist")
		}
	}

	// Load and parse the source config
	source, err := m.loadSourceConfig(repo)
	if err != nil {
		return nil, err
	}

	// Load the source's plugins
	plugins, err := m.loadSourcePlugins(repo, source)
	if err != nil {
		return nil, seer.Wrap("load_source_plugins_from_cache", err)
	}

	return plugins, nil
}

// FetchRemoteSource implements spec.Manager.
func (m *ManagerV1) FetchRemoteSource(url *spec.RemoteSource) (spec.SourceWithPlugins, error) {
	repo, err := m.cloneRepository(url)
	if err != nil {
		return spec.SourceWithPlugins{}, err
	}

	// Load and parse the source config
	source, err := m.loadSourceConfig(repo)
	if err != nil {
		return spec.SourceWithPlugins{}, err
	}

	version, err := m.getVersionID(repo, source)
	if err != nil {
		return spec.SourceWithPlugins{}, seer.Wrap("get_version_id", err)
	}

	// Load the source's plugins
	plugins, err := m.loadSourcePlugins(repo, source)
	if err != nil {
		return spec.SourceWithPlugins{}, seer.Wrap("load_source_plugins", err)
	}

	return spec.SourceWithPlugins{
		Source:  source,
		Plugins: plugins,
		Version: version,
	}, nil
}

// AddRemoteSource implements spec.Manager.
func (m *ManagerV1) AddRemoteSource(args spec.AddRemoteSourceArgs) (spec.SourceWithPlugins, error) {
	// Check if the source already exists
	exists, err := m.repository.PluginRepository().SourceExists(&repository.FindSourceByURLArgs{
		WorkspaceID: args.WorkspaceID,
		Source:      args.Remote,
	})
	if err != nil {
		return spec.SourceWithPlugins{}, seer.Wrap("source_exists", err)
	}

	if exists {
		return spec.SourceWithPlugins{}, apperrors.BadRequest("source already exists")
	}

	result, err := m.FetchRemoteSource(args.Remote)
	if err != nil {
		return spec.SourceWithPlugins{}, seer.Wrap("fetch_remote_source", err)
	}

	_, err = m.repository.PluginRepository().
		CreateRemoteSource(&repository.CreateRemoteSourceArgs{
			WorkspaceID:        args.WorkspaceID,
			Name:               result.Source.Name(),
			Description:        result.Source.Description(),
			Author:             result.Source.Author(),
			VersioningStrategy: queries.VersioningStrategy(result.Source.VersioningStrategy()),
			GitURL:             args.Remote.RawURL(),
			AuthMethod:         queries.PluginSourceAuthMethodNone,
			VersionID:          result.Version,
		})
	if err != nil {
		return spec.SourceWithPlugins{}, seer.Wrap("create_remote_source", err)
	}

	return result, nil
}

func (m *ManagerV1) loadSourcePlugins(
	repo *git.Repository,
	source *SourceV1,
) ([]spec.Plugin, error) {
	if len(source.Plugins()) == 0 {
		return nil, apperrors.BadRequest("no plugins found in source")
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, seer.Wrap("get_default_worktree", err)
	}
	if worktree == nil {
		return nil, seer.New("get_default_worktree", "worktree is nil")
	}

	// If versioning if by tag, checkout the latest tag
	if source.VersioningStrategy() == spec.VersioningStrategyTag {
		versionID, err := m.getVersionID(repo, source)
		if err != nil {
			return nil, seer.Wrap("get_version_id_in_cache_fn", err)
		}

		tagRef, err := repo.Tag(versionID)
		if err != nil {
			return nil, seer.Wrap("get_tag_ref", err)
		}

		//nolint:all
		if err := worktree.Checkout(&git.CheckoutOptions{
			Hash:  tagRef.Hash(),
			Force: true,
		}); err != nil {
			return nil, seer.Wrap("checkout_latest_tag", err)
		}
	}

	wg := sync.WaitGroup{}
	plugins := make([]spec.Plugin, 0)
	for _, dirname := range source.Plugins() {
		wg.Add(1)
		go func(dirname string) {
			defer wg.Done()

			plugin, err := m.getPlugin(worktree.Filesystem.Root(), dirname)
			if err != nil {
				log.Error().
					Err(err).
					Str("plugin_path", path.Join(worktree.Filesystem.Root(), dirname)).
					Msg("failed to load plugin")
				return
			}

			plugins = append(plugins, plugin)
		}(dirname)
	}
	wg.Wait()

	if len(plugins) == 0 {
		return nil, seer.New("cache_source_plugins", "no plugins found")
	}

	return plugins, nil
}

func (m *ManagerV1) getPlugin(sourceRoot, relativePluginPath string) (*PluginV1, error) {
	pluginPath := path.Join(sourceRoot, relativePluginPath)
	stat, err := os.Stat(pluginPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, seer.New("get_plugin", "plugin path does not exist")
		}
		return nil, seer.Wrap("stat_plugin_path", err)
	}

	if !stat.IsDir() {
		return nil, seer.New("get_plugin", "plugin path is not a directory")
	}

	// Check if the plugin directory contains a plugin.toml file
	pluginConfigPath := path.Join(pluginPath, PluginConfigFileName)
	if err := lib.FileExists(pluginConfigPath); err != nil {
		return nil, seer.Wrap("fileexists_plugin_config", err)
	}

	// Check if the output folder exists
	outputDirPath := path.Join(pluginPath, DirOutput)
	if err := lib.DirExists(outputDirPath); err != nil {
		return nil, seer.Wrap("direxists_plugin_output_path", err)
	}

	// Check for the output WASM and the SHA256 file
	var (
		outputWasmPath   = path.Join(outputDirPath, OutputWasmFile)
		outputSha256Path = path.Join(outputDirPath, OutputSHA256File)
		hasSHA256        bool
	)
	if err := lib.FileExists(outputWasmPath); err != nil {
		return nil, seer.Wrap("fileexists_plugin_output_wasm", err)
	}
	if err := lib.FileExists(outputSha256Path); err == nil {
		hasSHA256 = true
	}

	// Read the plugin config
	content, err := os.ReadFile(pluginConfigPath)
	if err != nil {
		return nil, seer.Wrap("read_plugin_config", err)
	}

	plugin, err := ParsePluginV1(content)
	if err != nil {
		return nil, seer.Wrap("parse_plugin_config", err)
	}

	// Set the current checksum for the plugin
	generatedChecksum, err := lib.CalculateChecksum(outputWasmPath)
	if err != nil {
		return nil, seer.Wrap("calculate_plugin_checksum", err, "failed to calculate checksum")
	}
	plugin.OutputChecksum = generatedChecksum

	// Verify the SHA256 file if it exists
	if !plugin.HasVerificationSHA256() && hasSHA256 {
		plugin.ProvidedChecksum, err = os.ReadFile(outputSha256Path)
		if err != nil {
			return nil, seer.Wrap("read_plugin_sha256", err)
		}

		if matches := lib.CompareBytes(plugin.OutputChecksum, plugin.ProvidedChecksum); !matches {
			return nil, seer.New("compare_sha256", "SHA256 does not match, file may be corrupted")
		}
	}

	plugin.AbsolutePathInCache = pluginPath
	return plugin, nil
}

func (m *ManagerV1) cloneRepository(
	remote *spec.RemoteSource,
) (*git.Repository, error) {
	tempSourceDir := path.Join(m.config.Plugins.Directory, DirTmp, remote.Name())

	// Remove the directory if it already exists
	if err := os.RemoveAll(tempSourceDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, seer.Wrap("remove_tmp_source_dir", err)
	}

	// The temporary source directory is used to process the source's content
	if err := os.MkdirAll(tempSourceDir, os.ModePerm); err != nil {
		return nil, seer.Wrap("mkdir_tmp_source", err)
	}

	// NOTE: we can clone in-memory but for safety reasons, we obviously will not do that
	repo, err := git.PlainClone(tempSourceDir, false, &git.CloneOptions{
		URL:               remote.RawURL(),
		Depth:             1,
		ShallowSubmodules: true,
		Tags:              git.AllTags,
	})
	if err != nil {
		return nil, seer.Wrap("clone_source_to_tmp", err)
	}

	return repo, nil
}

func (m *ManagerV1) getVersionID(repo *git.Repository, source spec.Source) (string, error) {
	var versionID string
	switch source.VersioningStrategy() {
	case spec.VersioningStrategyCommit:
		latestRef, err := repo.Head()
		if err != nil {
			return "", seer.Wrap("get_head_ref", err)
		}

		commit, err := repo.CommitObject(latestRef.Hash())
		if err != nil {
			return "", seer.Wrap("get_commit_object", err)
		}

		versionID = commit.Hash.String()

	case spec.VersioningStrategyTag:
		tags, err := repo.Tags()
		if err != nil {
			return "", seer.Wrap("get_tags", err)
		}
		defer tags.Close()

		latestTag, err := tags.Next()
		if err != nil {
			return "", seer.Wrap("get_latest_tag", err)
		}

		if latestTag == nil {
			return "", seer.New("get_latest_tag", "latest tag is nil")
		}

		versionID = latestTag.Name().String()

	default:
		return "", seer.New("invalid_versioning_strategy", "invalid versioning strategy")
	}

	return versionID, nil
}

func (m *ManagerV1) loadSourceConfig(repo *git.Repository) (*SourceV1, error) {
	// Process the source metadata
	worktree, err := repo.Worktree()
	if err != nil {
		return &SourceV1{}, seer.Wrap("get_default_worktree", err)
	}
	if worktree == nil {
		return &SourceV1{}, seer.New("get_default_worktree", "worktree is nil")
	}

	sourceConfig, err := worktree.Filesystem.Open("source.toml")
	if err != nil {
		return &SourceV1{}, seer.Wrap("open_source_config", err)
	}
	defer sourceConfig.Close() //nolint:all

	sourceBytes, err := io.ReadAll(sourceConfig)
	if err != nil {
		log.Debug().Str("source_config", string(sourceBytes)).Msg("source config")
		return &SourceV1{}, seer.Wrap("read_source_config", err)
	}

	source, err := ParseSourceV1(sourceBytes)
	if err != nil {
		return &SourceV1{}, seer.Wrap("parse_source_config", err, "source has an invalid config")
	}

	return source, nil
}

func (m *ManagerV1) removeSourceDirectories(workspaceUUID string, name string) error {
	// Remote the tmp directories
	tmpDir := path.Join(m.config.Plugins.Directory, DirTmp, name)
	if err := os.RemoveAll(tmpDir); err != nil {
		return seer.Wrap("remove_tmp_source_dir", err)
	}

	// Remove the source directories - `.plugins/<workspace_id>/<source_name>/[plugins]`
	sourceDir := path.Join(m.config.Plugins.Directory, workspaceUUID, name)
	if err := os.RemoveAll(sourceDir); err != nil {
		return seer.Wrap("remove_source_dir", err)
	}

	return nil
}

var _ spec.Manager = (*ManagerV1)(nil)
