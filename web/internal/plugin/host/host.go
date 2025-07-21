package host

import (
	"context"

	"go.trulyao.dev/hubble/web/internal/job"
	"go.trulyao.dev/hubble/web/internal/plugin/host/boundhost"
	"go.trulyao.dev/hubble/web/internal/plugin/spec"
	"go.trulyao.dev/hubble/web/internal/repository"
)

type Host struct {
	repository repository.Repository
	queueFn    job.QueueFn
}

func NewHost(repo repository.Repository, queueFn job.QueueFn) *Host {
	return &Host{repository: repo, queueFn: queueFn}
}

func (h *Host) Bind(
	ctx context.Context,
	identifier string,
	privileges []spec.Privilege,
) *boundhost.BoundHost {
	return boundhost.New(ctx, h.repository, identifier, privileges, h.queueFn)
}
