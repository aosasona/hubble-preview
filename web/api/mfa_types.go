package api

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/robin"
)

type MfaMethod string

const (
	MfaMethodEmail MfaMethod = "email"
	MfaMethodTotp  MfaMethod = "totp"
)

type MfaHandler interface {
	// GetUserMfaState retrieves the MFA settings for the user
	GetUserMfaState(
		ctx *robin.Context,
		data robin.Void,
	) (MfaStateResponse, error)

	// CreateEmailAccount creates an email account for MFA but does not activate it - it will indirectly initiate an MFA session
	CreateEmailAccount(
		ctx *robin.Context,
		data MfaEmailAccountRequest,
	) (MfaEmailAccountResponse, error)

	// StartTotpEnrollmentSession starts a TOTP enrollment session for the user to add a new authenticator account
	StartTotpEnrollmentSession(
		ctx *robin.Context,
		name string,
	) (MfaTotpEnrollmentSession, error)

	CompleteTotpEnrollment(
		ctx *robin.Context,
		data CompleteTotpEnrollmentRequest,
	) (MfaActivateAccountResponse, error)

	// ActivateEmailAccount activates an email account for MFA, this is similar to the verification process during an multi-factor authentication usage
	ActivateEmailAccount(
		ctx *robin.Context,
		data MfaActivateEmailAccountRequest,
	) (MfaActivateAccountResponse, error)

	ResendMfaEmail(
		ctx *robin.Context,
		data MfaResendEmailRequest,
	) (MfaResendEmailResponse, error)

	InitiateAuthSession(
		ctx *robin.Context,
		data MfaInitiateAuthRequest,
	) (MfaInitiateAuthResponse, error)

	VerifyAuthSession(
		ctx *robin.Context,
		data MfaVerifyRequest,
	) (MfaVerifyResponse, error)

	RenameAccount(
		ctx *robin.Context,
		data MfaRenameAccountRequest,
	) (MfaRenameAccountResponse, error)

	DeleteAccount(
		ctx *robin.Context,
		data MfaDeleteAccountRequest,
	) (MfaDeleteAccountResponse, error)

	SetPreferredAccount(ctx *robin.Context, accountId string) (PreferredAccountResponse, error)

	// FindSession finds an MFA session by its ID
	FindSession(ctx *robin.Context, sessinId string) (FindSessionResponse, error)

	// RegenerateBackupCodes regenerates the backup codes for the user
	RegenerateBackupCodes(ctx *robin.Context, _ robin.Void) ([]string, error)
}

type (
	MfaStateResponse struct {
		Enabled            bool                `json:"enabled"`
		Accounts           []models.MfaAccount `json:"accounts"`
		PreferredAccountID pgtype.UUID         `json:"preferred_account_id" mirror:"type:string"`
	}

	MfaMeta struct {
		// If the user has MFA enabled, this will be true
		Enabled bool `json:"enabled"`

		// If the user has MFA enabled, this will contain the current MFA session
		Session *models.MfaSession `json:"session"`

		// If the user has MFA enabled, this will contain the user's MFA accounts (e.g. email, TOTP)
		Accounts models.MfaClientAccounts `json:"accounts"`
	}

	MfaSignInData struct {
		Enabled   bool   `json:"enabled"`
		SessionId string `json:"session_id"`
	}
)

type (
	MfaVerifyRequest struct {
		// SessionID is the ID of the MFA session that the user is currently in
		SessionID string `json:"session_id"      validate:"required,uuid"`
		// Token is the token that the user is verifying with
		Code string `json:"code"            validate:"required,min=6,max=8"`
		// IsBackupCode is true if the user is verifying with a backup code instead of the session's code
		UseBackupCode bool `json:"use_backup_code"`
	}

	MfaVerifyResponse struct {
		Message string `json:"message"`

		// The user's workspaces (owned and shared)
		Workspaces []models.Workspace `json:"workspaces"`

		// The user's account
		User *models.User `json:"user"`
	}
)

type (
	MfaEmailAccountRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	MfaEmailAccountResponse struct {
		// SessionID is the ID that should be used to find the MFA session the user is currently in.
		SessionID string `json:"session_id"`

		// Email is the email address that the MFA account was created for
		Email string `json:"email"`
	}
)

type (
	MfaActivateEmailAccountRequest struct {
		SessionID string `json:"session_id" validate:"required"`
		Token     string `json:"token"      validate:"required,min=8,max=8"`
	}

	MfaActivateAccountResponse struct {
		// AccountID is the ID of the account that the MFA was activated for
		AccountID pgtype.UUID `json:"account_id" mirror:"type:string"`

		// BackupCodes are the backup codes that the user should store in a safe place - these are one-time use codes and are only present when the user just enabled MFA
		BackupCodes *[]string `json:"backup_codes"`
	}
)

type (
	MfaInitiateAuthRequest struct {
		AccountID string `json:"account_id" mirror:"type:string" validate:"required,uuid"`

		// PrevSessionID is the ID of the previous session that the user was in, this will be used to look up the user data. Ideally, this endpoint will never be accessed without first initiating a session via a normal sign-in
		PrevSessionID string `json:"prev_session_id" validate:"required"`
	}

	MfaInitiateAuthResponse struct {
		SessionID string `json:"session_id"`
	}
)

type MfaScope string

const (
	MfaScopeSetup MfaScope = "setup"
	MfaScopeLogin MfaScope = "login"
)

type (
	MfaResendEmailRequest struct {
		SessionID string   `json:"session_id" validate:"required"`
		Scope     MfaScope `json:"scope"      validate:"required,oneof=setup login" mirror:"type:'setup' | 'login'"`
	}

	MfaResendEmailResponse struct {
		SessionID string `json:"session_id"`
		Message   string `json:"message"`
	}
)

type (
	MfaRenameAccountRequest struct {
		AccountID string `json:"account_id" mirror:"type:string" validate:"required"`
		Name      string `json:"name"                            validate:"required,min=1,max=32,mixed_name"`
	}

	MfaRenameAccountResponse struct {
		AccountID pgtype.UUID `json:"account_id" mirror:"type:string"`
		Name      string      `json:"name"`
	}
)

type (
	MfaDeleteAccountRequest struct {
		AccountID string `json:"account_id" mirror:"type:string" validate:"required"`

		// NOTE: we will take a password here instead of a token since the user is deleting their account and may not have access to that account anymore
		Password string `json:"password" validate:"required"`
	}

	MfaDeleteAccountResponse struct {
		AccountID pgtype.UUID `json:"account_id" mirror:"type:string"`
	}
)

type PreferredAccountResponse struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
}

type FindSessionResponse struct {
	Session  models.MfaSession        `json:"session"`
	Accounts models.MfaClientAccounts `json:"accounts"`
}

type MfaTotpEnrollmentSession struct {
	SessionID string `json:"session_id"`
	Secret    string `json:"secret"`
	Image     string `json:"image"`
}

type (
	CompleteTotpEnrollmentRequest struct {
		SessionID string `json:"session_id" validate:"required,uuid"`
		Code      string `json:"code"       validate:"required,min=6,max=6"`
	}
)
