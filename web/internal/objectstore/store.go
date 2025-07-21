package objectstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/adelowo/gulter"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.trulyao.dev/seer"
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(entries)
type Bucket string

const (
	DefaultPresignedURLExpiration = time.Minute * 5
)

type Store struct {
	endpoint string
	client   *minio.Client
	bucket   Bucket
}

type MinioOptions struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	IsDev     bool
}

var ErrInvalidEndpoint = errors.New("invalid endpoint")

func NewMinioStore(opts MinioOptions) (*Store, error) {
	if opts.Endpoint == "" {
		return nil, ErrInvalidEndpoint
	}

	client, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: !opts.IsDev,
	})
	if err != nil {
		return nil, seer.Wrap("create_client", err)
	}

	return &Store{
		endpoint: opts.Endpoint,
		client:   client,
	}, nil
}

// WithBucket returns a new Store with the bucket set
func (m *Store) WithBucket(b Bucket) *Store {
	s := *m
	s.bucket = b
	return &s
}

func (m *Store) Client() *minio.Client {
	return m.client
}

func (s *Store) Upload(
	ctx context.Context,
	r io.Reader,
	opts *gulter.UploadFileOptions,
) (*gulter.UploadedFileMetadata, error) {
	if s.bucket == "" {
		return nil, seer.Wrap("no_bucket", errors.New("no bucket set"))
	}

	b := new(bytes.Buffer)

	r = io.TeeReader(r, b)

	n, err := io.Copy(io.Discard, r)
	if err != nil {
		return nil, err
	}

	seeker, err := gulter.ReaderToSeeker(b)
	if err != nil {
		return nil, err
	}

	_, err = s.client.PutObject(
		ctx,
		s.bucket.String(),
		opts.FileName,
		seeker,
		n,
		minio.PutObjectOptions{
			UserMetadata:         opts.Metadata,
			AutoChecksum:         minio.ChecksumCRC32C,
			SendContentMd5:       false,
			DisableContentSha256: true,
		},
	)
	if err != nil {
		return nil, err
	}

	return &gulter.UploadedFileMetadata{
		FolderDestination: s.bucket.String(),
		Size:              n,
		Key:               opts.FileName,
	}, nil
}

func (s *Store) Path(ctx context.Context, opts gulter.PathOptions) (string, error) {
	if s.bucket == "" {
		return "", seer.Wrap("no_bucket", errors.New("no bucket set"))
	}

	if !opts.IsSecure {
		url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, opts.Key)
		return url, nil
	}

	presignedUrl, err := s.client.PresignedGetObject(
		ctx,
		s.bucket.String(),
		opts.Key,
		opts.ExpirationTime,
		url.Values{},
	)
	if err != nil {
		return "", err
	}

	return presignedUrl.String(), nil
}

func (s *Store) GetPresignedUrl(ctx context.Context, fileId string) (*url.URL, error) {
	urlParams := url.Values{}
	url, err := s.client.PresignedGetObject(
		ctx,
		BucketEntries.String(),
		fileId,
		DefaultPresignedURLExpiration,
		urlParams,
	)
	if err != nil {
		return nil, seer.Wrap("get_presigned_url", err)
	}

	return url, nil
}

func (s *Store) Close() error {
	return nil
}

var _ gulter.Storage = (*Store)(nil)
