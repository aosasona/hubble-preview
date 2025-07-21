package models

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/pkg/document"
)

type InstalledPlugin struct {
	ID pgtype.UUID `json:"id"              mirror:"type:string"`
	// A unique identifier for the plugin, this is generated in the system as a hash from the source data and the workspace itself. It is also used to identify local files related to the plugin.
	PluginIdentifier  string               `json:"identifier"`
	WorkspaceID       int32                `json:"workspace_id"`
	SourceID          pgtype.UUID          `json:"source_id"       mirror:"type:string"`
	PluginName        string               `json:"name"`
	PluginDescription pgtype.Text          `json:"description"     mirror:"type:string"`
	Scope             queries.PluginScope  `json:"scope"           mirror:"type:'global' | 'workspace'"`
	ExecutionModes    []queries.PluginMode `json:"modes"           mirror:"type:Array<'on_create' | 'background'>"`
	Types             []document.EntryType `json:"targets"         mirror:"type:Array<'link' | 'audio' | 'video' | 'image' | 'pdf' | 'interchange' | 'epub' | 'word_document' | 'presentation' | 'spreadsheet' | 'html' | 'markdown' | 'plain_text' | 'archive' | 'code' | 'comment' | 'other' | '*'>"`
	// The SHA-256 of the WASM file provided as part of the build
	VersionSha         string          `json:"version_sha"`
	LastUpdatedAt      time.Time       `json:"last_updated_at"`
	AddedAt            time.Time       `json:"added_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
	Metadata           []byte          `json:"metadata"`
	Tags               []string        `json:"tags"`
	DeclaredPrivileges spec.Privileges `json:"privileges"`
}

// Description implements spec.ComparablePlugin.
func (i *InstalledPlugin) Description() string {
	return i.PluginDescription.String
}

// Name implements spec.ComparablePlugin.
func (i *InstalledPlugin) Name() string {
	return i.PluginName
}

// Modes implements spec.ComparablePlugin.
func (i *InstalledPlugin) Modes() []queries.PluginMode {
	return i.ExecutionModes
}

// Checksum implements spec.ComparablePlugin.
func (i *InstalledPlugin) Checksum() string {
	return i.VersionSha
}

// Privileges implements spec.ComparablePlugin.
func (i *InstalledPlugin) Privileges() spec.Privileges {
	return i.DeclaredPrivileges
}

// Targets implements spec.ComparablePlugin.
func (i *InstalledPlugin) Targets() []document.EntryType {
	return i.Types
}

func (i *InstalledPlugin) From(plugin *queries.InstalledPlugin) {
	specPrivileges := make([]spec.Privilege, 0)
	for _, privilege := range plugin.Privileges {
		identifier, err := spec.ParsePerm(privilege.Identifier)
		if err != nil {
			continue
		}

		specPrivileges = append(specPrivileges, spec.Privilege{
			Identifier:  identifier,
			Description: privilege.Description,
		})
	}

	*i = InstalledPlugin{
		ID:                 plugin.ID,
		PluginIdentifier:   plugin.PluginIdentifier,
		WorkspaceID:        plugin.WorkspaceID.Int32,
		SourceID:           plugin.SourceID,
		PluginName:         plugin.Name,
		PluginDescription:  plugin.Description,
		Scope:              plugin.Scope,
		ExecutionModes:     plugin.Modes,
		Types:              plugin.EntryTypes,
		VersionSha:         plugin.VersionSha,
		LastUpdatedAt:      plugin.LastUpdatedAt.Time,
		AddedAt:            plugin.AddedAt.Time,
		UpdatedAt:          plugin.UpdatedAt.Time,
		Metadata:           plugin.Metadata,
		Tags:               plugin.Tags,
		DeclaredPrivileges: specPrivileges,
	}
}

type PluginSource struct {
	ID pgtype.UUID `json:"id"                  mirror:"type:string"`
	// The workspace that this plugin source was added to. This is used to determine the scope of the plugin source.
	WorkspaceID        int32                          `json:"workspace_id"`
	Name               string                         `json:"name"`
	Description        string                         `json:"description"`
	Author             string                         `json:"author"`
	DisabledAt         time.Time                      `json:"disabled_at"`
	VersioningStrategy queries.VersioningStrategy     `json:"versioning_strategy" mirror:"type:'commit' | 'tag'"`
	SourceURL          spec.RemoteSource              `json:"source_url"          mirror:"type:string"`
	AuthMethod         queries.PluginSourceAuthMethod `json:"-"`
	VersionID          string                         `json:"version_id"`
	SyncStatus         queries.PluginSyncStatus       `json:"sync_status"`
	LastSyncError      string                         `json:"last_sync_error"`
	LastSyncedAt       time.Time                      `json:"last_synced_at"`
	AddedAt            time.Time                      `json:"added_at"`
	UpdatedAt          time.Time                      `json:"updated_at"`
}

func (p *PluginSource) IsDisabled() bool {
	return p.DisabledAt != (time.Time{})
}

// From converts a PluginSource from the database to a PluginSource model.
func (p *PluginSource) From(source *queries.PluginSource) {
	sourceURL, err := spec.ParseRemoteSource(source.GitRemote.String)
	if err != nil {
		log.Error().Err(err).Msg("failed to parse remote source")
	}

	*p = PluginSource{
		ID:                 source.ID,
		WorkspaceID:        source.WorkspaceID.Int32,
		Name:               source.Name,
		Description:        source.Description.String,
		Author:             source.Author,
		DisabledAt:         source.DisabledAt.Time,
		VersioningStrategy: source.VersioningStrategy,
		SourceURL:          sourceURL,
		AuthMethod:         source.AuthMethod,
		VersionID:          source.VersionID.String,
		SyncStatus:         source.SyncStatus,
		LastSyncError:      source.LastSyncError.String,
		LastSyncedAt:       source.LastSyncedAt.Time,
		AddedAt:            source.AddedAt.Time,
		UpdatedAt:          source.UpdatedAt.Time,
	}
}

var _ spec.ComparablePlugin = (*InstalledPlugin)(nil)
