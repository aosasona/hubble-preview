package api

import (
	"fmt"
	"net/http"
	"time"

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
)

type InitiateAuthMfaParams struct {
	request    *http.Request
	config     *config.Config
	mailer     mail.Mailer
	state      *models.MfaState
	user       *models.User
	otpManager otp.Manager

	mfaRepo  repository.MfaRepository
	userRepo repository.UserRepository
}

func initiateAuthMfaWithPreferredAccount(
	params InitiateAuthMfaParams,
) (*MfaMeta, error) {
	var (
		meta MfaMeta

		preferredAccount *models.MfaClientAccount
	)

	if len(params.state.Accounts) == 0 {
		return &meta, apperrors.New("no MFA accounts found", http.StatusBadRequest)
	}

	accounts := params.state.Accounts.ToClientAccounts(func(account *models.MfaClientAccount) {
		preferredAccount = account
	})

	// Create a new MFA session
	mfaSession, err := params.mfaRepo.CreateSession(params.user.ID, preferredAccount.ID)
	if err != nil {
		return &meta, err
	}

	// Send the email code if the preferred method is email
	ipAddr, _ := lib.GetRequestIP(params.request)
	if ipAddr == "" {
		log.Warn().Msg("failed to get IP address")
		ipAddr = "unknown"
	}

	if preferredAccount.Type == queries.MfaAccountTypeEmail {
		if err := sendMfaEmailCode(SendMfaEmailCodeParams{
			config:     params.config,
			mailer:     params.mailer,
			userRepo:   params.userRepo,
			otpManager: params.otpManager,

			email:   preferredAccount.EmailAddress,
			user:    params.user,
			session: &mfaSession,
			reason:  mail.MfaReasonLogin,
			ipAddr:  ipAddr,
		}); err != nil {
			log.Error().Err(err).Msg("failed to send MFA email code")
			return &meta, apperrors.New(
				"we were unable to send the MFA email code, please try again later",
				http.StatusInternalServerError,
			)
		}
	}

	meta = MfaMeta{
		Enabled:  params.state.Enabled,
		Session:  &mfaSession,
		Accounts: accounts,
	}

	return &meta, nil
}

// sendEmailVerificationToken sends an email verification token to the user.
func (a *authHandler) sendEmailVerificationToken(u *models.User, force ...bool) {
	if !a.config.Flags.Email {
		log.Warn().Msg("email services are disabled")
		return
	}

	// If the email is already verified, don't try to another email
	if u.EmailVerified {
		return
	}

	// If force is true, delete the existing token and generate a new one (useful for testing)
	if len(force) > 0 && force[0] {
		if err := a.otpManager.DeleteToken(otp.TokenUserAccountVerification, u.ID); err != nil {
			log.Error().Err(err).Msg("failed to delete email verification token")
			return
		}
	}

	token, err := a.otpManager.GenerateToken(otp.TokenUserAccountVerification, u.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate email verification token")
		return
	}

	validFor, _ := a.otpManager.GetTokenTimeToLive(otp.TokenUserAccountVerification)
	if err := a.mailer.Send(
		u.Email,
		templates.TemplateConfirmEmail,
		mail.TokenEmailParams{
			FirstName: lib.ToTitleCase(u.FirstName),
			Email:     u.Email,
			Code:      token,
			Link:      fmt.Sprintf("%s/auth/verify-email?email=%s&token=%s", a.config.AppUrl, u.Email, token),
			ValidFor:  validFor.Minutes(),
			SentAt:    lib.ToHumanReadableDate(time.Now()),
		},
	); err != nil {
		log.Error().Err(err).Msg("failed to send email verification email")
		return
	}
}

type generateAuthSessionParams struct {
	authRepo     repository.AuthRepository
	userId       int32
	cookieSecret string
	setCookie    func(cookie *http.Cookie)
	secure       bool
}

func generateAuthSession(params generateAuthSessionParams) (*models.AuthSession, error) {
	// Generate a new session token
	session, err := params.authRepo.GenerateAuthSession(params.userId)
	if err != nil {
		return nil, err
	}

	// Set the session token as part of the cookie
	signedCookie := authlib.SignCookie(authlib.SignCookieArgs{
		Name:   authlib.CookieAuthSession,
		Value:  session.Token,
		Secret: params.cookieSecret,
	})
	params.setCookie(&http.Cookie{
		Name:     authlib.CookieAuthSession,
		Value:    signedCookie.Base64EncodedValue(),
		HttpOnly: true,
		Secure:   params.secure,
		SameSite: http.SameSiteDefaultMode,
		Path:     "/",
		Expires:  time.Now().Add(repository.AuthSessionTTL),
	})

	return &session, nil
}
