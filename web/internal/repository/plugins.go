package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/pkg/document"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
)

var ErrSourceNotFound = apperrors.BadRequest("plugin source not found")

type (
	CreateRemoteSourceArgs struct {
		WorkspaceID        int32
		Name               string
		Description        string
		Author             string
		VersioningStrategy queries.VersioningStrategy
		GitURL             string
		AuthMethod         queries.PluginSourceAuthMethod
		VersionID          string
	}

	FindSourceByURLArgs struct {
		WorkspaceID int32
		Source      *spec.RemoteSource
	}

	FindSourceByIDArgs struct {
		WorkspaceID int32
		SourceID    string
	}

	DeleteSourceArgs struct {
		WorkspaceID int32
		Source      *spec.RemoteSource
		SourceID    string
	}

	SourceExistsArgs struct {
		WorkspaceID int32
		Source      *spec.RemoteSource
	}

	FindSourcesByWorkspaceIDArgs struct {
		WorkspceID int32
		Pagination PaginationParams
	}
	FindSourcesByWorkspaceIDResult struct {
		TotalCount int64
		List       []models.PluginSource
	}

	FindInstalledPluginsByWorkspaceIDArgs struct {
		WorkspaceID int32
	}

	PluginSourceDetails struct {
		ID   pgtype.UUID       `json:"id"         mirror:"type:string"`
		Name string            `json:"name"`
		URL  spec.RemoteSource `json:"source_url" mirror:"type:string"`
	}

	InstalledPluginWithSource struct {
		*models.InstalledPlugin `                    json:"plugin"`
		Source                  PluginSourceDetails `json:"source"`
	}

	UpsertPluginArgs struct {
		Identifier          string `validate:"required,ascii"`
		WorkspaceID         int32  `validate:"required"`
		SourceID            pgtype.UUID
		Name                string `validate:"required,mixed_name"`
		Description         string `validate:"required,ascii"`
		Modes               []queries.PluginMode
		Targets             []document.EntryType
		Checksum            string `validate:"required,alphanum"`
		PluginLastUpdatedAt time.Time
		Privileges          []queries.PluginPrivilege
	}

	RemovePluginArgs struct {
		WorkspaceID int32
		Identifier  string
	}

	FindOnCreatePluginForEntryArgs struct {
		EntryID           int32
		WorkspacePublicID pgtype.UUID
	}
)

type PluginRepository interface {
	// CreatePluginSource creates a new plugin source in the database.
	CreateRemoteSource(args *CreateRemoteSourceArgs) (*models.PluginSource, error)

	// UpsertInstalledPlugin upserts an installed plugin in the database.
	UpsertInstalledPlugin(args *UpsertPluginArgs) (models.InstalledPlugin, error)

	// RemoveInstalledPlugin removes an installed plugin from the database.
	RemoveInstalledPlugin(args *RemovePluginArgs) error

	// FindSourceByURL finds a plugin source in a workspace by its URL
	FindSourceByURL(args *FindSourceByURLArgs) (*models.PluginSource, error)

	// FindSourceByID finds a plugin source in a workspace by its ID
	FindSourceByID(args *FindSourceByIDArgs) (*models.PluginSource, error)

	// FindInstalledPluginsByWorkspaceID finds all installed plugins in a workspace.
	FindInstalledPluginsByWorkspaceID(
		args *FindInstalledPluginsByWorkspaceIDArgs,
	) ([]InstalledPluginWithSource, error)

	// FindInstalledPlugin finds an installed plugin in a workspace by its identifier.
	FindInstalledPlugin(workspaceID int32, identifier string) (InstalledPluginWithSource, error)

	// FindSourcesByWorkspaceID finds all plugin sources in a workspace.
	FindSourcesByWorkspaceID(
		args *FindSourcesByWorkspaceIDArgs,
	) (FindSourcesByWorkspaceIDResult, error)

	// RemoveSource removes a plugin source from a workspace.
	RemoveSource(args *DeleteSourceArgs) error

	// RemoveSourcePlugins removes all plugins associated with a source.
	RemoveSourcePlugins(workspaceId int32, source *spec.RemoteSource) error

	// SourceExists checks if a plugin source exists in a workspace by its URL.
	SourceExists(args *FindSourceByURLArgs) (bool, error)

	FindOnCreatePluginForEntry(
		ctx context.Context,
		args *FindOnCreatePluginForEntryArgs,
	) ([]models.InstalledPlugin, error)
}

type pluginRepo struct {
	*baseRepo
}

// FindOnCreatePluginForEntry implements PluginRepository.
func (p *pluginRepo) FindOnCreatePluginForEntry(
	ctx context.Context,
	args *FindOnCreatePluginForEntryArgs,
) ([]models.InstalledPlugin, error) {
	rows, err := p.queries.FindOnCreatePluginsForType(ctx, queries.FindOnCreatePluginsForTypeParams{
		EntryID:           args.EntryID,
		WorkspacePublicID: args.WorkspacePublicID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	plugins := make([]models.InstalledPlugin, 0)
	for i := range rows {
		row := &rows[i]
		plugin := new(models.InstalledPlugin)
		plugin.From(&row.InstalledPlugin)
		plugins = append(plugins, *plugin)
	}

	return plugins, nil
}

// RemoveInstalledPlugin implements PluginRepository.
func (p *pluginRepo) RemoveInstalledPlugin(args *RemovePluginArgs) error {
	return p.queries.RemoveInstalledPlugin(context.TODO(), queries.RemoveInstalledPluginParams{
		WorkspaceID:      lib.PgInt4(args.WorkspaceID),
		PluginIdentifier: args.Identifier,
	})
}

// UpsertInstalledPlugin implements PluginRepository.
func (p *pluginRepo) UpsertInstalledPlugin(args *UpsertPluginArgs) (models.InstalledPlugin, error) {
	if err := lib.ValidateStruct(args); err != nil {
		return models.InstalledPlugin{}, err
	}

	row, err := p.queries.UpsertInstalledPlugin(context.TODO(), queries.UpsertInstalledPluginParams{
		Identifier:          args.Identifier,
		WorkspaceID:         lib.PgInt4(args.WorkspaceID),
		SourceID:            args.SourceID,
		Name:                args.Name,
		Description:         lib.PgText(args.Description),
		Modes:               args.Modes,
		Targets:             args.Targets,
		Checksum:            args.Checksum,
		PluginLastUpdatedAt: pgtype.Timestamptz{Time: args.PluginLastUpdatedAt, Valid: true},
		Privileges:          args.Privileges,
	})
	if err != nil {
		return models.InstalledPlugin{}, err
	}

	installedPlugin := new(models.InstalledPlugin)
	installedPlugin.From(&row)
	return *installedPlugin, nil
}

// FindInstalledPlugin implements PluginRepository.
func (p *pluginRepo) FindInstalledPlugin(
	workspaceID int32,
	identifier string,
) (InstalledPluginWithSource, error) {
	row, err := p.queries.FindInstalledPlugin(context.TODO(), queries.FindInstalledPluginParams{
		WorkspaceID:      lib.PgInt4(workspaceID),
		PluginIdentifier: identifier,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InstalledPluginWithSource{}, apperrors.BadRequest(
				"plugin not installed in this workspace",
			)
		}
		return InstalledPluginWithSource{}, err
	}

	plugin := new(models.InstalledPlugin)
	plugin.From(&row.InstalledPlugin)

	remote, err := spec.ParseRemoteSource(row.SourceUrl.String)
	if err != nil {
		return InstalledPluginWithSource{}, err
	}

	return InstalledPluginWithSource{
		InstalledPlugin: plugin,
		Source: PluginSourceDetails{
			ID:   row.SourceID,
			Name: row.SourceName,
			URL:  remote,
		},
	}, nil
}

// FindInstalledPluginsByWorkspaceID implements PluginRepository.
func (p *pluginRepo) FindInstalledPluginsByWorkspaceID(
	args *FindInstalledPluginsByWorkspaceIDArgs,
) ([]InstalledPluginWithSource, error) {
	rows, err := p.queries.FindInstalledPluginsByWorkspaceID(
		context.TODO(),
		lib.PgInt4(args.WorkspaceID),
	)
	if err != nil {
		return nil, err
	}

	plugins := make([]InstalledPluginWithSource, 0)
	for i := range rows {
		row := &rows[i]
		plugin := new(models.InstalledPlugin)
		plugin.From(&row.InstalledPlugin)

		remote, err := spec.ParseRemoteSource(row.SourceUrl.String)
		if err != nil {
			log.Error().
				Err(err).
				Msg("failed to parse remote source in FindInstalledPluginsByWorkspaceID")
			continue
		}

		pluginWithSource := InstalledPluginWithSource{
			InstalledPlugin: plugin,
			Source: PluginSourceDetails{
				ID:   row.SourceID,
				Name: row.SourceName,
				URL:  remote,
			},
		}
		plugins = append(plugins, pluginWithSource)
	}

	return plugins, nil
}

// RemoveSourcePlugins implements PluginRepository.
func (p *pluginRepo) RemoveSourcePlugins(workspaceId int32, source *spec.RemoteSource) error {
	return p.queries.RemoveSourcePlugins(context.TODO(), queries.RemoveSourcePluginsParams{
		WorkspaceID: lib.PgInt4(workspaceId),
		GitRemote:   lib.PgText(source.RawURL()),
	})
}

// FindSourcesByWorkspaceID implements PluginRepository.
func (p *pluginRepo) FindSourcesByWorkspaceID(
	args *FindSourcesByWorkspaceIDArgs,
) (FindSourcesByWorkspaceIDResult, error) {
	rows, err := p.queries.FindPluginSourcesByWorkspaceID(
		context.TODO(),
		queries.FindPluginSourcesByWorkspaceIDParams{
			Limit:       args.Pagination.Limit(),
			Offset:      args.Pagination.Offset(),
			WorkspaceID: lib.PgInt4(args.WorkspceID),
		},
	)
	if err != nil {
		return FindSourcesByWorkspaceIDResult{}, err
	}

	var totalCount int64 = 0
	sources := make([]models.PluginSource, 0)

	for i := range len(rows) {
		row := &rows[i]
		if totalCount == 0 {
			totalCount = row.TotalCount
		}

		sourceURL, err := spec.ParseRemoteSource(row.GitRemote.String)
		if err != nil {
			log.Error().Err(err).Msg("failed to parse remote source in FindSourcesByWorkspaceID")
			continue
		}

		source := models.PluginSource{
			ID:                 row.ID,
			WorkspaceID:        row.WorkspaceID.Int32,
			Name:               row.Name,
			Description:        row.Description.String,
			Author:             row.Author,
			DisabledAt:         row.DisabledAt.Time,
			VersioningStrategy: row.VersioningStrategy,
			SourceURL:          sourceURL,
			AuthMethod:         row.AuthMethod,
			VersionID:          row.VersionID.String,
			SyncStatus:         row.SyncStatus,
			LastSyncError:      row.LastSyncError.String,
			LastSyncedAt:       row.LastSyncedAt.Time,
			AddedAt:            row.AddedAt.Time,
			UpdatedAt:          row.UpdatedAt.Time,
		}
		sources = append(sources, source)
	}

	return FindSourcesByWorkspaceIDResult{
		TotalCount: totalCount,
		List:       lib.WithMaxSize(sources, args.Pagination.PerPage),
	}, nil
}

// SourceExists implements PluginRepository.
func (p *pluginRepo) SourceExists(args *FindSourceByURLArgs) (bool, error) {
	exists, err := p.queries.SourceExists(context.TODO(), queries.SourceExistsParams{
		WorkspaceID: lib.PgInt4(args.WorkspaceID),
		GitRemote:   lib.PgText(args.Source.RawURL()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return exists, nil
}

// RemoveSource implements PluginRepository.
func (p *pluginRepo) RemoveSource(args *DeleteSourceArgs) error {
	return p.queries.RemovePluginSource(context.TODO(), queries.RemovePluginSourceParams{
		WorkspaceID: lib.PgInt4(args.WorkspaceID),
		SourceID:    lib.PgUUIDString(args.SourceID),
		GitRemote:   lib.PgText(args.Source.RawURL()),
	})
}

// FindSourceByID implements PluginRepository.
func (p *pluginRepo) FindSourceByID(args *FindSourceByIDArgs) (*models.PluginSource, error) {
	row, err := p.queries.FindPluginSourceByID(context.TODO(), queries.FindPluginSourceByIDParams{
		WorkspaceID: lib.PgInt4(args.WorkspaceID),
		SourceID:    lib.PgUUIDString(args.SourceID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.BadRequest("plugin source not found")
		}

		return nil, err
	}

	source := new(models.PluginSource)
	source.From(&row)
	return source, nil
}

// FindSourceByURL implements PluginRepository.
func (p *pluginRepo) FindSourceByURL(args *FindSourceByURLArgs) (*models.PluginSource, error) {
	row, err := p.queries.FindSourceByGitRemote(context.TODO(), queries.FindSourceByGitRemoteParams{
		WorkspaceID: lib.PgInt4(args.WorkspaceID),
		GitRemote:   lib.PgText(args.Source.RawURL()),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSourceNotFound
		}
		return nil, err
	}

	source := new(models.PluginSource)
	source.From(&row)
	return source, nil
}

// CreateRemoteSource implements PluginRepository.
func (p *pluginRepo) CreateRemoteSource(
	args *CreateRemoteSourceArgs,
) (*models.PluginSource, error) {
	row, err := p.queries.CreateRemotePluginSource(
		context.TODO(),
		queries.CreateRemotePluginSourceParams{
			Name:               args.Name,
			Description:        lib.PgText(args.Description),
			Author:             args.Author,
			VersioningStrategy: args.VersioningStrategy,
			GitRemote:          lib.PgText(args.GitURL),
			AuthMethod:         args.AuthMethod,
			VersionID:          lib.PgText(args.VersionID),
			WorkspaceID:        args.WorkspaceID,
		},
	)
	if err != nil {
		return &models.PluginSource{}, err
	}

	source := new(models.PluginSource)
	source.From(&row)
	return source, nil
}

var _ PluginRepository = (*pluginRepo)(nil)
