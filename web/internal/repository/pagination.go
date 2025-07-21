package repository

import (
	"math"

	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
)

const (
	DefaultPage    int32 = 1
	DefaultPerPage int32 = 50
)

var (
	ErrInvalidPage     = apperrors.BadRequest("page must be greater than 0")
	ErrInvalidPageSize = apperrors.BadRequest("page size must be greater than 0")
)

var DefaultPaginationParams = PaginationParams{
	Page:    DefaultPage,
	PerPage: DefaultPerPage,
}

type (
	PaginationParams struct {
		Page    int32 `json:"page"     validate:"min=1"`
		PerPage int32 `json:"per_page" validate:"min=5,max=100"`
	}

	PaginationState struct {
		NextPage     *int32 `json:"next_page"`
		PreviousPage *int32 `json:"previous_page"`
		CurrentPage  int32  `json:"current_page"`
		TotalPages   int64  `json:"total_pages"`
		TotalCount   int64  `json:"total_count"`
	}
)

func (p PaginationParams) Validate() error {
	if p.Page < 1 {
		return ErrInvalidPage
	}

	if p.PerPage < 1 {
		return ErrInvalidPageSize
	}

	return nil
}

func (p PaginationParams) Offset() int32 {
	perPage := DefaultPerPage
	if p.PerPage > 0 {
		perPage = p.PerPage
	}

	page := DefaultPage
	if p.Page > 0 {
		page = p.Page
	}

	return (page - 1) * perPage
}

func (p PaginationParams) Limit() int32 {
	if p.PerPage < 1 {
		return DefaultPerPage
	}

	return p.PerPage + 1
}

type PageStateArgs struct {
	CurrentCount int
	TotalCount   int64
}

func (p PaginationParams) ToState(args PageStateArgs) PaginationState {
	var prevPage, nextPage *int32
	if p.Page > 1 {
		prevPage = lib.Ref(p.Page - 1)
	}

	if args.CurrentCount > 0 && args.CurrentCount > int(p.PerPage) {
		nextPage = lib.Ref(p.Page + 1)
	} else if args.TotalCount > int64(p.Page*p.PerPage) {
		nextPage = lib.Ref(p.Page + 1)
	}

	return PaginationState{
		NextPage:     nextPage,
		PreviousPage: prevPage,
		CurrentPage:  p.Page,
		TotalPages:   int64(math.Ceil(float64(args.TotalCount) / float64(p.Limit()))),
		TotalCount:   args.TotalCount,
	}
}
