package api

import (
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/robin"
	"golang.org/x/sync/errgroup"
)

type userHandler struct {
	*baseHandler
}

// Me implements UserHandler.
func (u *userHandler) Me(ctx *robin.Context, _ robin.Void) (MeResponse, error) {
	var response MeResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		user, err := authlib.ExtractUser(ctx, u.repos.UserRepository())
		if err != nil {
			return err
		}
		response.User = *user
		return nil
	})

	eg.Go(func() error {
		workspaces, err := u.repos.WorkspaceRepository().FindAllByUserID(auth.UserID)
		if err != nil {
			return err
		}

		response.Workspaces = workspaces
		return nil
	})

	if err := eg.Wait(); err != nil {
		return response, err
	}

	return response, nil
}

func (u *userHandler) SaveProfile(
	ctx *robin.Context,
	data SaveProfileRequest,
) (models.User, error) {
	var (
		user *models.User
		err  error
	)

	if err := lib.ValidateStruct(&data); err != nil {
		return models.User{}, err
	}

	if user, err = authlib.ExtractUser(ctx, u.repos.UserRepository()); err != nil {
		return models.User{}, err
	}

	// Ensure the username does not already exists
	usernameExists, err := u.repos.UserRepository().UsernameExists(data.Username)
	if err != nil {
		return models.User{}, err
	}

	if usernameExists && user.Username != data.Username {
		return models.User{}, apperrors.NewValidationError(apperrors.ErrorMap{
			"username": []string{"Username is already taken"},
		})
	}

	updatedUser, err := u.repos.UserRepository().UpdateProfile(
		user.ID,
		repository.UpdateProfileParams{
			FirstName: data.FirstName,
			LastName:  data.LastName,
			Username:  data.Username,
		},
	)
	if err != nil {
		return models.User{}, err
	}

	return *updatedUser, nil
}

var _ UserHandler = (*userHandler)(nil)
