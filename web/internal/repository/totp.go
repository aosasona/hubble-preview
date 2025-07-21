package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/seer"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

type TOTPRepository interface {
	// FindSecretsWithMinVersion is a method to retrieve all TOTP secrets with the minimum version
	FindOutdatedSecrets(maxVersion int16) ([]models.TotpSecret, error)

	// FindSecretByAccountID retrieves a TOTP secret by account ID
	FindSecretByAccountID(accountID pgtype.UUID) (*models.TotpSecret, error)

	// CreateSecret is a method to create a TOTP secret
	CreateSecret(secret *models.TotpSecret) (queries.TotpSecret, error)

	// UpdateSecret is a method to update a TOTP secret
	UpdateSecret(secret *models.TotpSecret) error

	// BatchUpdateSecrets is a method to update a chunk of TOTP secrets
	BatchUpdateSecrets(secrets []*models.TotpSecret) error

	// DeleteSecret is a method to delete a TOTP secret
	DeleteSecret(accountId pgtype.UUID, secretId int) error
}

type (
	totpRepo struct {
		*baseRepo
	}
)

func (r *totpRepo) FindOutdatedSecrets(version int16) ([]models.TotpSecret, error) {
	if version == 0 {
		return nil, seer.Wrap(
			"find_outdated_totp_secrets",
			errors.New("invalid version, expected integer greater than 0"),
		)
	}

	secrets, err := r.queries.FindOutdatedTotpHashes(context.TODO(), int32(version))
	if err != nil {
		return nil, err
	}

	var totpSecrets []models.TotpSecret
	for i := range secrets {
		secret := models.TotpSecret{
			ID:        secrets[i].ID,
			AccountID: secrets[i].AccountID,
			Hash:      secrets[i].Hash,
			Version:   secrets[i].Version,
		}

		totpSecrets = append(totpSecrets, secret)
	}

	return totpSecrets, nil
}

func (r *totpRepo) FindSecretByAccountID(accountID pgtype.UUID) (*models.TotpSecret, error) {
	if !accountID.Valid {
		return nil, seer.Wrap(
			"find_totp_secret_by_account_id",
			errors.New("account_id is required"),
		)
	}

	secret, err := r.queries.FindTotpSecretByAccountId(context.TODO(), accountID)
	if err != nil {
		return nil, err
	}

	return &models.TotpSecret{
		ID:        secret.ID,
		AccountID: secret.AccountID,
		Hash:      secret.Hash,
		Version:   secret.Version,
	}, nil
}

func (r *totpRepo) CreateSecret(secret *models.TotpSecret) (queries.TotpSecret, error) {
	if !secret.AccountID.Valid {
		return queries.TotpSecret{}, seer.Wrap(
			"create_totp_secret",
			errors.New("account_id is required"),
		)
	}

	if secret.Version == 0 {
		secret.Version = 1
	}

	if len(secret.Hash) == 0 {
		return queries.TotpSecret{}, seer.Wrap("create_totp_secret", errors.New("hash is required"))
	}

	result, err := r.queries.CreateTotpSecret(context.TODO(), queries.CreateTotpSecretParams{
		AccountID: secret.AccountID,
		Hash:      secret.Hash,
		Version:   secret.Version,
	})
	if err != nil {
		return queries.TotpSecret{}, err
	}

	return result, nil
}

func (r *totpRepo) UpdateSecret(secret *models.TotpSecret) error {
	if secret.ID == 0 {
		return seer.Wrap("create_totp_secret", errors.New("id is required"))
	}

	if secret.Version == 0 {
		return seer.Wrap("create_totp_secret", errors.New("non-zero version is required"))
	}

	if len(secret.Hash) == 0 {
		return seer.Wrap("create_totp_secret", errors.New("hash is required"))
	}

	return r.queries.UpdateTotpHash(context.TODO(), queries.UpdateTotpHashParams{
		Hash:    secret.Hash,
		Version: secret.Version,
		ID:      secret.ID,
	})
}

func (r *totpRepo) BatchUpdateSecrets(secrets []*models.TotpSecret) error {
	tx, err := r.pool.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return seer.Wrap("create_db_pool", err)
	}

	eg := new(errgroup.Group)

	for _, secret := range secrets {
		eg.Go(func() error {
			err := r.queries.UpdateTotpHash(context.TODO(), queries.UpdateTotpHashParams{
				Hash:    secret.Hash,
				Version: secret.Version,
				ID:      secret.ID,
			})
			if err != nil {
				return seer.Wrap("update_totp_hash", err)
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		if err := tx.Rollback(context.Background()); err != nil {
			return seer.Wrap("rollback_batch_update_totp_hash", err)
		}

		return seer.Wrap("errgroup_update_totp_hash", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return seer.Wrap("commit_batch_update_totp_hash", err)
	}

	return nil
}

func (r *totpRepo) DeleteSecret(accountId pgtype.UUID, secretId int) error {
	return r.queries.DeleteTotpSecret(context.TODO(), queries.DeleteTotpSecretParams{
		ID:        int32(secretId),
		AccountID: accountId,
	})
}

var _ TOTPRepository = (*totpRepo)(nil)
