package procedure

import (
	"time"

	"go.trulyao.dev/hubble/web/internal/ratelimit"
)

var RateLimits = map[string]ratelimit.Limit{
	"*":                   {MaxRequests: 75, Interval: 1 * time.Minute}, // default rate limit
	ListWorkspaceEntries:  {MaxRequests: 100, Interval: 1 * time.Minute},
	ListCollectionEntries: {MaxRequests: 100, Interval: 1 * time.Minute},

	SignIn:         {MaxRequests: 10, Interval: 1 * time.Hour},
	SignUp:         {MaxRequests: 25, Interval: 30 * time.Minute},
	VerifyEmail:    {MaxRequests: 10, Interval: 1 * time.Hour},
	ChangePassword: {MaxRequests: 10, Interval: 1 * time.Hour},

	RequestPasswordReset:     {MaxRequests: 5, Interval: 1 * time.Hour},
	RequestEmailVerification: {MaxRequests: 5, Interval: 1 * time.Hour},
	RequestEmailChange:       {MaxRequests: 12, Interval: 6 * time.Hour},
	VerifyEmailChange:        {MaxRequests: 12, Interval: 6 * time.Hour},

	MfaCreateEmailAccount:    {MaxRequests: 15, Interval: 30 * time.Minute},
	MfaActivateEmailAccount:  {MaxRequests: 15, Interval: 30 * time.Minute},
	MfaResendEmail:           {MaxRequests: 5, Interval: 1 * time.Hour},
	MfaInitiateAuth:          {MaxRequests: 10, Interval: 1 * time.Hour},
	MfaRegenerateBackupCodes: {MaxRequests: 5, Interval: 7 * 24 * time.Hour},
	MfaStartTotpEnrolment:    {MaxRequests: 15, Interval: 30 * time.Minute},
	MfaCompleteTotpEnrolment: {MaxRequests: 15, Interval: 30 * time.Minute},
}
