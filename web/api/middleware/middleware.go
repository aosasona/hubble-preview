package middleware

import (
	"github.com/adelowo/gulter"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/objectstore"
	"go.trulyao.dev/hubble/web/internal/ratelimit"
	"go.trulyao.dev/hubble/web/internal/repository"
	"go.trulyao.dev/robin"
)

type Middleware interface {
	WithAuth(ctx *robin.Context) error
	WithGulter(*gulter.Gulter, []string) func(ctx *robin.Context) error
	WithRateLimit(ctx *robin.Context) error
}

type middleware struct {
	repository  repository.Repository
	config      *config.Config
	rateLimiter ratelimit.RateLimiter
	objectStore *objectstore.Store
}

type Deps struct {
	Repository  repository.Repository
	Config      *config.Config
	RateLimiter ratelimit.RateLimiter
}

func New(deps *Deps) Middleware {
	return &middleware{
		repository:  deps.Repository,
		config:      deps.Config,
		rateLimiter: deps.RateLimiter,
	}
}
