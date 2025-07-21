package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/golang-queue/queue/core"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/plugin/host"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/llm"
	"go.trulyao.dev/hubble/web/pkg/ograph"
	"go.trulyao.dev/seer"
)

type handler struct {
	config      *config.Config
	objectStore *objectstore.Store
	repos       repository.Repository
	wasmRuntime *host.Runtime
	llm         *llm.LLM
}

func NewHandler(
	config *config.Config,
	repos repository.Repository,
	objectStore *objectstore.Store,
	wasmRuntime *host.Runtime,
	llm *llm.LLM,
) *handler {
	return &handler{
		config:      config,
		repos:       repos,
		objectStore: objectStore,
		wasmRuntime: wasmRuntime,
		llm:         llm,
	}
}

func (h *handler) HandleEntry(ctx context.Context, message core.TaskMessage) error {
	payload := new(job.EntryJob)
	if err := json.Unmarshal(message.Payload(), payload); err != nil {
		return err
	}

	entry, err := h.repos.EntryRepository().FindByID(&repository.FindbyIdArgs{
		InternalID: payload.ID,
		PublicID:   pgtype.UUID{Bytes: [16]byte{}, Valid: false},
	})
	if err != nil {
		return seer.Wrap("find_by_id_in_handle_entry", err)
	}

	// Find a installedPlugins that would work for this entry
	installedPlugins, err := h.repos.PluginRepository().
		FindOnCreatePluginForEntry(ctx, &repository.FindOnCreatePluginForEntryArgs{
			EntryID:           entry.ID,
			WorkspacePublicID: entry.Workspace.ID,
		})
	if err != nil {
		return seer.Wrap("find_on_create_plugin_for_entry", err)
	}

	// Leave if there are none
	if len(installedPlugins) == 0 {
		log.Info().
			Str("id", entry.PublicID.String()).
			Str("type", entry.Type.String()).
			Msg("no supported plugin found for entry")

		return nil
	}

	// Update the entry's status to processing
	if err := h.repos.EntryRepository().UpdateQueue(&repository.UpdateQueueArgs{
		Status:  queries.EntryStatusProcessing,
		EntryID: entry.ID,
	}); err != nil {
		log.Error().
			Err(err).
			Str("entry_id", entry.PublicID.String()).
			Msg("failed to update entry status")
		return seer.Wrap("update_entry_status_in_handle_entry", err)
	}

	// Include the URL (for links) and Minio URL (for attachments) in the entry
	var url string
	switch meta := entry.Metadata.(type) {
	case ograph.Metadata:
		url = meta.Link
	case models.FileMetadata:
		u, err := h.objectStore.GetPresignedUrl(ctx, entry.FileID)
		if err != nil {
			log.Error().
				Err(err).
				Str("entry_id", entry.PublicID.String()).
				Msg("failed to get presigned URL")
			return seer.Wrap("get_presigned_url_in_handle_entry", err)
		}
		url = u.String()
	}

	pluginArgs := host.OnCreateArgs{
		Entry: &entry,
		URL:   url,
	}

	// Run plugins OnCreate method
	succeeded := false
	for i := range installedPlugins {
		installedPlugin := &installedPlugins[i]
		plugin, err := h.wasmRuntime.LoadPlugin(ctx, installedPlugin)
		if err != nil {
			log.Error().
				Err(err).
				Str("plugin_id", installedPlugin.PluginIdentifier).
				Str("plugin_name", installedPlugin.Name()).
				Msg("failed to load plugin")
			continue
		}

		if err := plugin.OnCreate(ctx, &pluginArgs); err != nil {
			// NOTE: we will handle failure later
			log.Error().
				Err(err).
				Str("plugin_id", installedPlugin.PluginIdentifier).
				Str("plugin_name", installedPlugin.Name()).
				Str("entry_id", entry.PublicID.String()).
				Msg("failed to run on_create hook")
			continue
		}

		// If at least one plugin succeeded, we can update the flag
		succeeded = true
	}

	// Update the entry's status
	status := queries.EntryStatusFailed
	if succeeded {
		status = queries.EntryStatusCompleted
	}

	if err := h.repos.EntryRepository().UpdateQueue(&repository.UpdateQueueArgs{
		Status:  status,
		EntryID: entry.ID,
	}); err != nil {
		log.Error().
			Err(err).
			Str("entry_id", entry.PublicID.String()).
			Msg("failed to update entry status")
	}

	return nil
}

func (h *handler) HandleChunkEmbedding(ctx context.Context, message core.TaskMessage) error {
	payload := new(job.ChunkEmbeddingJob)
	if err := json.Unmarshal(message.Payload(), payload); err != nil {
		return err
	}

	if payload.ID == 0 {
		return errors.New("chunk ID is 0")
	}
	log.Info().Int32("chunk_id", payload.ID).Msg("processing chunk embedding job")

	// Check if chunk can be processed or skip it
	canProcess, err := h.repos.EntryRepository().CanEmbedChunk(payload.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check if chunk can be processed")
		return seer.Wrap("check_if_chunk_can_be_processed", err)
	}
	if !canProcess {
		return nil
	}

	// Mark the chunk as processing
	if err := h.repos.EntryRepository().UpdateSemanticVectorState(&repository.UpdateChunkSemanticVectorArgs{
		ChunkID: payload.ID,
		Vector:  []float32{},
		Status:  queries.EntryChunkEmbeddingStatusProcessing,
		Error:   nil,
	}); err != nil {
		log.Error().Err(err).Msg("failed to update chunk embedding status")
		return seer.Wrap("update_chunk_embedding_status_in_queue", err)
	}

	embeddings, err := h.llm.GenerateEmbedding(ctx, payload.Content)
	if err != nil {
		return seer.Wrap(
			"generate_embedding_in_queue",
			fmt.Errorf("failed to generate embedding: %w", err),
		)
	}

	if embeddings == nil {
		//nolint:exhaustruct
		if emErr := h.repos.EntryRepository().UpdateSemanticVectorState(&repository.UpdateChunkSemanticVectorArgs{
			ChunkID: payload.ID,
			Status:  queries.EntryChunkEmbeddingStatusFailed,
			Error:   err,
		}); emErr != nil {
			return seer.Wrap("update_chunk_embedding_status_in_queue", emErr)
		}

		return err
	}

	if err = h.repos.EntryRepository().
		UpdateSemanticVectorState(&repository.UpdateChunkSemanticVectorArgs{
			ChunkID: payload.ID,
			Vector:  embeddings,
			Status:  queries.EntryChunkEmbeddingStatusDone,
			Error:   err,
		}); err != nil {
		log.Error().Err(err).Msg("failed to update chunk embedding status")
		return seer.Wrap("update_chunk_embedding_status_in_queue", err)
	}

	log.Info().Int32("chunk_id", payload.ID).Msg("chunk embedding job completed")
	return nil
}
