package api

import (
	"go.trulyao.dev/hubble/web/internal/models"
	"go.trulyao.dev/robin"
)

type AuthHandler interface {
	SignIn(*robin.Context, SignInRequest) (SignInResponse, error)

	SignUp(*robin.Context, SignUpRequest) (SignUpResponse, error)

	SignOut(*robin.Context, robin.Void) (robin.Void, error)

	VerifyEmail(*robin.Context, EmailVerificationRequest) (EmailVerificationResponse, error)

	ResendVerificationEmail(*robin.Context, string) (string, error)

	SendPasswordResetEmail(
		*robin.Context,
		SendPasswordResetEmailRequest,
	) (SendPasswordResetEmailResponse, error)

	ChangePassword(*robin.Context, ChangePasswordRequest) (ChangePasswordResponse, error)

	// ChangeEmail is used to create a new email change request (i.e. sending the OTP)
	RequestEmailChange(
		*robin.Context,
		RequestEmailChangeRequest,
	) (RequestEmailChangeResponse, error)

	// VerifyEmailChange is used to verify a new email address
	VerifyEmailChange(*robin.Context, VerifyEmailChangeRequest) (EmailVerificationResponse, error)
}

type (
	SignUpRequest struct {
		FirstName string `json:"first_name" validate:"required,min=2,max=50,alpha"`
		LastName  string `json:"last_name"  validate:"required,min=2,max=50,alpha"`
		Username  string `json:"username"   validate:"required,username,min=2,max=50"`
		Email     string `json:"email"      validate:"required,email"`
		Password  string `json:"password"   validate:"required,min=8,max=50"`
	}

	SignUpResponse struct {
		UserID              string      `json:"user_id"               mirror:"type:string"`
		Username            string      `json:"username"`
		AvailableMfaMethods []MfaMethod `json:"available_mfa_methods" mirror:"type:Array<'email' | 'totp'>"`
	}
)

type (
	SignInRequest struct {
		Email    string `json:"email"    validate:"required,email"`
		Password string `json:"password" validate:"required,min=8,max=50"`
	}

	SignInResponse struct {
		// The user's email
		Email string `json:"email"`

		// If the user has not verified their email, this will be true
		RequiresEmailVerification bool `json:"requires_email_verification"`

		// The user's workspaces (owned and shared)
		Workspaces []models.Workspace `json:"workspaces"`

		// Multi-factor authentication metadata
		Mfa MfaSignInData `json:"mfa"`

		// The user's account
		User *models.User `json:"user"`
	}
)

type (
	VerifyEmailChangeRequest struct {
		Email string `json:"email" validate:"required,email"`
		Code  string `json:"code"  validate:"required,min=8,max=8"`
	}

	EmailVerificationRequest struct {
		Email string `json:"email" validate:"required,email"`
		Token string `json:"token" validate:"required,min=8,max=8"`
	}

	EmailVerificationResponse struct {
		Email string `json:"email"`
	}
)

type (
	SendPasswordResetEmailRequest struct {
		Email string `json:"email" validate:"required,email"`
		Scope string `json:"scope" validate:"required,oneof=reset change" mirror:"type:'reset' | 'change'"`
	}

	SendPasswordResetEmailResponse struct {
		Message string `json:"message"`
	}
)

type (
	ChangePasswordRequest struct {
		Email string `json:"email" validate:"required,email"`

		EmailServicesEnabled bool   `json:"-"`
		Token                string `json:"token"            validate:"required_if=EmailServicesEnabled true,min=8,max=8"`
		CurrentPassword      string `json:"current_password" validate:"required_if=EmailServicesEnabled false,min=0,max=50"`

		NewPassword     string `json:"new_password"     validate:"required,min=8,max=50"`
		ConfirmPassword string `json:"confirm_password" validate:"required,min=8,max=50,eqfield=NewPassword"`
		Scope           string `json:"scope"            validate:"required,oneof=reset change"               mirror:"type:'reset' | 'change'"`
	}

	ChangePasswordResponse struct {
		Message string `json:"message"`
		Scope   string `json:"scope"`
	}
)

type (
	RequestEmailChangeRequest struct {
		Email string `json:"email" validate:"required,email"`
	}

	RequestEmailChangeResponse struct {
		Email string `json:"email"`
	}
)
