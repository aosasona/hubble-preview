package api

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/document"
	"go.trulyao.dev/robin"
)

type PluginHandler interface {
	// FindSourceByURL loads a remote source (without persistence) by its URL
	FindSourceByURL(
		ctx *robin.Context,
		request FindSourceByURLRequest,
	) (FindSourceByURLResponse, error)

	// AddSource adds a remote source to a workspace
	AddSource(ctx *robin.Context, request AddSourceRequest) (AddSourceResponse, error)

	// RemoveSource removes a remote source from a workspace
	RemoveSource(ctx *robin.Context, request RemoveSourceRequest) (RemoveSourceResponse, error)

	// ListSources lists all sources in a workspace
	ListSources(ctx *robin.Context, request ListSourcesRequest) (ListSourcesResponse, error)

	// ListPlugins lists all plugins in a workspace
	ListPlugins(ctx *robin.Context, request ListPluginsRequest) (ListPluginsResponse, error)

	// InstallPlugin adds a plugin to a workspace
	InstallPlugin(ctx *robin.Context, request InstallPluginRequest) (PluginActionResponse, error)

	// UpdatePlugin updates a plugin in a workspace
	UpdatePlugin(ctx *robin.Context, request InstallPluginRequest) (PluginActionResponse, error)

	// RemovePlugin removes a plugin from a workspace
	RemovePlugin(ctx *robin.Context, request RemovePluginRequest) (PluginActionResponse, error)
}

type (
	FindSourceByURLRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
		URL         string `json:"url"          validate:"required"`
	}

	FindSourceByURLResponse struct {
		Source  spec.Source   `json:"source"  mirror:"type:import('$/lib/server/types').SourceV1"`
		Plugins []spec.Plugin `json:"plugins" mirror:"type:import('$/lib/server/types').PluginV1[]"`
	}

	AddSourceRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
		URL         string `json:"url"          validate:"required"`
	}

	AddSourceResponse struct {
		Source spec.Source `json:"source" mirror:"type:import('$/lib/server/types').SourceV1"`
	}

	ListSourcesRequest struct {
		WorkspaceID string                      `json:"workspace_id" validate:"required,uuid"`
		Pagination  repository.PaginationParams `json:"pagination"   validate:"required"`
	}

	ListSourcesResponse struct {
		Sources    []models.PluginSource      `json:"sources"`
		Pagination repository.PaginationState `json:"pagination"`
	}

	RemoveSourceRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
		SourceID    string `json:"source_id"    validate:"required,uuid"`
	}

	RemoveSourceResponse struct {
		WorkspaceSlug string `json:"workspace_slug"`
	}

	ListPluginsRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
	}

	PluginListItemSource struct {
		ID   pgtype.UUID `json:"id"   mirror:"type:string"`
		Name string      `json:"name"`
		URL  string      `json:"url"`
	}

	PluginListItem struct {
		Identifier  string               `json:"identifier"`
		Name        string               `json:"name"`
		Description string               `json:"description"`
		Author      string               `json:"author"`
		Source      PluginListItemSource `json:"source"`
		Privileges  []spec.Privilege     `json:"privileges"`
		Targets     []document.EntryType `json:"targets"`
		Installed   bool                 `json:"installed"`
		Updatable   bool                 `json:"updatable"`
	}

	PluginState struct {
		SourceID   pgtype.UUID `json:"source_id"`
		Identifier string      `json:"identifier"`
		Installed  bool        `json:"installed"`
		Updatable  bool        `json:"updatable"`
	}

	ListPluginsResponse struct {
		Plugins []PluginListItem `json:"plugins"`
	}

	InstallPluginRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
		SourceID    string `json:"source_id"    validate:"required,uuid"`
		Name        string `json:"name"         validate:"required,ascii"`
	}

	RemovePluginRequest struct {
		WorkspaceID string `json:"workspace_id" validate:"required,uuid"`
		PluginID    string `json:"plugin_id"    validate:"required,uuid"`
	}

	PluginActionResponse struct {
		PluginName string `json:"plugin_name"`
		SourceName string `json:"source_name"`
	}
)
