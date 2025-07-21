package auth

import (
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/robin"
)

func ExtractAuthSession(ctx *robin.Context) (models.AuthSession, error) {
	auth, ok := ctx.Get(StateKeySession).(models.AuthSession)
	if !ok {
		return models.AuthSession{}, apperrors.ErrIncompleteSession
	}

	return auth, nil
}

func ExtractUser(ctx *robin.Context, repo repository.UserRepository) (*models.User, error) {
	session, err := ExtractAuthSession(ctx)
	if err != nil {
		return &models.User{}, err
	}

	user, err := repo.FindUserByID(session.UserID)
	if err != nil {
		return &models.User{}, err
	}

	return user, nil
}
