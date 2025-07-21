package repository

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/kv"
	"go.trulyao.dev/hubble/web/internal/models"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	"go.trulyao.dev/seer"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

var ErrMfaAccountNotFound = apperrors.New("account not found", http.StatusNotFound)

type MfaRepository interface {
	// IsEnabled checks if MFA is enabled for a user
	IsEnabled(userID int32) (bool, error)

	// LoadState retrieves the MFA settings for a user
	LoadState(userID int32) (models.MfaState, error)

	// FindAccountByEmail finds an MFA account by email
	FindAccountByEmail(email string) (*models.MfaAccount, error)

	// FindAccountById finds an MFA account by its ID
	FindAccountById(accountId pgtype.UUID) (*models.MfaAccount, error)

	// CreateEmailAccount creates an email account for MFA but does not activate it
	CreateEmailAccount(userId int32, email string) (*models.MfaAccount, error)

	// CreateTotpAccount creates an active TOTP account for MFA
	CreateTotpAccount(userId int32, name string) (*models.MfaAccount, error)

	// ActivateAccount activates an MFA account for a user
	ActivateAccount(accountId pgtype.UUID, userId int32) error

	// DeleteAccount deletes an MFA account and updates the preferred account if necessary
	DeleteAccount(accountId pgtype.UUID, userId int32) error

	// SetPreferredAccount sets a preferred MFA account for a user
	SetPreferredAccount(accountId pgtype.UUID, userId int32) (*models.MfaAccount, error)

	// NameIsTaken checks if an MFA account name is already in use by the user
	NameIsTaken(userId int32, name string) (bool, error)

	// RenameAccount renames an MFA account
	RenameAccount(params RenameAccountParams) error

	// SetLastUsed updates the last used timestamp for an MFA account
	SetLastUsed(accountId pgtype.UUID) error

	// CreateSession creates a new MFA session for a user
	CreateSession(userId int32, accountId pgtype.UUID) (models.MfaSession, error)

	// FindSession finds an MFA session by its ID
	FindSession(sessionId string) (models.MfaSession, error)

	// DeleteSession deletes an MFA session
	DeleteSession(sessionId string) error

	// CreateMfaEnrollmentSession creates a new MFA enrollment session for a user
	CreateTotpEnrollmentSession(
		params *CreateTotpEnrollmentSessionParams,
	) (models.MfaSession, error)

	// FindTotpEnrollmentSession finds a TOTP enrollment session by its ID
	FindTotpEnrollmentSession(sessionId string) (models.MfaSession, error)

	// GenerateBackupCodes generates backup codes for a user to use in case they lose access to their MFA device
	GenerateBackupCodes(userId int32) ([]string, error)

	// VerifyBackupCode verifies a backup code for a user and deletes it if it is valid
	VerifyBackupCode(userId int32, code string) error

	// CanGenerateBackupCodes checks if a user can generate backup codes
	CanGenerateBackupCodes(userId int32) (bool, error)
}

const (
	// MfaSessionDuration is the duration for which an MFA session is valid for after which the user is forced to restart the process
	MfaSessionDuration = time.Minute * 25
)

type mfaRepo struct {
	*baseRepo
}

type (
	RenameAccountParams struct {
		AccountId pgtype.UUID
		UserId    int32
		Name      string
	}

	CreateTotpEnrollmentSessionParams struct {
		User        *models.User
		AccountName string

		// HostUrl is also known as the issuer
		HostUrl string
	}
)

func (m *mfaRepo) IsEnabled(userID int32) (bool, error) {
	return m.queries.MfaEnabled(context.TODO(), userID)
}

func (m *mfaRepo) LoadState(userID int32) (models.MfaState, error) {
	var settings models.MfaState

	enabled, err := m.queries.MfaEnabled(context.TODO(), userID)
	if err != nil {
		return settings, err
	}
	settings.Enabled = enabled

	accounts, err := m.queries.FindActiveMfaAccountsByUserId(context.TODO(), userID)
	if err != nil {
		log.Error().Err(err).Int32("user_id", userID).Msg("failed to find active MFA accounts")
		return settings, err
	}

	for _, account := range accounts {
		settings.Accounts = append(settings.Accounts, models.MfaAccount{
			ID:           account.ID,
			Type:         account.AccountType,
			Meta:         json.RawMessage(account.Meta),
			Active:       account.Active,
			UserID:       account.UserID,
			RegisteredAt: account.CreatedAt,
			LastUsedAt:   account.LastUsedAt,
			Preferred:    account.Preferred,
		})

		if account.Preferred {
			settings.PreferredAccountID = account.ID
		}
	}

	return settings, nil
}

func (m *mfaRepo) NameIsTaken(userId int32, name string) (bool, error) {
	return m.queries.MfaAccountNameExists(context.TODO(), queries.MfaAccountNameExistsParams{
		Name:   name,
		UserID: userId,
	})
}

func (m *mfaRepo) FindAccountByEmail(email string) (*models.MfaAccount, error) {
	account, err := m.queries.FindMfaAccountByEmail(context.TODO(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &models.MfaAccount{}, ErrMfaAccountNotFound
		}

		return &models.MfaAccount{}, err
	}

	return &models.MfaAccount{
		ID:           account.ID,
		Type:         account.AccountType,
		Meta:         json.RawMessage(account.Meta),
		Active:       account.Active,
		UserID:       account.UserID,
		RegisteredAt: account.CreatedAt,
		LastUsedAt:   account.LastUsedAt,
		Preferred:    account.Preferred,
	}, nil
}

func (m *mfaRepo) FindAccountById(accountId pgtype.UUID) (*models.MfaAccount, error) {
	account, err := m.queries.FindMfaAccountById(context.TODO(), accountId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &models.MfaAccount{}, ErrMfaAccountNotFound
		}

		return &models.MfaAccount{}, err
	}

	return &models.MfaAccount{
		ID:           account.ID,
		Type:         account.AccountType,
		Meta:         json.RawMessage(account.Meta),
		Active:       account.Active,
		UserID:       account.UserID,
		RegisteredAt: account.CreatedAt,
		LastUsedAt:   account.LastUsedAt,
		Preferred:    account.Preferred,
	}, nil
}

func (m *mfaRepo) DeleteAccount(accountId pgtype.UUID, userId int32) error {
	// Find the account itself
	account, err := m.queries.FindMfaAccountById(context.TODO(), accountId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMfaAccountNotFound
		}

		return err
	}

	// Ensure the user has permission to delete the account
	if account.UserID != userId {
		return apperrors.NewAuthorizationError(
			"you do not have permission to delete this account",
			http.StatusForbidden,
		)
	}

	// NOTE: Transaction to delete the account and update the preferred account if necessary
	tx, err := m.pool.BeginTx(context.TODO(), pgx.TxOptions{})
	if err != nil {
		return err
	}

	txQueries := m.queries.WithTx(tx)
	eg := new(errgroup.Group)

	// Handle the preferred account
	if account.Preferred {
		// Find all active accounts for the user
		accountIds, err := m.queries.FindActiveMfaAccountIdsByUserId(context.TODO(), userId)
		if err != nil {
			log.Error().
				Err(err).
				Int32("user_id", userId).
				Msg("failed to find active MFA account IDs")

			return err
		}

		eg.Go(func() error {
			// If there is only one account, we can just delete it
			if len(accountIds) == 1 {
				return nil
			}

			// Find the position of the account being deleted and set the next account as preferred
			var nextAccountId pgtype.UUID
			for i, id := range accountIds {
				if id == accountId {
					// If the account being deleted is the last one, set the previous account as preferred
					if i == len(accountIds)-1 {
						nextAccountId = accountIds[i-1]
					} else {
						nextAccountId = accountIds[i+1]
					}
					break
				}
			}

			return txQueries.SetPreferredMfaAccount(
				context.TODO(),
				queries.SetPreferredMfaAccountParams{
					ID:     nextAccountId,
					UserID: userId,
				},
			)
		})
	}

	// Delete the account
	eg.Go(func() error {
		return txQueries.DeleteMfaAccount(context.TODO(), queries.DeleteMfaAccountParams{
			ID:     accountId,
			UserID: userId,
		})
	})

	// Wait for all operations to complete
	if err := eg.Wait(); err != nil {
		if err := tx.Rollback(context.Background()); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				log.Error().Err(err).Msg("failed to rollback transaction")
			}
		}

		return err
	}

	// Commit the transaction
	if err := tx.Commit(context.Background()); err != nil {
		if errors.Is(err, pgx.ErrTxClosed) {
			log.Warn().Msg("transaction was already closed")
		}

		return err
	}

	return nil
}

func (m *mfaRepo) SetPreferredAccount(
	accountId pgtype.UUID,
	userId int32,
) (*models.MfaAccount, error) {
	account, err := m.FindAccountById(accountId)
	if err != nil {
		return nil, err
	}

	if account.UserID != userId {
		return account, apperrors.New(
			"insufficient permissions to perform this action",
			http.StatusForbidden,
		)
	}

	if err := m.queries.SetPreferredMfaAccount(context.TODO(),
		queries.SetPreferredMfaAccountParams{
			ID:     accountId,
			UserID: userId,
		},
	); err != nil {
		return nil, err
	}

	account.Preferred = true
	return account, nil
}

func (m *mfaRepo) RenameAccount(params RenameAccountParams) error {
	exists, err := m.queries.MfaAccountNameExists(
		context.TODO(),
		queries.MfaAccountNameExistsParams{
			Name:   params.Name,
			UserID: params.UserId,
		},
	)
	if err != nil {
		log.Error().
			Err(err).
			Int32("user_id", params.UserId).
			Str("name", params.Name).
			Msg("failed to check if MFA account name exists")
		return err
	}

	if exists {
		return apperrors.NewValidationError(apperrors.ErrorMap{
			"name": []string{"already in use by another account"},
		})
	}

	_, err = m.queries.RenameMfaAccount(context.TODO(), queries.RenameMfaAccountParams{
		NewName:   strings.TrimSpace(params.Name),
		AccountID: params.AccountId,
		UserID:    params.UserId,
	})

	return err
}

func (m *mfaRepo) SetLastUsed(accountId pgtype.UUID) error {
	return m.queries.SetMfaAccountLastUsed(
		context.TODO(),
		accountId,
	)
}

func (m *mfaRepo) ActivateAccount(accountId pgtype.UUID, userId int32) error {
	return m.queries.ActivateMfaAccount(context.TODO(), queries.ActivateMfaAccountParams{
		ID:     accountId,
		UserID: userId,
	})
}

func (m *mfaRepo) CreateEmailAccount(userId int32, email string) (*models.MfaAccount, error) {
	emailMeta := models.EmailMeta{Email: email}
	meta, err := json.Marshal(emailMeta)
	if err != nil {
		log.Error().Err(err).Str("email", email).Msg("failed to marshal email MFA meta")
		return &models.MfaAccount{}, err
	}

	account, err := m.queries.CreateEmailMfaAccount(
		context.TODO(),
		queries.CreateEmailMfaAccountParams{
			UserID: userId,
			Meta:   meta,
		},
	)
	if err != nil {
		return &models.MfaAccount{}, err
	}

	return &models.MfaAccount{
		ID:           account.ID,
		Type:         account.AccountType,
		Meta:         json.RawMessage(account.Meta),
		Active:       account.Active,
		UserID:       account.UserID,
		RegisteredAt: account.CreatedAt,
		LastUsedAt:   account.LastUsedAt,
		Preferred:    account.Preferred,
	}, nil
}

func (m *mfaRepo) CreateTotpAccount(userId int32, name string) (*models.MfaAccount, error) {
	TotpMeta := models.TotpMeta{Name: name}
	meta, err := json.Marshal(TotpMeta)
	if err != nil {
		log.Error().Err(err).Str("name", name).Msg("failed to marshal TOTP MFA meta")
		return &models.MfaAccount{}, err
	}

	account, err := m.queries.CreateTotpMfaAccount(
		context.TODO(),
		queries.CreateTotpMfaAccountParams{
			UserID: userId,
			Meta:   meta,
		},
	)
	if err != nil {
		return &models.MfaAccount{}, err
	}

	return &models.MfaAccount{
		ID:           account.ID,
		Type:         account.AccountType,
		Meta:         json.RawMessage(account.Meta),
		Active:       account.Active,
		UserID:       account.UserID,
		RegisteredAt: account.CreatedAt,
		LastUsedAt:   account.LastUsedAt,
		Preferred:    account.Preferred,
	}, nil
}

func (m *mfaRepo) CreateSession(userId int32, accountId pgtype.UUID) (models.MfaSession, error) {
	account, err := m.queries.FindMfaAccountById(context.TODO(), accountId)
	if err != nil {
		return models.MfaSession{}, err
	}

	sessionType := models.FromAccountType(account.AccountType)
	meta, err := models.DecodeAccountMeta(sessionType, []byte(account.Meta))
	if err != nil {
		return models.MfaSession{}, err
	}

	sessionId := uuid.New()
	session := models.MfaSession{
		ID:          sessionId.String(),
		AccountID:   accountId,
		UserID:      userId,
		Meta:        meta,
		AccountType: account.AccountType,
		SessionType: sessionType,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(MfaSessionDuration),
	}

	if err := m.store.SetJsonWithTTL(
		kv.KeyMfaSession(sessionId.String()),
		session,
		MfaSessionDuration,
	); err != nil {
		return models.MfaSession{}, err
	}

	return session, nil
}

func (m *mfaRepo) FindSession(sessionId string) (models.MfaSession, error) {
	var session models.MfaSession

	// Check if the session has expired
	expired, err := m.store.ExpiresAt(kv.KeyMfaSession(sessionId))
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return session, apperrors.New(
				"invalid or expired session, please start the process again",
				http.StatusNotFound,
			)
		}

		return session, err
	}

	// If the token has expired, return false
	if expired.Before(time.Now().Add(time.Second)) {
		return session, apperrors.New(
			"this session has expired, please start the process again",
			http.StatusGone,
		)
	}

	if err := m.store.GetJson(kv.KeyMfaSession(sessionId), &session); err != nil {
		return models.MfaSession{}, err
	}

	return session, nil
}

func (m *mfaRepo) DeleteSession(sessionId string) error {
	return m.store.Delete(kv.KeyMfaSession(sessionId))
}

func (m *mfaRepo) CreateTotpEnrollmentSession(
	params *CreateTotpEnrollmentSessionParams,
) (models.MfaSession, error) {
	meta := models.TotpEnrollmentMeta{
		Name:      params.AccountName,
		RawKeyUrl: "",
	}

	sessionId := uuid.New()
	session := models.MfaSession{
		ID:          sessionId.String(),
		AccountType: queries.MfaAccountTypeTotp,
		SessionType: models.MfaSessionTypeTotpEnrollment,
		UserID:      params.User.ID,
		Meta:        &meta,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(MfaSessionDuration),
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      params.HostUrl,
		AccountName: fmt.Sprintf("%s (%s)", params.AccountName, params.User.Email),
	})
	if err != nil {
		return models.MfaSession{}, seer.Wrap("generate_totp_key", err)
	}
	meta.RawKeyUrl = key.String()

	if err := m.store.SetJsonWithTTL(
		kv.KeyTotpEnrolmentSession(sessionId.String()),
		session,
		MfaSessionDuration,
	); err != nil {
		return models.MfaSession{}, err
	}

	return session, nil
}

func (m *mfaRepo) FindTotpEnrollmentSession(sessionId string) (models.MfaSession, error) {
	var session models.MfaSession

	// Check if the session has expired
	expired, err := m.store.ExpiresAt(kv.KeyTotpEnrolmentSession(sessionId))
	if err != nil {
		if err == kv.ErrKeyNotFound {
			return session, apperrors.New(
				"invalid or expired session, please start the process again",
				http.StatusNotFound,
			)
		}

		return session, err
	}

	// If the session has expired, return false
	if expired.Before(time.Now().Add(time.Second)) {
		return session, apperrors.New(
			"this session has expired, please start the process again",
			http.StatusGone,
		)
	}

	if err := m.store.GetJson(kv.KeyTotpEnrolmentSession(sessionId), &session); err != nil {
		return models.MfaSession{}, err
	}

	return session, nil
}

// GenerateBackupCodes generates backup codes for a user to use in case they lose access to their MFA device
func (m *mfaRepo) GenerateBackupCodes(userId int32) ([]string, error) {
	if err := m.queries.DeleteAllBackupCodes(context.TODO(), userId); err != nil {
		log.Error().Err(err).Int32("user_id", userId).Msg("failed to delete existing backup codes")
		return nil, err
	}
	var codes []string
	for i := 0; i < 8; i++ {
		code, err := lib.GenerateToken("mfa_backup_code", 8)
		if err != nil {
			return nil, err
		}

		if err := m.queries.SaveBackupCodes(context.TODO(), queries.SaveBackupCodesParams{
			UserID:      userId,
			HashedToken: code.Encoded(),
		}); err != nil {
			log.Error().Err(err).Int32("user_id", userId).Msg("failed to save backup code")
			return nil, err
		}

		codes = append(codes, code.String())
	}

	return codes, nil
}

func (m *mfaRepo) VerifyBackupCode(userId int32, code string) error {
	codes, err := m.queries.FindBackupCodesByUserId(context.TODO(), userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Error().Err(err).Int32("user_id", userId).Msg("no backup codes found for this user")
			return apperrors.New(
				"unable to load backup codes, is MFA enabled for this user?",
				http.StatusBadRequest,
			)
		}
	}

	var found struct {
		id *int32
		mu sync.Mutex
	}

	eg := new(errgroup.Group)
	for _, backupCode := range codes {
		eg.Go(func() error {
			// Attempt to verify the code
			valid, err := lib.VerifyPassword(lib.VerifyPasswordParams{
				Password: code,
				Hash:     backupCode.HashedToken,
			})
			if err != nil {
				return err
			}

			// We don't want to return an error if the code is invalid, another one might be valid
			if !valid {
				return nil
			}

			// If the code has already been used, return an error
			if backupCode.UsedAt.Valid {
				return apperrors.New(
					"this backup code has already been used",
					http.StatusBadRequest,
				)
			}

			// Mark the code as used
			if err := m.queries.MarkBackupCodeAsUsed(context.TODO(),
				queries.MarkBackupCodeAsUsedParams{UserID: userId, CodeID: backupCode.ID},
			); err != nil {
				return err
			}

			found.mu.Lock()
			defer found.mu.Unlock()
			if found.id == nil {
				found.id = &backupCode.ID
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	if found.id == nil {
		return apperrors.New("invalid backup code provided", http.StatusBadRequest)
	}

	return nil
}

func (m *mfaRepo) CanGenerateBackupCodes(userId int32) (bool, error) {
	canGenerate, err := m.queries.CanGenerateBackupCodes(context.TODO(), userId)
	if err != nil {
		return false, err
	}

	return canGenerate, nil
}

var _ MfaRepository = (*mfaRepo)(nil)
