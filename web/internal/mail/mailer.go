package mail

import (
	"html/template"
	"net/smtp"

	"github.com/domodwyer/mailyak/v3"
	"go.trulyao.dev/hubble/web/internal/config"
	"go.trulyao.dev/hubble/web/internal/mail/templates"
	"go.trulyao.dev/seer"
)

type (
	Mailer interface {
		Queue(to string, name template.Template, data any) error
		Send(to string, name templates.Template, data any) error
	}
)

type (
	DefaultMailer struct {
		config *config.SmtpConfig
	}

	DefaultMailerParams struct {
		SmtpConfig *config.SmtpConfig
	}
)

type (
	// TokenEmailParams contains the parameters required for token-related emails.
	TokenEmailParams struct {
		FirstName string
		Email     string
		Code      string
		Link      string
		ValidFor  float64
		SentAt    string
	}

	MfaRequirementReason string

	MfaEmailParams struct {
		TokenEmailParams
		Reason MfaRequirementReason
		IpAddr string
	}

	WorkspaceInviteEmailParams struct {
		// Email is the email of the recipient
		Email string

		// InviterName is the name of the user who invited the recipient
		InitiatorName string

		// WorkspaceName is the name of the workspace
		WorkspaceName string

		// Host is the hostname of the Hubble instance
		Host string

		// InviteID is the ID of the invite
		InviteID string

		// SentAt is the token of the invite
		SentAt string
	}
)

const (
	MfaReasonLogin MfaRequirementReason = "sign in to your account"
	MfaReasonSetup MfaRequirementReason = "set up multi-factor authentication"
)

func NewDefaultMailer(params *DefaultMailerParams) (*DefaultMailer, error) {
	return &DefaultMailer{config: params.SmtpConfig}, nil
}

func (mailer *DefaultMailer) client() *mailyak.MailYak {
	return mailyak.New(
		mailer.config.Addr(),
		smtp.PlainAuth(
			"",
			mailer.config.Username,
			mailer.config.Password,
			mailer.config.Host,
		),
	)
}

// QueueMail queues an email to be sent.
//
// TODO: implement email queueing
func (mailer *DefaultMailer) Queue(to string, name template.Template, data any) error {
	panic("unimplemented")
}

// SendMail sends an email instantly.
//
// WARNING: This should not be used for bulk emails
func (mailer *DefaultMailer) Send(to string, template templates.Template, data any) error {
	client := mailer.client()

	client.To(to)
	client.From(mailer.config.From())
	client.FromName("Hubble")
	client.Subject(template.Title())

	if err := templates.Execute(client.HTML(), template, data); err != nil {
		return seer.Wrap("execute_template", err)
	}

	if err := client.Send(); err != nil {
		return seer.Wrap("send_mail", err)
	}

	return nil
}

var (
	_ Mailer = (*DefaultMailer)(nil)
	_ Mailer = (*NoopMailer)(nil)
)
