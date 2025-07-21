package api

import (
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/robin"
)

type UserHandler interface {
	// Me returns a comprehensive view of the user's account (e.g. profile, workspaces, settings, etc.)
	Me(ctx *robin.Context, _ robin.Void) (MeResponse, error)

	SaveProfile(ctx *robin.Context, data SaveProfileRequest) (models.User, error)
}

type (
	MeResponse struct {
		User       models.User        `json:"user"`
		Workspaces []models.Workspace `json:"workspaces"`
	}

	SaveProfileRequest struct {
		FirstName string `json:"first_name" validate:"required,min=2,max=50,alpha"`
		LastName  string `json:"last_name"  validate:"required,min=2,max=50,alpha"`
		Username  string `json:"username"   validate:"required,username,min=2,max=50"`
	}
)
