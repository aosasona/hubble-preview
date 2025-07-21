package middleware

import (
	"context"
	"net/http"
	"sync"

	"github.com/adelowo/gulter"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/seer"
)

type GulterOptions struct {
	// Bucket is the bucket to store files in
	Bucket objectstore.Bucket
	// MaxFileSize is the maximum size of a file that can be uploaded
	MaxFileSize int64
	// AllowedMimeTypes is a list of allowed mime types, leave empty to allow all
	AllowedMimeTypes []string
	// ObjectStore is the object store to use
	ObjectStore *objectstore.Store
}

type gulterError struct {
	err error
	mu  sync.RWMutex
}

func (e *gulterError) Err() error {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return e.err
}

func (e *gulterError) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.err = nil
}

func (e *gulterError) Set(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.err = err
}

var sharedGulterError = &gulterError{}

func CreateGulterInstance(opts *GulterOptions) (*gulter.Gulter, error) {
	maxFileSize := opts.MaxFileSize
	if maxFileSize == 0 {
		maxFileSize = 128 * 1024 * 1024 // 128MB
	}

	validators := []gulter.ValidationFunc{}
	if len(opts.AllowedMimeTypes) > 0 {
		validators = append(validators, gulter.MimeTypeValidator(opts.AllowedMimeTypes...))
	}

	instance, err := gulter.New(
		gulter.WithMaxFileSize(maxFileSize),
		gulter.WithIgnoreNonExistentKey(true),
		gulter.WithValidationFunc(gulter.ChainValidators(validators...)),
		gulter.WithNameFuncGenerator(func(_ string) string {
			return uuid.NewString()
		}),
		gulter.WithStorage(opts.ObjectStore),
		gulter.WithErrorResponseHandler(func(err error) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// NOTE: don't write status header or anything at all, robin will handle it
				sharedGulterError.Set(err)
			}
		}),
	)
	if err != nil {
		return nil, err
	}

	// Attempt to create bucket if it doesn't exist
	exists, err := opts.ObjectStore.Client().BucketExists(context.TODO(), opts.Bucket.String())
	if err != nil {
		return nil, seer.Wrap("bucket_exists", err)
	}

	if !exists {
		err = opts.ObjectStore.Client().
			MakeBucket(context.TODO(), opts.Bucket.String(), minio.MakeBucketOptions{})
		if err != nil {
			return nil, seer.Wrap("make_bucket", err)
		}
	}

	return instance, nil
}

func (m *middleware) WithGulter(
	instance *gulter.Gulter,
	keys []string,
) func(ctx *robin.Context) error {
	return func(ctx *robin.Context) error {
		sharedGulterError.Reset()

		instance.Upload(keys...)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				corsOpts := &robin.CorsOptions{
					Origins:          []string{"http://localhost:5173"},
					AllowCredentials: true,
					Methods:          []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					Headers:          []string{"Content-Type", "Authorization", "Origin"},
				}

				if m.config.InDevelopment() {
					robin.CorsHandler(w, corsOpts)
				}

				if m.config.Debug() {
					log.Debug().Msg("gulter handler called")
				}

				files, err := gulter.FilesFromContext(r)
				if err != nil {
					log.Error().Err(err).Msg("failed to get files from context")
					return
				}

				if len(files) == 0 {
					return
				}

				ctx.Set("files", files)
			}),
		).ServeHTTP(ctx.Response(), ctx.Request())

		return sharedGulterError.Err()
	}
}
