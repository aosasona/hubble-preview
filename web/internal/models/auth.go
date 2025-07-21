package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
)

type (
	AuthSession struct {
		UserID int32 `json:"user_id"`
		// Token is the Argon2 hashed version of the token.
		Token string `json:"token"`
		// Suspended indicates if the session has been suspended; this is different from a full revocation or logout in that the session can be re-enabled.
		Suspended bool `json:"suspended"`
		// TTL is a redundant field that contains the original TTL used to calculate the ExpiresAt field.
		TTL time.Duration
		// IssuedAt is the time the session was issued.
		IssuedAt time.Time `json:"issued_at"`
		// ExpiresAt is the time the session will expire.
		ExpiresAt time.Time `json:"expires_at"`
	}

	MfaAccount struct {
		ID           pgtype.UUID            `json:"id"            mirror:"type:string"`
		Type         queries.MfaAccountType `json:"type"          mirror:"type:'email' | 'totp'"`
		Meta         json.RawMessage        `json:"meta"          mirror:"type:{ name: string; email: string } | { name: string }"`
		Active       bool                   `json:"active"`
		UserID       int32                  `json:"user_id"`
		RegisteredAt pgtype.Timestamptz     `json:"registered_at" mirror:"type:string"`
		LastUsedAt   pgtype.Timestamptz     `json:"last_used_at"  mirror:"type:string"`
		Preferred    bool                   `json:"preferred"`
	}

	MfaState struct {
		Enabled            bool        `json:"enabled"`
		Accounts           MfaAccounts `json:"accounts"`
		PreferredAccountID pgtype.UUID `json:"preferred_account_id" mirror:"type:string"`
	}

	MfaSession struct {
		ID          string                 `json:"id"           mirror:"type:string"`
		AccountID   pgtype.UUID            `json:"account_id"   mirror:"type:string"`
		AccountType queries.MfaAccountType `json:"type"         mirror:"type:'email' | 'totp'"`
		SessionType MfaSessionType         `json:"session_type"`
		UserID      int32                  `json:"user_id"`
		Meta        MfaSessionMeta         `json:"meta"         mirror:"type:{ name?: string, email?: string }"`
		CreatedAt   time.Time              `json:"created_at"   mirror:"type:number"`
		ExpiresAt   time.Time              `json:"expires_at"   mirror:"type:number"`
	}
)

type (
	// MfaClientAccount represents an MFA account for the client with sensitive information removed
	MfaClientAccount struct {
		ID                   pgtype.UUID            `json:"id"        mirror:"type:string"`
		Name                 *string                `json:"name"`
		EmailAddress         string                 `json:"-"`
		RedactedEmailAddress *string                `json:"email"`
		Type                 queries.MfaAccountType `json:"type"      mirror:"type:'email' | 'totp'"`
		Preferred            bool                   `json:"preferred"`
	}

	MfaAccounts []MfaAccount

	MfaClientAccounts []MfaClientAccount
)

func (account MfaAccount) ToClientAccount(idx ...int) MfaClientAccount {
	clientAccount := MfaClientAccount{
		ID:                   account.ID,
		Name:                 lib.Ref(""),
		EmailAddress:         "",
		RedactedEmailAddress: lib.Ref(""),
		Type:                 account.Type,
		Preferred:            account.Preferred,
	}

	var accountType MfaSessionType
	if account.Type == queries.MfaAccountTypeTotp {
		accountType = MfaSessionTypeTotp
	} else {
		accountType = MfaSessionTypeEmail
	}

	meta, err := DecodeAccountMeta(accountType, account.Meta)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode account meta")
		return clientAccount
	}

	// If we have a name set, use that
	if meta.AccountName() != "" {
		clientAccount.Name = lib.Ref(meta.AccountName())
	} else {
		if len(idx) > 0 {
			clientAccount.Name = lib.Ref(fmt.Sprintf("Account %d", idx[0]+1))
		} else {
			clientAccount.Name = lib.Ref("Unnamed Account")
		}
	}

	switch meta := meta.(type) {
	case *EmailMeta:
		clientAccount.RedactedEmailAddress = lib.Ref(lib.RedactEmail(meta.Email, 3))
		clientAccount.EmailAddress = meta.Email
	}

	return clientAccount
}

func (accs MfaAccounts) ToClientAccounts(
	onPreferred ...func(*MfaClientAccount),
) MfaClientAccounts {
	var clientAccounts []MfaClientAccount
	for idx, account := range accs {
		clientAccount := account.ToClientAccount(idx)

		if account.Preferred && len(onPreferred) > 0 {
			onPreferred[0](&clientAccount)
		}

		clientAccounts = append(clientAccounts, clientAccount)
	}

	return clientAccounts
}

func (ca MfaClientAccounts) FindById(id pgtype.UUID) *MfaClientAccount {
	for _, account := range ca {
		if account.ID == id {
			return &account
		}
	}

	return nil
}

func DecodeAccountMeta(
	accountType MfaSessionType,
	rawMeta []byte,
) (MfaSessionMeta, error) {
	var meta MfaSessionMeta

	switch accountType {
	case MfaSessionTypeEmail:
		var emailMeta EmailMeta
		if err := json.Unmarshal(rawMeta, &emailMeta); err != nil {
			log.Error().
				Err(err).
				Bytes("meta", rawMeta).
				Msg("failed to unmarshal MFA account meta")
			return meta, err
		}
		meta = &emailMeta

	case MfaSessionTypeTotp:
		var totpMeta TotpMeta
		if err := json.Unmarshal(rawMeta, &totpMeta); err != nil {
			log.Error().
				Err(err).
				Bytes("meta", rawMeta).
				Msg("failed to unmarshal MFA account meta")
			return meta, err
		}
		meta = &totpMeta

	case MfaSessionTypeTotpEnrollment:
		var totpEnrollmentMeta TotpEnrollmentMeta
		if err := json.Unmarshal(rawMeta, &totpEnrollmentMeta); err != nil {
			log.Error().
				Err(err).
				Bytes("meta", rawMeta).
				Msg("failed to unmarshal MFA account meta")
			return meta, err
		}
		meta = &totpEnrollmentMeta

	default:
		return meta, apperrors.New(
			"invalid MFA account type",
			http.StatusInternalServerError,
		)
	}

	return meta, nil
}

func FromAccountType(accountType queries.MfaAccountType) MfaSessionType {
	switch accountType {
	case queries.MfaAccountTypeTotp:
		return MfaSessionTypeTotp
	case queries.MfaAccountTypeEmail:
		return MfaSessionTypeEmail
	default:
		panic("invalid account type, this should never happen")
	}
}
