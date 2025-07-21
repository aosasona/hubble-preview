package kv

import (
	"strconv"

	"go.trulyao.dev/hubble/web/pkg/lib"
)

const (
	NamespaceUser      namespace = "user"
	NamespaceSession   namespace = "session"
	NamespaceSystem    namespace = "system"
	NamespaceWorkspace namespace = "workspace"
	NamespaceEntry     namespace = "entry"
)

const (
	ColVerificationToken  collection = "verification_token"
	ColPasswordResetToken collection = "pw_reset_token"
	ColEmailChangeToken   collection = "email_change_token"
	ColEmailReservation   collection = "email_reservation" // When a user changes their email, the new email is reserved here until it is verified to ensure that another user does not try to change to that same email

	ColAuthSession          collection = "auth"
	ColMfaSession           collection = "mfa"
	ColTotpEnrolmentSession collection = "totp_enrolment"
	ColEmailMfaToken        collection = "email_mfa_token"

	ColRateLimit collection = "rate_limit"

	ColWorkspaceSlugId = "workspace_slug_id"
	ColUserInvite      = "user_invite"

	ColLinkMetadata = "url_metadata"
)

// Keys
var (
	// KeyEmailVerificationToken creates a key for an email verification token.
	KeyEmailVerificationToken = func(id int32) KeyContainer {
		return Key(NamespaceUser, ColVerificationToken, strconv.Itoa(int(id)))
	}

	// KeyPasswordResetToken creates a key for a password reset token.
	KeyPasswordResetToken = func(id int32) KeyContainer {
		return Key(NamespaceUser, ColPasswordResetToken, strconv.Itoa(int(id)))
	}

	// KeyEmailChangeToken creates a key for an email change token.
	KeyEmailChangeToken = func(id int32) KeyContainer {
		return Key(NamespaceUser, ColEmailChangeToken, strconv.Itoa(int(id)))
	}

	// KeyEmailReservation creates a key for an email reservation.
	KeyEmailReservation = func(email string) KeyContainer {
		return Key(NamespaceUser, ColEmailReservation, email)
	}

	// KeyAuthSession creates a key for a user's auth session.
	KeyAuthSession = func(token string) KeyContainer {
		return Key(NamespaceSession, ColAuthSession, token)
	}

	// KeyMfaSession creates a key for a user's MFA session.
	KeyMfaSession = func(sessionId string) KeyContainer {
		return Key(NamespaceSession, ColMfaSession, sessionId)
	}

	KeyEmailMfaToken = func(sessionId string) KeyContainer {
		return Key(NamespaceSession, ColEmailMfaToken, sessionId)
	}

	KeyTotpEnrolmentSession = func(sessionId string) KeyContainer {
		return Key(NamespaceSession, ColTotpEnrolmentSession, sessionId)
	}

	KeyRateLimiterState = func(scope string) KeyContainer {
		return Key(NamespaceSystem, ColRateLimit, "state::"+scope)
	}

	KeyWorkspaceSlugId = func(slug string) KeyContainer {
		return Key(NamespaceWorkspace, ColWorkspaceSlugId, slug)
	}

	KeyWorkspaceInvite = func(workspaceID int32, email string) KeyContainer {
		return Key(NamespaceWorkspace, ColUserInvite, strconv.Itoa(int(workspaceID))+"::"+email)
	}

	KeyLinkMetadata = func(url string) KeyContainer {
		url = lib.Slugify(url)
		return Key(NamespaceEntry, ColLinkMetadata, url)
	}
)
