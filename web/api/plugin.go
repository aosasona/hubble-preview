package api

import (
	"slices"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/hubble/web/pkg/rbac"
	"go.trulyao.dev/robin"
	"golang.org/x/sync/errgroup"
)

type pluginHandler struct {
	*baseHandler
}

// RemovePlugin implements PluginHandler.
func (p *pluginHandler) RemovePlugin(
	ctx *robin.Context,
	request RemovePluginRequest,
) (PluginActionResponse, error) {
	var response PluginActionResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:all
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermUninstallPlugin) {
		return response, rbac.ErrPermissionDenied
	}

	plugin, err := p.repos.PluginRepository().FindInstalledPlugin(result.ID, request.PluginID)
	if err != nil {
		return response, err
	}

	if err := p.pluginManager.RemovePlugin(&spec.RemovePluginArgs{
		WorkspaceID: result.ID,
		Identifier:  plugin.PluginIdentifier,
	}); err != nil {
		log.Error().Err(err).Msg("failed to remove plugin")
		return response, apperrors.ServerError("failed to remove plugin")
	}

	return PluginActionResponse{
		PluginName: plugin.Name(),
		SourceName: plugin.Source.Name,
	}, nil
}

// InstallPlugin implements PluginHandler.
func (p *pluginHandler) InstallPlugin(
	ctx *robin.Context,
	request InstallPluginRequest,
) (PluginActionResponse, error) {
	return p.addPlugin(addPluginArgs{
		ctx:     ctx,
		request: request,
		update:  false,
	})
}

// UpdatePlugin implements PluginHandler.
func (p *pluginHandler) UpdatePlugin(
	ctx *robin.Context,
	request InstallPluginRequest,
) (PluginActionResponse, error) {
	return p.addPlugin(addPluginArgs{
		ctx:     ctx,
		request: request,
		update:  true,
	})
}

// AddPlugin implements PluginHandler.
type addPluginArgs struct {
	ctx     *robin.Context
	request InstallPluginRequest
	update  bool
}

func (p *pluginHandler) addPlugin(args addPluginArgs) (PluginActionResponse, error) {
	var response PluginActionResponse

	auth, err := authlib.ExtractAuthSession(args.ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(args.request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:all
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermInstallPlugin) {
		return response, rbac.ErrPermissionDenied
	}

	source, err := p.repos.PluginRepository().FindSourceByID(&repository.FindSourceByIDArgs{
		WorkspaceID: result.ID,
		SourceID:    args.request.SourceID,
	})
	if err != nil {
		return response, err
	}

	err = p.pluginManager.InstallPlugin(&spec.InstallPluginArgs{
		WorkspaceID: result.ID,
		PluginName:  args.request.Name,
		Source:      &source.SourceURL,
		PullLatest:  args.update,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to install plugin")

		msg := "failed to install plugin"
		if args.update {
			msg = "failed to update plugin"
		}
		return response, apperrors.ServerError(msg)
	}

	return PluginActionResponse{
		PluginName: args.request.Name,
		SourceName: source.Name,
	}, nil
}

// ListPlugins implements PluginHandler.
func (p *pluginHandler) ListPlugins(
	ctx *robin.Context,
	request ListPluginsRequest,
) (ListPluginsResponse, error) {
	var response ListPluginsResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:all
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	addedSources, err := p.repos.PluginRepository().
		FindSourcesByWorkspaceID(&repository.FindSourcesByWorkspaceIDArgs{
			WorkspceID: result.ID,
			Pagination: repository.PaginationParams{
				Page:    1,
				PerPage: 1_000, // No way you actually have a thousand sources, right? right, CHARLIE? RIGHT????
			},
		})
	if err != nil {
		return response, err
	}

	sourcesIdMap := make(map[string]pgtype.UUID)
	for i := range addedSources.List {
		addedSource := &addedSources.List[i]
		sourcesIdMap[addedSource.SourceURL.String()] = addedSource.ID
	}

	installedPlugins, err := p.repos.PluginRepository().
		FindInstalledPluginsByWorkspaceID(&repository.FindInstalledPluginsByWorkspaceIDArgs{
			WorkspaceID: result.ID,
		})
	if err != nil {
		return response, err
	}

	getPluginState := func(plugin spec.Plugin, sourceURL string) PluginState {
		var state PluginState
		if sid, ok := sourcesIdMap[sourceURL]; ok {
			state.SourceID = sid
		}

		for i := range installedPlugins {
			installedPlugin := installedPlugins[i]

			if installedPlugin.Name() == plugin.Name() &&
				installedPlugin.Source.URL.String() == sourceURL {

				state.SourceID = installedPlugin.SourceID
				state.Identifier = installedPlugin.PluginIdentifier
				state.Installed = true
				if state.Installed {
					state.Updatable = plugin.IsDifferentTo(installedPlugin.InstalledPlugin)
				}
				break
			}
		}

		return state
	}

	eg := new(errgroup.Group)
	eg.SetLimit(10)
	plugins := make([]PluginListItem, 0)

	// Load all plugins for the sources
	for i := range addedSources.List {
		addedSource := &addedSources.List[i]
		eg.Go(func() error {
			source, err := p.pluginManager.FetchRemoteSource(&addedSource.SourceURL)
			if err != nil {
				return err
			}

			for j := range source.Plugins {
				plugin := source.Plugins[j]
				state := getPluginState(plugin, source.Source.URL())
				item := PluginListItem{
					Identifier:  state.Identifier,
					Name:        plugin.Name(),
					Description: plugin.Description(),
					Author:      source.Source.Author(),
					Privileges:  plugin.Privileges(),
					Source: PluginListItemSource{
						ID:   state.SourceID,
						Name: source.Source.Name(),
						URL:  source.Source.URL(),
					},
					Installed: state.Installed,
					Updatable: state.Updatable,
					Targets:   plugin.Targets(),
				}
				plugins = append(plugins, item)
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return response, err
	}

	slices.SortFunc(plugins, func(a PluginListItem, b PluginListItem) int {
		if a.Name == b.Name {
			return 0
		}
		if a.Name < b.Name {
			return -1
		}
		return 1
	})

	return ListPluginsResponse{Plugins: plugins}, nil
}

// RemoveSource implements PluginHandler.
func (p *pluginHandler) RemoveSource(
	ctx *robin.Context,
	request RemoveSourceRequest,
) (RemoveSourceResponse, error) {
	var response RemoveSourceResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:all
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermRemovePluginSource) {
		return response, rbac.ErrPermissionDenied
	}

	source, err := p.repos.PluginRepository().FindSourceByID(&repository.FindSourceByIDArgs{
		WorkspaceID: result.ID,
		SourceID:    request.SourceID,
	})
	if err != nil {
		return response, err
	}

	if err := p.pluginManager.RemoveSource(result.ID, &source.SourceURL); err != nil {
		return response, err
	}

	return RemoveSourceResponse{WorkspaceSlug: result.Workspace.Slug}, nil
}

// ListSources implements PluginHandler.
func (p *pluginHandler) ListSources(
	ctx *robin.Context,
	request ListSourcesRequest,
) (ListSourcesResponse, error) {
	var response ListSourcesResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId}, //nolint:all
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermListPluginSources) {
		return response, rbac.ErrPermissionDenied
	}

	sources, err := p.repos.PluginRepository().
		FindSourcesByWorkspaceID(&repository.FindSourcesByWorkspaceIDArgs{
			WorkspceID: result.ID,
			Pagination: request.Pagination,
		})
	if err != nil {
		return response, err
	}

	return ListSourcesResponse{
		Sources: sources.List,
		// Pagination: request.Pagination.ToState(sources.TotalCount),
		Pagination: request.Pagination.ToState(repository.PageStateArgs{
			CurrentCount: len(sources.List),
			TotalCount:   sources.TotalCount,
		}),
	}, nil
}

// AddSource implements PluginHandler.
func (p *pluginHandler) AddSource(
	ctx *robin.Context,
	request AddSourceRequest,
) (AddSourceResponse, error) {
	var response AddSourceResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId, Slug: ""},
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermAddPluginSource) {
		return response, rbac.ErrPermissionDenied
	}

	source, err := spec.ParseRemoteSource(request.URL)
	if err != nil {
		return response, apperrors.NewValidationError(apperrors.ErrorMap{
			"url": {"invalid url"},
		})
	}

	exists, err := p.repos.PluginRepository().SourceExists(&repository.FindSourceByURLArgs{
		WorkspaceID: result.ID,
		Source:      &source,
	})
	if err != nil {
		return response, err
	}
	if exists {
		return response, apperrors.NewValidationError(apperrors.ErrorMap{
			"url": {"Source has already been added to this workspace"},
		})
	}

	created, err := p.pluginManager.AddRemoteSource(spec.AddRemoteSourceArgs{
		WorkspaceID: result.ID,
		Remote:      &source,
	})
	if err != nil {
		return response, err
	}

	return AddSourceResponse{Source: created.Source}, nil
}

// FindSourceByURL implements PluginHandler.
func (p *pluginHandler) FindSourceByURL(
	ctx *robin.Context,
	request FindSourceByURLRequest,
) (FindSourceByURLResponse, error) {
	var response FindSourceByURLResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	workspaceId, err := lib.UUIDFromString(request.WorkspaceID)
	if err != nil {
		return response, err
	}

	result, err := p.repos.WorkspaceRepository().FindWithMembershipStatus(
		repository.PublicIdOrSlug{PublicID: workspaceId, Slug: ""},
		auth.UserID,
	)
	if err != nil {
		return response, err
	}

	if !result.MembershipStatus.Role.Can(rbac.PermViewPluginSource) {
		return response, rbac.ErrPermissionDenied
	}

	remoteSource, err := spec.ParseRemoteSource(request.URL)
	if err != nil {
		return response, apperrors.NewValidationError(apperrors.ErrorMap{
			"url": {"invalid url"},
		})
	}

	sourceWithPlugins, err := p.pluginManager.FetchRemoteSource(&remoteSource)
	if err != nil {
		return response, err
	}

	return FindSourceByURLResponse{
		Source:  sourceWithPlugins.Source,
		Plugins: sourceWithPlugins.Plugins,
	}, nil
}

var _ PluginHandler = (*pluginHandler)(nil)
