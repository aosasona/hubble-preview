package mail

import (
	"html/template"

	"go.trulyao.dev/hubble/web/internal/mail/templates"
)

type NoopMailer struct{}

// Queue implements Mailer.
func (n *NoopMailer) Queue(_ string, _ template.Template, _ any) error {
	return nil
}

// Send implements Mailer.
func (n *NoopMailer) Send(_ string, _ templates.Template, _ any) error {
	return nil
}

func NewNoopMailer() *NoopMailer {
	return &NoopMailer{}
}
