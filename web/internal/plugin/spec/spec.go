package spec

import (
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/pkg/document"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(commit,tag)
type VersioningStrategy string

type (
	SourceWithPlugins struct {
		Source  Source   `json:"source"  mirror:"type:SourceV1"`
		Plugins []Plugin `json:"plugins" mirror:"type:Array<PluginV1>"`
		Version string   `json:"version"`
	}
)

type (
	ComparablePlugin interface {
		// Name returns the name of the plugin
		Name() string

		// Description returns the description of the plugin
		Description() string

		// Checksum returns the checksum of the plugin binary
		Checksum() string

		// Targets can be any of the supported entry types or '*' for all types
		Targets() []document.EntryType

		// Privileges returns the privileges required by the plugin
		Privileges() Privileges

		// Modes returns the modes of the plugin
		Modes() []queries.PluginMode
	}

	Plugin interface {
		Name() string

		Description() string

		Modes() []queries.PluginMode

		// CachePath returns the absolute path to the plugin in the cache
		CachePath() string

		// Checksum returns the checksum of the plugin binary
		Checksum() string

		// Targets can be any of the supported entry types or '*' for all types
		Targets() []document.EntryType

		// Privileges returns the privileges required by the plugin
		Privileges() []Privilege

		// HasVerificationSHA256 return true if the plugin has provided a .sha256 file to verify the binary
		HasVerificationSHA256() bool

		// IsDifferentVersion returns true if the plugin is a different version than the one current one
		IsDifferentTo(other ComparablePlugin) bool
	}
)

type Source interface {
	Name() string

	Author() string

	URL() string

	Description() string

	// Plugins is the list of plugins in the source (their directories)
	Plugins() []string

	// VersioningStrategy is the versioning strategy of the source
	VersioningStrategy() VersioningStrategy
}

type (
	AddRemoteSourceArgs struct {
		WorkspaceID int32
		Remote      *RemoteSource
	}

	GeneratePluginIdentifierArgs struct {
		WorkspaceID int32
		PluginName  string
		SourceURL   string
	}

	InstallPluginArgs struct {
		WorkspaceID int32
		PluginName  string
		Source      *RemoteSource
		PullLatest  bool
	}

	RemovePluginArgs struct {
		WorkspaceID int32
		Identifier  string
	}
)

type Manager interface {
	// FetchRemoteSource fetches a remote source and its plugins
	FetchRemoteSource(url *RemoteSource) (SourceWithPlugins, error)

	// AddRemoteSource adds a remote source to a workspace.
	AddRemoteSource(args AddRemoteSourceArgs) (SourceWithPlugins, error)

	// GetSourcePlugins returns the plugins in a source
	GetSourcePlugins(remote *RemoteSource) ([]Plugin, error)

	// GetSourcePluginsFromCache returns the plugins in a source from the cache (tmp directory)
	GetSourcePluginsFromCache(remote *RemoteSource) ([]Plugin, error)

	// RemoveSource removes a source from a workspace.
	RemoveSource(workspaceID int32, source *RemoteSource) error

	/*
		GeneratePluginIdentifier generates a unique identifier for a plugin; this is used for things like the secure store and plugin cache.

		The identifier is a combination of the workspace ID, plugin name, and source URL. The result is a SHA-256 hash of the canonical string formed by: "<workspaceID>|<sourceURL>|<pluginName>"

		It returns the first 16 bytes of the hash (128 bits) as a hex-encoded string.
	*/
	GeneratePluginIdentifier(args *GeneratePluginIdentifierArgs) string

	// InstallPlugin installs a plugin from a source for a particular workspace.
	InstallPlugin(args *InstallPluginArgs) error

	// RemovePlugin removes a plugin from a workspace.
	RemovePlugin(args *RemovePluginArgs) error

	// BelongsToCore returns true if the source belongs to a core plugin repository
	BelongsToCore(source *RemoteSource) bool

	// InstallCorePlugins installs all core plugins for a workspace.
	InstallCorePlugins(workspaceID int32) error

	// Close closes the manager
	Close() error
}
