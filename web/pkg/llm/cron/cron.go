// This cron is used to find chunks that need to be indexed
package llmcron

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/queue"
	"go.trulyao.dev/hubble/web/internal/repository"
)

const DefaultEmbeddingInterval = time.Minute * 30 // Default interval for embedding jobs

type Cron struct {
	config     *config.Config
	queue      *queue.Queue
	repository repository.Repository

	scheduler    gocron.Scheduler
	queuedChunks []int32 // We need to keep track of chunks we have queued already so we don't keep sending the same ones over and over
}

func NewCron(
	config *config.Config,
	repo repository.Repository,
	queue *queue.Queue,
) (*Cron, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	if repo == nil {
		return nil, errors.New("repository is nil")
	}

	if queue == nil {
		return nil, errors.New("queue is nil")
	}

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("failed to create scheduler: %w", err)
	}

	c := Cron{
		config:       config,
		repository:   repo,
		scheduler:    scheduler,
		queue:        queue,
		queuedChunks: []int32{},
	}

	return &c, nil
}

func (c *Cron) Start() error {
	if !c.config.LLM.EnabledEmbeddings() {
		log.Info().Str("source", "embedding_cron").Msg("embedding cron is disabled")
		return nil
	}

	_, err := c.scheduler.NewJob(
		gocron.DurationJob(DefaultEmbeddingInterval),
		gocron.NewTask(c.loadUnindexedChunks),
	)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	c.scheduler.Start()
	return nil
}

func (c *Cron) Stop() error {
	return c.scheduler.Shutdown()
}

func (c *Cron) loadUnindexedChunks() {
	chunks, err := c.repository.EntryRepository().FindUnindexedChunks()
	if err != nil {
		log.Error().Err(err).Str("source", "embedding_cron").Msg("failed to find unindexed chunks")
		return
	}

	if len(chunks) == 0 {
		log.Info().Str("source", "embedding_cron").Msg("no unindexed chunks found")
		return
	}

	for _, chunk := range chunks {
		if c.isChunkQueued(chunk.ID) {
			log.Info().Str("source", "embedding_cron").Msgf("chunk %d already queued", chunk.ID)
			continue
		}

		job := job.ChunkEmbeddingJob{ID: chunk.ID, Content: chunk.Content}
		if err := c.queue.Add(&job); err != nil {
			log.Error().
				Err(err).
				Str("source", "embedding_cron").
				Msgf("failed to queue chunk %d", chunk.ID)
			continue
		}
	}
}

func (c *Cron) isChunkQueued(id int32) bool {
	return slices.Contains(c.queuedChunks, id)
}
