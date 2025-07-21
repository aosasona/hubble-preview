package queue

import (
	"errors"
	"time"

	"github.com/golang-queue/queue"
	"github.com/golang-queue/queue/job"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	appjob "go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/plugin/host"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/hubble/web/pkg/llm"
	"go.trulyao.dev/seer"
)

var ErrUnsupportedJobType = errors.New("unsupported job type")

const DefaultRetryInterval = 2 * 60 // 2 minutes

const (
	DefaultEntryQueueSize     = 15
	DefaultEmbeddingQueueSize = 10
)

const (
	DefaultChunkEmbeddingDuration  = 2 * time.Minute // Chunk embedding jobs are allowed to run for this long
	DefaultEntryProcessingDuration = 5 * time.Minute // Entry processing jobs are allowed to run for this long
)

type Queue struct {
	config      *config.Config
	repos       repository.Repository
	wasmRuntime *host.Runtime
	handler     *handler
	objectStore *objectstore.Store
	llm         *llm.LLM

	entries        *queue.Queue
	chunkEmbedding *queue.Queue
}

// New creates a new queue instance
func New(
	config *config.Config,
	repos repository.Repository,
	objectStore *objectstore.Store,
	wasmRuntime *host.Runtime,
	llm *llm.LLM,
) *Queue {
	handler := NewHandler(config, repos, objectStore, wasmRuntime, llm)
	return &Queue{
		repos:       repos,
		config:      config,
		handler:     handler,
		wasmRuntime: wasmRuntime,
		objectStore: objectStore,
		llm:         llm,
		entries: queue.NewPool(
			DefaultEntryQueueSize,
			queue.WithRetryInterval(DefaultRetryInterval),
			queue.WithFn(handler.HandleEntry),
			queue.WithLogger(&logger{}),
		),
		chunkEmbedding: queue.NewPool(
			DefaultEmbeddingQueueSize,
			queue.WithRetryInterval(DefaultRetryInterval),
			queue.WithFn(handler.HandleChunkEmbedding),
			queue.WithLogger(&logger{}),
		),
	}
}

func (q *Queue) Start() error {
	defer func() {
		if err := recover(); err != nil {
			q.entries.Release()
		}
	}()

	q.entries.Start()

	if !q.config.LLM.EnabledEmbeddings() {
		q.chunkEmbedding.Start()
	}

	return nil
}

func (q *Queue) Close() error {
	q.entries.Release()
	if !q.config.LLM.EnabledEmbeddings() {
		q.chunkEmbedding.Release()
	}
	return nil
}

// Load loads existing jobs from the database into the appropriate queue
func (q *Queue) Load() error {
	if q.llm == nil {
		log.Error().Msg("llm is nil")
		return errors.New("llm is nil")
	}

	// Load queued entries from the database
	queuedEntryIds, err := q.repos.EntryRepository().FindQueuedIDs()
	if err != nil {
		return err
	}

	for _, id := range queuedEntryIds {
		job := &appjob.EntryJob{ID: id}
		if err := q.Add(job); err != nil {
			return seer.Wrap("load_job_into_queue", err)
		}
	}

	// Load chunk embedding jobs
	chunks, err := q.repos.EntryRepository().FindUnindexedChunks()
	if err != nil {
		return err
	}

	for _, chunk := range chunks {
		job := &appjob.ChunkEmbeddingJob{
			ID:      chunk.ID,
			Content: chunk.Content,
		}

		if err := q.Add(job); err != nil {
			return seer.Wrap("load_job_into_queue", err)
		}
	}

	return nil
}

// Add adds a new job to the appropriate underlying queue
func (q *Queue) Add(payload appjob.Job) error {
	switch payload := payload.(type) {
	case *appjob.EntryJob:
		//nolint:exhaustruct
		return q.entries.Queue(payload, job.AllowOption{
			RetryDelay: job.Time(DefaultRetryInterval),
			RetryMin:   job.Time(time.Minute * 5),
			RetryMax:   job.Time(time.Minute * 20),
			Timeout:    job.Time(DefaultEntryProcessingDuration),
		})

	case *appjob.ChunkEmbeddingJob:
		if !q.config.LLM.EnabledEmbeddings() {
			return nil
		}

		//nolint:exhaustruct
		return q.chunkEmbedding.Queue(payload, job.AllowOption{
			RetryDelay: job.Time(DefaultRetryInterval),
			RetryMin:   job.Time(time.Minute * 2),
			RetryMax:   job.Time(time.Minute * 10),
			Timeout:    job.Time(DefaultChunkEmbeddingDuration),
		})

	case *appjob.EntryChunkEmbeddingJob:
		if !q.config.LLM.EnabledEmbeddings() {
			return nil
		}

		log.Info().Str("source", "queue").Msg("received entries chunk embedding job")
		ids := lib.UniqueSlice(payload.Entries)
		if len(ids) == 0 {
			return nil
		}

		for _, id := range ids {
			chunks, err := q.repos.EntryRepository().FindUnindexedEntryChunks(id)
			if err != nil {
				log.Error().
					Err(err).
					Str("source", "queue").
					Str("job", "EntryChunkEmbeddingJob").
					Msg("failed to find unindexed chunks")

				return seer.Wrap("find_unindexed_chunks", err)
			}

			if len(chunks) == 0 {
				continue
			}

			for _, chunk := range chunks {
				job := &appjob.ChunkEmbeddingJob{
					ID:      chunk.ID,
					Content: chunk.Content,
				}

				if err := q.Add(job); err != nil {
					log.Error().
						Err(err).
						Str("source", "queue").
						Str("job", "EntryChunkEmbeddingJob").
						Msg("failed to queue chunk embedding job")

					continue
				}
			}
		}

		return nil

	default:
		return ErrUnsupportedJobType
	}
}

// AddMany adds multiple jobs to the queue
func (q *Queue) AddMany(jobs ...appjob.Job) error {
	for _, j := range jobs {
		if err := q.Add(j); err != nil {
			return err
		}
	}

	return nil
}
