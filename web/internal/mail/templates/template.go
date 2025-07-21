package templates

import (
	"bytes"
	"embed"
	"html/template"
	"io"

	"github.com/rs/zerolog/log"
	"go.trulyao.dev/seer"
)

//go:generate go tool github.com/abice/go-enum --marshal

//go:embed *.html
var emailTemplates embed.FS

// ENUM(confirm_email,reset_password,change_password,email_mfa,change_email,invite_user_to_workspace)
type Template int

func (t Template) Title() string {
	switch t {
	case TemplateConfirmEmail:
		return "Confirm Email"
	case TemplateResetPassword:
		return "Reset Password"
	case TemplateChangePassword:
		return "Change Password"
	case TemplateEmailMfa:
		return "Two-Factor Authentication"
	case TemplateChangeEmail:
		return "Change Email"
	case TemplateInviteUserToWorkspace:
		return "Workspace Invitation"
	default:
		return "Hubble"
	}
}

// GetTemplateContent returns the content of the template (without the layout).
func GetTemplateContent(name Template) ([]byte, error) {
	if !name.IsValid() {
		return nil, ErrInvalidTemplate
	}

	return emailTemplates.ReadFile(name.String() + ".html")
}

// GetTemplateWithLayout returns the template with the layout applied.
func GetTemplateWithLayout(name Template, data any) ([]byte, error) {
	output := bytes.NewBuffer([]byte{})
	if err := Execute(output, name, data); err != nil {
		log.Error().Err(err).Msg("failed to render template in GetTemplateWithLayout")
		return nil, seer.Wrap("render_template", err)
	}

	return output.Bytes(), nil
}

func Execute(output io.Writer, name Template, data any) error {
	var err error
	t := template.New("layout.html")

	if t, err = t.ParseFS(emailTemplates, name.String()+".html", "layout.html"); err != nil {
		log.Error().Err(err).Msg("failed to parse email layout in Execute")
		return seer.Wrap("parse_layout", err)
	}

	if err := t.Execute(output, data); err != nil {
		log.Error().Err(err).Msg("failed to execute template in Execute")
		return seer.Wrap("execute_template", err)
	}

	return nil
}
