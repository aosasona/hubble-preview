package api

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pquerna/otp/totp"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/mail"
	"go.trulyao.dev/hubble/web/internal/mail/templates"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/otp"
	"go.trulyao.dev/hubble/web/internal/repository"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/robin"
	"golang.org/x/sync/errgroup"
)

var (
	RegexEmailCode = regexp.MustCompile(`^[a-zA-Z0-9]{8}$`)
	RegexTotpCode  = regexp.MustCompile(`^[0-9]{6}$`)
)

var ErrInvalidTotpCode = apperrors.New(
	"The provided TOTP code is invalid, please try again",
	http.StatusBadRequest,
)

type mfaHandler struct {
	*baseHandler
}

func (m *mfaHandler) GetUserMfaState(
	ctx *robin.Context,
	data robin.Void,
) (MfaStateResponse, error) {
	var (
		response MfaStateResponse
		auth     models.AuthSession
		err      error
	)

	if auth, err = authlib.ExtractAuthSession(ctx); err != nil {
		return response, err
	}

	state, err := m.repos.MfaRepository().LoadState(auth.UserID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get MFA settings")
		return response, err
	}

	response.Enabled = state.Enabled
	response.Accounts = state.Accounts
	response.PreferredAccountID = state.PreferredAccountID

	return response, nil
}

/*
CreateEmailAccount creates an email account for MFA but does not activate it
*/
func (m *mfaHandler) CreateEmailAccount(
	ctx *robin.Context,
	data MfaEmailAccountRequest,
) (MfaEmailAccountResponse, error) {
	var (
		response  MfaEmailAccountResponse
		accountId pgtype.UUID
	)

	if !m.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	user, err := authlib.ExtractUser(ctx, m.repos.UserRepository())
	if err != nil {
		return response, err
	}

	// Attempt to find existing email MFA account
	account, err := m.repos.MfaRepository().FindAccountByEmail(data.Email)
	if err != nil {
		if !errors.Is(err, repository.ErrMfaAccountNotFound) {
			return response, err
		}

		// If the account does not exist, create it
		account, err = m.repos.MfaRepository().CreateEmailAccount(user.ID, data.Email)
		if err != nil {
			return response, err
		}
	}

	if account.Active || account.UserID != user.ID {
		return response, apperrors.NewValidationError(apperrors.ErrorMap{
			"email": []string{
				"This email has already been set up for multi-factor authentication",
			},
		})
	}
	accountId = account.ID

	// Create a new MFA session
	session, err := m.repos.MfaRepository().CreateSession(user.ID, accountId)
	if err != nil {
		return response, err
	}

	// Send MFA email code
	ipAddr, _ := lib.GetRequestIP(ctx.Request())
	if ipAddr == "" {
		log.Warn().Msg("failed to get IP address")
		ipAddr = "unknown"
	}
	if err := sendMfaEmailCode(SendMfaEmailCodeParams{
		config:     m.config,
		mailer:     m.mailer,
		userRepo:   m.repos.UserRepository(),
		otpManager: m.otpManager,

		email:   data.Email,
		user:    user,
		session: &session,
		reason:  mail.MfaReasonSetup,
		ipAddr:  ipAddr,
	}); err != nil {
		return response, err
	}

	response.Email = data.Email
	response.SessionID = session.ID

	return response, nil
}

func (m *mfaHandler) ActivateEmailAccount(
	ctx *robin.Context,
	data MfaActivateEmailAccountRequest,
) (MfaActivateAccountResponse, error) {
	var (
		user        *models.User
		response    MfaActivateAccountResponse
		backupCodes = make([]string, 0)

		err error
	)

	if !m.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	if user, err = authlib.ExtractUser(ctx, m.repos.UserRepository()); err != nil {
		return response, err
	}

	session, err := m.repos.MfaRepository().FindSession(data.SessionID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to find MFA session")
		return response, err
	}

	if _, err = m.requireEmailMfaSession(&session, user); err != nil {
		return response, err
	}

	// Check if the provided token is valid
	if err = m.otpManager.VerifyToken(otp.VerifyTokenParams{
		Type:          otp.TokenUserMfaVerification,
		Identifier:    session.ID,
		ProvidedToken: data.Token,
	}); err != nil {
		return response, err
	}

	// Load the mfa details and generate new backup codes if this is the first time
	mfaEnabled, err := m.repos.MfaRepository().IsEnabled(user.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to check if MFA is enabled")
		return response, err
	}

	// If MFA was not enabled (i.e. this is the first time), generate backup codes
	if !mfaEnabled {
		backupCodes, err = m.repos.MfaRepository().GenerateBackupCodes(user.ID)
		if err != nil {
			log.Error().Err(err).Msg("failed to generate backup codes")
			return response, err
		}

		response.BackupCodes = &backupCodes
	}

	// Activate the MFA Account
	if err := m.repos.MfaRepository().ActivateAccount(session.AccountID, user.ID); err != nil {
		log.Error().Err(err).Msg("failed to activate MFA account")
		return response, err
	}

	response.AccountID = session.AccountID
	return response, nil
}

func (m *mfaHandler) StartTotpEnrollmentSession(
	ctx *robin.Context,
	name string,
) (MfaTotpEnrollmentSession, error) {
	var (
		response MfaTotpEnrollmentSession
		user     *models.User
		err      error
	)

	if user, err = authlib.ExtractUser(ctx, m.repos.UserRepository()); err != nil {
		return response, err
	}

	// Check if the name is taken by another account
	taken, err := m.repos.MfaRepository().NameIsTaken(user.ID, name)
	if err != nil {
		return response, err
	}

	if taken {
		return response, apperrors.NewValidationError(apperrors.ErrorMap{
			"name": []string{"This name is already in use"},
		})
	}

	// Create a new enrollment session
	session, err := m.repos.MfaRepository().
		CreateTotpEnrollmentSession(&repository.CreateTotpEnrollmentSessionParams{
			User:        user,
			AccountName: name,
			HostUrl:     m.config.AppUrl,
		})
	if err != nil {
		return response, err
	}

	meta, ok := session.Meta.(*models.TotpEnrollmentMeta)
	if !ok {
		log.Error().
			Any("meta", session.Meta).
			Msg("invalid TOTP enrollment meta, expected type TotpEnrollmentMeta")
		return response, apperrors.New(
			"failed to create TOTP enrollment session",
			http.StatusInternalServerError,
		)
	}

	key, err := meta.Key()
	if err != nil {
		log.Error().Err(err).Msg("failed to get TOTP key from meta")
		return response, apperrors.New(
			"failed to create TOTP enrollment session",
			http.StatusInternalServerError,
		)
	}

	response.SessionID = session.ID
	response.Secret = key.Secret()

	qrCode, err := key.Image(250, 250)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate TOTP QR code")
		return response, apperrors.New(
			"failed to create TOTP enrollment session",
			http.StatusInternalServerError,
		)
	}

	imageBuffer := bytes.NewBuffer(nil)
	if err := png.Encode(imageBuffer, qrCode); err != nil {
		log.Error().Err(err).Msg("failed to encode TOTP QR code")
		return response, apperrors.New(
			"failed to create TOTP enrollment session",
			http.StatusInternalServerError,
		)
	}

	image := base64.StdEncoding.EncodeToString(imageBuffer.Bytes())
	response.Image = fmt.Sprintf("data:image/png;base64,%s", image)

	return response, nil
}

func (m *mfaHandler) CompleteTotpEnrollment(
	ctx *robin.Context,
	data CompleteTotpEnrollmentRequest,
) (MfaActivateAccountResponse, error) {
	var (
		response MfaActivateAccountResponse
		auth     models.AuthSession
		err      error
	)

	if err = lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	if auth, err = authlib.ExtractAuthSession(ctx); err != nil {
		return response, err
	}

	session, err := m.repos.MfaRepository().FindTotpEnrollmentSession(data.SessionID)
	if err != nil {
		return response, err
	}

	// Check if the session belongs to the user
	if session.UserID != auth.UserID {
		return response, apperrors.New(
			"you are not authorized to complete this action",
			http.StatusForbidden,
		)
	}

	// Check if the session is a TOTP enrollment session
	if session.AccountType != queries.MfaAccountTypeTotp ||
		session.SessionType != models.MfaSessionTypeTotpEnrollment {
		return response, apperrors.New(
			"invalid session type",
			http.StatusBadRequest,
		)
	}

	meta, ok := session.Meta.(*models.TotpEnrollmentMeta)
	if !ok {
		log.Error().Err(err).Msg("invalid TOTP enrollment meta, expected type TotpEnrollmentMeta")
		return response, apperrors.New(
			"failed to complete TOTP enrollment",
			http.StatusInternalServerError,
		)
	}

	key, err := meta.Key()
	if err != nil {
		log.Error().Err(err).Msg("failed to get TOTP key from meta")
		return response, apperrors.New(
			"failed to complete TOTP enrollment",
			http.StatusInternalServerError,
		)
	}

	// Validate the TOTP code
	if !totp.Validate(data.Code, key.Secret()) {
		return response, ErrInvalidTotpCode
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		// Load the mfa details and generate new backup codes if this is the first time
		mfaEnabled, err := m.repos.MfaRepository().IsEnabled(auth.UserID)
		if err != nil {
			log.Error().Err(err).Msg("failed to check if MFA is enabled")
			return err
		}

		// If MFA was not enabled (i.e. this is the first time), generate backup codes
		if !mfaEnabled {
			backupCodes, err := m.repos.MfaRepository().GenerateBackupCodes(auth.UserID)
			if err != nil {
				log.Error().Err(err).Msg("failed to generate backup codes")
				return err
			}

			response.BackupCodes = &backupCodes
		}

		return nil
	})

	eg.Go(func() error {
		// Create a new TOTP account
		account, err := m.repos.MfaRepository().CreateTotpAccount(auth.UserID, meta.AccountName())
		if err != nil {
			return err
		}

		// Encrypt the TOTP key
		encryptedKey, err := lib.EncryptAES(lib.EncryptAesParams{
			Key:       m.config.GetLatestTotpKey(),
			PlainText: []byte(key.Secret()),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to encrypt TOTP key")
			return apperrors.New(
				"failed to complete TOTP enrollment",
				http.StatusInternalServerError,
			)
		}

		if _, err := m.repos.TOTPRepository().CreateSecret(&models.TotpSecret{
			AccountID: account.ID,
			Hash:      encryptedKey,
			Version:   m.config.LatestTotpKeyVersion(),
		}); err != nil {
			log.Error().Err(err).Msg("failed to create TOTP secret")
			return apperrors.New(
				"failed to complete TOTP enrollment",
				http.StatusInternalServerError,
			)
		}

		response.AccountID = account.ID
		return nil
	})

	if err := eg.Wait(); err != nil {
		return response, err
	}

	return response, nil
}

// InitiateAuthSession initiates the MFA process
func (m *mfaHandler) InitiateAuthSession(
	ctx *robin.Context,
	data MfaInitiateAuthRequest,
) (MfaInitiateAuthResponse, error) {
	var response MfaInitiateAuthResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	prevSession, err := m.repos.MfaRepository().FindSession(data.PrevSessionID)
	if err != nil {
		return response, err
	}

	// Check if the user has MFA enabled
	state, err := m.repos.MfaRepository().LoadState(prevSession.UserID)
	if err != nil {
		return response, err
	}

	// If MFA is not enabled, return an error
	if !state.Enabled {
		return response, apperrors.New("MFA is not enabled", http.StatusForbidden)
	}

	var (
		account *models.MfaAccount
		user    *models.User

		eg = new(errgroup.Group)
	)

	// MARK: account
	eg.Go(func() error {
		accountId, err := lib.UUIDFromString(data.AccountID)
		if err != nil {
			return apperrors.New("invalid account ID", http.StatusBadRequest)
		}

		account, err = m.repos.MfaRepository().FindAccountById(accountId)
		return err
	})

	// MARK: user
	eg.Go(func() error {
		user, err = m.repos.UserRepository().FindUserByID(prevSession.UserID)
		return err
	})

	if err := eg.Wait(); err != nil {
		return response, err
	}

	// Send the email code if the preferred method is email
	ipAddr, _ := lib.GetRequestIP(ctx.Request())
	if ipAddr == "" {
		log.Warn().Msg("failed to get IP address")
		ipAddr = "unknown"
	}

	// Create a new MFA session
	session, err := m.repos.MfaRepository().CreateSession(prevSession.UserID, account.ID)
	if err != nil {
		return response, err
	}

	accounts := state.Accounts.ToClientAccounts()
	// Send email code if the account is an email account
	if account.Type == queries.MfaAccountTypeEmail {
		clientAccount := accounts.FindById(account.ID)
		if err := sendMfaEmailCode(SendMfaEmailCodeParams{
			config:     m.config,
			mailer:     m.mailer,
			userRepo:   m.repos.UserRepository(),
			otpManager: m.otpManager,

			email:   clientAccount.EmailAddress,
			user:    user,
			session: &session,
			reason:  mail.MfaReasonLogin,
			ipAddr:  ipAddr,
		}); err != nil {
			log.Error().Err(err).Msg("failed to send MFA email code")
			return response, apperrors.New(
				"we were unable to send the email, please try again later",
				http.StatusInternalServerError,
			)
		}
	}

	// Delete the previous auth session
	go func() {
		if err := m.repos.MfaRepository().DeleteSession(prevSession.ID); err != nil {
			log.Error().Err(err).Msg("failed to delete previous MFA session")
		}
	}()

	response.SessionID = session.ID
	return response, nil
}

// VerifyAuthSession verifies the MFA session for an authentication session
func (m *mfaHandler) VerifyAuthSession(
	ctx *robin.Context,
	data MfaVerifyRequest,
) (MfaVerifyResponse, error) {
	var response MfaVerifyResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	session, err := m.repos.MfaRepository().FindSession(data.SessionID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to find MFA session")
		return response, err
	}

	// -- MARK: Handle backup codes
	switch {
	case data.UseBackupCode:
		if !RegexEmailCode.MatchString(data.Code) {
			return response, apperrors.New("invalid backup code", http.StatusBadRequest)
		}

		if err := m.repos.MfaRepository().VerifyBackupCode(session.UserID, data.Code); err != nil {
			return response, err
		}

	// -- MARK: Handle email verification
	case session.AccountType == queries.MfaAccountTypeEmail:
		if !RegexEmailCode.MatchString(data.Code) {
			return response, apperrors.New("invalid backup code", http.StatusBadRequest)
		}

		// Check if the provided token is valid
		if err = m.otpManager.VerifyToken(otp.VerifyTokenParams{
			Type:          otp.TokenUserMfaVerification,
			Identifier:    session.ID,
			ProvidedToken: data.Code,
		}); err != nil {
			return response, err
		}

		go func() {
			if err := m.repos.MfaRepository().SetLastUsed(session.AccountID); err != nil {
				log.Error().Err(err).Msg("failed to update MFA session")
			}
		}()

	// -- MARK: Handle TOTP verification
	case session.AccountType == queries.MfaAccountTypeTotp:
		if !RegexTotpCode.MatchString(data.Code) {
			return response, apperrors.New("invalid backup code", http.StatusBadRequest)
		}

		// Load the TOTP secret
		secret, err := m.repos.TOTPRepository().FindSecretByAccountID(session.AccountID)
		if err != nil {
			return response, err
		}

		// Decrypt the TOTP secret
		plainTextSecret, err := lib.DecryptAES(lib.DecryptAesParams{
			Key:        m.config.GetLatestTotpKey(),
			CipherText: secret.Hash,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to decrypt TOTP secret")
			return response, apperrors.New(
				"unable to verify TOTP code at this time, please try again later",
				http.StatusInternalServerError,
			)
		}

		// Verify the TOTP code
		if !totp.Validate(data.Code, string(plainTextSecret)) {
			return response, ErrInvalidTotpCode
		}

		go func() {
			if err := m.repos.MfaRepository().SetLastUsed(session.AccountID); err != nil {
				log.Error().Err(err).Msg("failed to update MFA session")
			}
		}()
	}

	// Delete the MFA session
	go func() {
		if err := m.repos.MfaRepository().DeleteSession(session.ID); err != nil {
			log.Error().Err(err).Msg("failed to delete MFA session")
		}
	}()

	// Generate a new session token
	_, err = generateAuthSession(generateAuthSessionParams{
		authRepo:     m.repos.AuthRepository(),
		userId:       session.UserID,
		cookieSecret: m.config.Keys.CookieSecret,
		setCookie:    ctx.SetCookie,
		secure:       m.config.InProduction() || m.config.InStaging(),
	})
	if err != nil {
		return response, err
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		user, err := m.repos.UserRepository().FindUserByID(session.UserID)
		if err != nil {
			return err
		}

		response.User = user
		return nil
	})

	eg.Go(func() error {
		workspaces, err := m.repos.WorkspaceRepository().FindAllByUserID(session.UserID)
		if err != nil {
			return err
		}

		response.Workspaces = workspaces
		return nil
	})

	if err := eg.Wait(); err != nil {
		return response, err
	}

	response.Message = "Multi-factor authentication successful, redirecting..."

	return response, nil
}

func (m *mfaHandler) ResendMfaEmail(
	ctx *robin.Context,
	data MfaResendEmailRequest,
) (MfaResendEmailResponse, error) {
	var response MfaResendEmailResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	if !m.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	session, err := m.repos.MfaRepository().FindSession(data.SessionID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to find MFA session")
		return response, err
	}
	response.SessionID = session.ID

	user, err := m.repos.UserRepository().FindUserByID(session.UserID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to find user")
		return response, err
	}

	accountEmail, err := m.requireEmailMfaSession(&session, user)
	if err != nil {
		return response, err
	}

	ipAddr, _ := lib.GetRequestIP(ctx.Request())
	if ipAddr == "" {
		log.Warn().Msg("failed to get IP address")
		ipAddr = "unknown"
	}

	// Send MFA email code
	reason := mail.MfaReasonLogin
	if data.Scope == MfaScopeSetup {
		reason = mail.MfaReasonSetup
	}
	if err := sendMfaEmailCode(SendMfaEmailCodeParams{
		config:     m.config,
		mailer:     m.mailer,
		userRepo:   m.repos.UserRepository(),
		otpManager: m.otpManager,

		email:   accountEmail,
		user:    user,
		session: &session,
		reason:  reason,
		ipAddr:  ipAddr,
	}); err != nil {
		return response, err
	}

	response.Message = "A new code has been sent to " + lib.RedactEmail(accountEmail, 3)
	return response, nil
}

func (m *mfaHandler) RenameAccount(
	ctx *robin.Context,
	data MfaRenameAccountRequest,
) (MfaRenameAccountResponse, error) {
	var response MfaRenameAccountResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	accountId, err := lib.UUIDFromString(data.AccountID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to scan account ID")
		return response, apperrors.New("invalid account ID", http.StatusBadRequest)
	}

	account, err := m.repos.MfaRepository().FindAccountById(accountId)
	if err != nil {
		return response, err
	}

	if account.UserID != auth.UserID {
		return response, apperrors.New("account does not belong to user", http.StatusForbidden)
	}

	if err := m.repos.MfaRepository().RenameAccount(repository.RenameAccountParams{
		AccountId: account.ID,
		UserId:    auth.UserID,
		Name:      data.Name,
	}); err != nil {
		return response, err
	}

	response.AccountID = accountId
	response.Name = data.Name

	return response, nil
}

func (m *mfaHandler) DeleteAccount(
	ctx *robin.Context,
	data MfaDeleteAccountRequest,
) (MfaDeleteAccountResponse, error) {
	var response MfaDeleteAccountResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	user, err := authlib.ExtractUser(ctx, m.repos.UserRepository())
	if err != nil {
		return response, err
	}

	// Verify that the password is valid
	var isValidPassword bool
	if isValidPassword, err = lib.VerifyPassword(lib.VerifyPasswordParams{
		Password: data.Password,
		Hash:     user.HashedPassword,
	}); err != nil {
		return response, err
	}

	if !isValidPassword {
		return response, apperrors.New("incorrect password provided", http.StatusUnauthorized)
	}

	accountId, err := lib.UUIDFromString(data.AccountID)
	if err != nil {
		log.Debug().Err(err).Msg("failed to convert account ID to valid UUID")
		return response, apperrors.New("invalid account ID", http.StatusBadRequest)
	}

	if err := m.repos.MfaRepository().DeleteAccount(accountId, user.ID); err != nil {
		return response, err
	}

	response.AccountID = accountId
	return response, nil
}

func (m *mfaHandler) SetPreferredAccount(
	ctx *robin.Context,
	accountId string,
) (PreferredAccountResponse, error) {
	var response PreferredAccountResponse

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	accountIdUUID, err := lib.UUIDFromString(accountId)
	if err != nil {
		return response, apperrors.New("invalid account ID", http.StatusBadRequest)
	}

	account, err := m.repos.MfaRepository().SetPreferredAccount(accountIdUUID, auth.UserID)
	if err != nil {
		return response, err
	}

	response.AccountID = account.ID.String()

	accountType := models.FromAccountType(account.Type)
	meta, _ := models.DecodeAccountMeta(accountType, account.Meta)
	if meta != nil && meta.AccountName() != "" {
		response.Name = meta.AccountName()
	}

	return response, nil
}

func (m *mfaHandler) FindSession(
	ctx *robin.Context,
	sessionId string,
) (FindSessionResponse, error) {
	var response FindSessionResponse

	if sessionId == "" {
		return response, apperrors.New(
			"session ID not provided",
			http.StatusBadRequest,
		)
	}

	session, err := m.repos.MfaRepository().FindSession(sessionId)
	if err != nil {
		return response, err
	}

	state, err := m.repos.MfaRepository().LoadState(session.UserID)
	if err != nil {
		return response, err
	}

	response.Session = session
	response.Accounts = state.Accounts.ToClientAccounts()

	return response, nil
}

func (m *mfaHandler) RegenerateBackupCodes(
	ctx *robin.Context,
	_ robin.Void,
) ([]string, error) {
	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return nil, err
	}

	// Ensure that the user has MFA enabled
	state, err := m.repos.MfaRepository().LoadState(auth.UserID)
	if err != nil {
		return nil, err
	}

	if !state.Enabled {
		return nil, apperrors.New(
			"multi-factor authentication is not enabled!",
			http.StatusForbidden,
		)
	}

	canGenerateBackupCodes, err := m.repos.MfaRepository().CanGenerateBackupCodes(auth.UserID)
	if err != nil {
		return nil, err
	}

	if !canGenerateBackupCodes {
		return nil, apperrors.New(
			"you cannot regenerate backup codes at this time, please try again later",
			http.StatusForbidden,
		)
	}

	backupCodes, err := m.repos.MfaRepository().GenerateBackupCodes(auth.UserID)
	if err != nil {
		return nil, err
	}

	return backupCodes, nil
}

type SendMfaEmailCodeParams struct {
	config     *config.Config
	mailer     mail.Mailer
	otpManager otp.Manager
	userRepo   repository.UserRepository

	email   string
	user    *models.User
	session *models.MfaSession
	reason  mail.MfaRequirementReason
	ipAddr  string
}

// sendMfaEmailCode sends an MFA email code to the user for verification
func sendMfaEmailCode(params SendMfaEmailCodeParams) error {
	if !params.config.Flags.Email {
		log.Warn().Msg("email services are disabled")
		return apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	token, err := params.otpManager.GenerateToken(otp.TokenUserMfaVerification, params.session.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate MFA email code")
		return err
	}

	validFor, err := params.otpManager.GetTokenTimeToLive(otp.TokenUserMfaVerification)
	if err != nil {
		return err
	}
	err = params.mailer.Send(
		params.email,
		templates.TemplateEmailMfa,
		mail.MfaEmailParams{
			TokenEmailParams: mail.TokenEmailParams{
				FirstName: lib.ToTitleCase(params.user.FirstName),
				Email:     params.email,
				Code:      token,
				Link: fmt.Sprintf(
					"%s/mfa/verify?email=%s&token=%s",
					params.config.AppUrl,
					params.email,
					token,
				),
				ValidFor: validFor.Minutes(),
				SentAt:   lib.ToHumanReadableDate(time.Now()),
			},
			Reason: params.reason,
			IpAddr: params.ipAddr,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to send MFA email code")
		return err
	}

	return nil
}

func (m *mfaHandler) requireEmailMfaSession(
	session *models.MfaSession,
	user *models.User,
) (string, error) {
	// Ensure the session belongs to the user
	if session.UserID != user.ID {
		log.Debug().Msg("session does not belong to user")
		return "", apperrors.New(
			"this MFA session does not belong to the requesting user",
			http.StatusForbidden,
		)
	}

	// Ensure the account is an email account
	if session.AccountType != queries.MfaAccountTypeEmail {
		log.Debug().Msg("session is not an email account")
		return "", apperrors.New(
			"this MFA session is not attached to an email account",
			http.StatusForbidden,
		)
	}

	meta, ok := session.Meta.(*models.EmailMeta)
	if !ok {
		log.Error().Any("meta", session.Meta).Msg("invalid MFA session meta")
		return "", apperrors.New(
			"session is broken, please start the process again",
			http.StatusInternalServerError,
		)
	}

	return meta.Email, nil
}

var _ MfaHandler = (*mfaHandler)(nil)
