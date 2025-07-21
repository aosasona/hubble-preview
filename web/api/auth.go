package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/matthewhartstonge/argon2"
	"go.trulyao.dev/hubble/web/internal/database/queries"
	"go.trulyao.dev/hubble/web/internal/mail"
	"go.trulyao.dev/hubble/web/internal/mail/templates"
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/hubble/web/internal/otp"
	apperrors "go.trulyao.dev/hubble/web/pkg/errors"
	"go.trulyao.dev/hubble/web/pkg/lib"
	authlib "go.trulyao.dev/hubble/web/pkg/lib/auth"
	"go.trulyao.dev/robin"
	"go.trulyao.dev/robin/types"
)

type authHandler struct {
	*baseHandler
}

// Basic sign up request (email + password)
func (a *authHandler) SignUp(ctx *robin.Context, data SignUpRequest) (SignUpResponse, error) {
	var response SignUpResponse
	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	username := data.Username
	if lib.Empty(username) {
		prefix := strings.Split(data.Email, "@")[0]
		username = fmt.Sprintf("%s%d",
			strings.ToLower(lib.Substring(prefix, 0, 12)),
			lib.RandomInt(100, 999),
		)
	}

	argon := argon2.DefaultConfig()
	hashedPassword, err := argon.HashEncoded([]byte(data.Password))
	if err != nil {
		return response, err
	}

	lower := strings.ToLower

	params := queries.CreateUserParams{
		FirstName:     lower(data.FirstName),
		LastName:      lower(data.LastName),
		Email:         lower(data.Email),
		Username:      lower(username),
		PasswordHash:  string(hashedPassword),
		EmailVerified: !a.config.Flags.Email,
	}

	u, err := a.repos.UserRepository().CreateUser(params)
	if err != nil {
		return response, err
	}

	response.UserID = u.PublicID
	response.Username = u.Username
	response.AvailableMfaMethods = []MfaMethod{MfaMethodTotp}

	if a.config.Flags.Email {
		response.AvailableMfaMethods = append(response.AvailableMfaMethods, MfaMethodEmail)

		go a.sendEmailVerificationToken(u)
	}

	return response, nil
}

// Basic sign in request (email only)
func (a *authHandler) SignIn(ctx *robin.Context, data SignInRequest) (SignInResponse, error) {
	var response SignInResponse

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	response.Email = data.Email
	user, err := a.repos.UserRepository().FindUserByEmail(data.Email)
	if err != nil {
		return response, err
	}

	isValidPassword, err := lib.VerifyPassword(lib.VerifyPasswordParams{
		Password: data.Password,
		Hash:     user.HashedPassword,
	})
	if err != nil {
		return response, err
	}

	if !isValidPassword {
		return response, apperrors.New("invalid email or password", http.StatusUnauthorized)
	}

	// If the user has not verified their account and email verification is enabled, send another email and force them to verify
	if !user.EmailVerified && a.config.Flags.Email {
		response.RequiresEmailVerification = true
		go a.sendEmailVerificationToken(user)
		return response, nil
	}

	// Check if the user has MFA enabled
	mfaState, err := a.repos.MfaRepository().LoadState(user.ID)
	if err != nil {
		return response, err
	}

	if mfaState.Enabled {
		meta, err := initiateAuthMfaWithPreferredAccount(InitiateAuthMfaParams{
			request:    ctx.Request(),
			config:     a.config,
			mailer:     a.mailer,
			state:      &mfaState,
			user:       user,
			mfaRepo:    a.repos.MfaRepository(),
			userRepo:   a.repos.UserRepository(),
			otpManager: a.otpManager,
		})

		response.Mfa = MfaSignInData{Enabled: meta.Enabled, SessionId: meta.Session.ID}
		return response, err
	}

	// Generate a new session token
	_, err = generateAuthSession(generateAuthSessionParams{
		authRepo:     a.repos.AuthRepository(),
		userId:       user.ID,
		cookieSecret: a.config.Keys.CookieSecret,
		setCookie:    ctx.SetCookie,
		secure:       a.config.InProduction() || a.config.InStaging(),
	})
	if err != nil {
		return response, err
	}

	// Load workspaces
	workspaces, err := a.repos.WorkspaceRepository().FindAllByUserID(user.ID)
	if err != nil {
		return SignInResponse{}, err
	}

	response.User = user
	response.Workspaces = workspaces

	return response, nil
}

// SignOut handles sign out requests.
func (a *authHandler) SignOut(ctx *robin.Context, _ robin.Void) (robin.Void, error) {
	session, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return robin.Void{}, err
	}

	// Revoke the session
	if err := a.repos.AuthRepository().RevokeAuthSession(session.Token); err != nil {
		return robin.Void{}, err
	}

	// Remove the cookie
	ctx.SetCookie(&http.Cookie{
		Name:     authlib.CookieAuthSession,
		Value:    "",
		HttpOnly: true,
		Secure:   a.config.InProduction() || a.config.InStaging(),
		SameSite: http.SameSiteDefaultMode,
		Expires:  time.Now().Add(-time.Hour),
	})

	return robin.Void{}, nil
}

// VerifyEmail handles email verification requests.
func (a *authHandler) VerifyEmail(
	ctx *robin.Context,
	data EmailVerificationRequest,
) (EmailVerificationResponse, error) {
	response := EmailVerificationResponse{}

	if !a.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	user, err := a.repos.UserRepository().FindUserByEmail(data.Email)
	if err != nil {
		return response, err
	}

	if user.EmailVerified {
		return response, apperrors.New("account is already verified", http.StatusConflict)
	}

	err = a.otpManager.VerifyToken(
		otp.VerifyTokenParams{
			Type:          otp.TokenUserAccountVerification,
			Identifier:    user.ID,
			ProvidedToken: data.Token,
		},
	)
	if err != nil {
		return response, err
	}

	user, err = a.repos.UserRepository().VerifyEmail(user.ID)
	if err != nil {
		return response, err
	}

	response.Email = user.Email
	return response, nil
}

func (a *authHandler) ResendVerificationEmail(
	ctx *robin.Context,
	email string,
) (string, error) {
	if !a.config.Flags.Email {
		return "", apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	user, err := a.repos.UserRepository().FindUserByEmail(email)
	if err != nil {
		return "", err
	}

	if user.EmailVerified {
		return "", apperrors.New("email is already verified", http.StatusConflict)
	}

	tk, err := a.otpManager.GenerateToken(otp.TokenUserAccountVerification, user.ID)
	if err != nil {
		return "", err
	}

	validFor, err := a.otpManager.GetTokenTimeToLive(otp.TokenUserAccountVerification)
	if err != nil {
		return "", err
	}
	if err := a.mailer.Send(
		user.Email,
		templates.TemplateConfirmEmail,
		mail.TokenEmailParams{
			FirstName: lib.ToTitleCase(user.FirstName),
			Email:     user.Email,
			Code:      tk,
			Link:      fmt.Sprintf("%s/auth/verify-email?email=%s&token=%s", a.config.AppUrl, user.Email, tk),
			ValidFor:  validFor.Minutes(),
			SentAt:    lib.ToHumanReadableDate(time.Now()),
		},
	); err != nil {
		return "", err
	}

	return fmt.Sprintf("Verification email sent to %s", user.Email), nil
}

/*
SendPasswordResetEmail sends a password reset email.

This will also be used for a password change request.
*/

func (a *authHandler) SendPasswordResetEmail(
	ctx *types.Context,
	data SendPasswordResetEmailRequest,
) (SendPasswordResetEmailResponse, error) {
	var response SendPasswordResetEmailResponse

	if !a.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	user, err := a.repos.UserRepository().FindUserByEmail(data.Email)
	if err != nil {
		return response, err
	}

	token, err := a.otpManager.GenerateToken(otp.TokenUserPasswordReset, user.ID)
	if err != nil {
		return response, err
	}

	url := fmt.Sprintf(
		"%s/auth/change-password?email=%s&token=%s",
		a.config.AppUrl,
		user.Email,
		token,
	)
	if data.Scope == "change" {
		url += "&scope=change"
	}

	template := templates.TemplateResetPassword
	if data.Scope == "change" {
		template = templates.TemplateChangePassword
	}

	validFor, err := a.otpManager.GetTokenTimeToLive(otp.TokenUserPasswordReset)
	if err != nil {
		return response, err
	}

	if err := a.mailer.Send(
		user.Email,
		template,
		mail.TokenEmailParams{
			FirstName: lib.ToTitleCase(user.FirstName),
			Email:     user.Email,
			Code:      token,
			Link:      url,
			ValidFor:  validFor.Minutes(),
			SentAt:    lib.ToHumanReadableDate(time.Now()),
		},
	); err != nil {
		return response, err
	}

	response.Message = fmt.Sprintf("Password reset email sent to %s", user.Email)
	if data.Scope == "change" {
		response.Message = fmt.Sprintf(
			"We've sent an email to %s with instructions on how to change your password",
			user.Email,
		)
	}

	return response, nil
}

// ChangePassword handles password change requests.
// TODO: add support for password change when email services are disabled (e.g. using the current password)
func (a *authHandler) ChangePassword(
	ctx *robin.Context,
	data ChangePasswordRequest,
) (ChangePasswordResponse, error) {
	var (
		response ChangePasswordResponse
		user     *models.User
		err      error
	)

	data.EmailServicesEnabled = a.config.Flags.Email
	if err = lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	if user, err = a.repos.UserRepository().FindUserByEmail(data.Email); err != nil {
		return response, err
	}

	err = a.otpManager.VerifyToken(otp.VerifyTokenParams{
		Type:          otp.TokenUserPasswordReset,
		Identifier:    user.ID,
		ProvidedToken: data.Token,
	})
	if err != nil {
		return response, err
	}

	_, err = a.repos.UserRepository().ChangePassword(user.ID, data.NewPassword)
	if err != nil {
		return response, err
	}

	switch data.Scope {
	case "reset":
		response.Message = "Your password has been reset successfully"
	case "change":
		response.Message = "Your password has been changed successfully"
	}

	response.Scope = data.Scope
	return response, nil
}

func (a *authHandler) RequestEmailChange(
	ctx *robin.Context,
	data RequestEmailChangeRequest,
) (RequestEmailChangeResponse, error) {
	var response RequestEmailChangeResponse

	if !a.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	user, err := a.repos.UserRepository().FindUserByID(auth.UserID)
	if err != nil {
		return response, err
	}

	code, err := a.repos.UserRepository().RequestEmailChange(auth.UserID, data.Email)
	if err != nil {
		return response, err
	}

	validFor, err := a.otpManager.GetTokenTimeToLive(otp.TokenUserEmailChange)
	if err != nil {
		return response, err
	}
	if err := a.mailer.Send(
		data.Email,
		templates.TemplateChangeEmail,
		mail.TokenEmailParams{
			FirstName: lib.ToTitleCase(user.FirstName),
			Email:     data.Email,
			Code:      code,
			ValidFor:  validFor.Minutes(),
			SentAt:    lib.ToHumanReadableDate(time.Now()),
		},
	); err != nil {
		return response, err
	}

	response.Email = data.Email
	return response, nil
}

func (a *authHandler) VerifyEmailChange(
	ctx *robin.Context,
	data VerifyEmailChangeRequest,
) (EmailVerificationResponse, error) {
	var response EmailVerificationResponse

	if !a.config.Flags.Email {
		return response, apperrors.New("email services are disabled", http.StatusServiceUnavailable)
	}

	if err := lib.ValidateStruct(&data); err != nil {
		return response, err
	}

	auth, err := authlib.ExtractAuthSession(ctx)
	if err != nil {
		return response, err
	}

	if err := a.repos.UserRepository().VerifyEmailChange(auth.UserID, data.Email, data.Code); err != nil {
		return response, err
	}

	response.Email = data.Email
	return response, nil
}

var _ AuthHandler = (*authHandler)(nil)
