package email

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"

	"github.com/Gooowan/matchup/modules/core/types"
)

//go:embed templates/*.html
var templateFS embed.FS

type EmailProvider interface {
	GetName() string
	GetDomain() string
	SendRawEmail(ctx context.Context, req RawEmailRequest) error
}

type TemplateID string

const (
	PasswordResetTemplate TemplateID = "password-reset"
	OTPCodeTemplate       TemplateID = "otp-code"
)

type RawEmailRequest struct {
	To       string
	From     string
	Subject  string
	HTMLBody string
	TextBody string
}

type EmailService struct {
	provider  EmailProvider
	sender    string
	templates *template.Template
}

func NewEmailService(provider EmailProvider, customTemplates *embed.FS, sender string, requiredTemplates []TemplateID) (*EmailService, error) {
	// Parse default templates first
	templates, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse default email templates: %w", err)
	}

	// If custom templates are provided, parse and merge them (custom templates override defaults)
	if customTemplates != nil {
		customTmpl, err := template.ParseFS(*customTemplates, "templates/*.html")
		if err != nil {
			return nil, fmt.Errorf("failed to parse custom email templates: %w", err)
		}

		// Merge custom templates into the default templates
		// Templates with the same name in customTemplates will override defaults
		for _, tmpl := range customTmpl.Templates() {
			if tmpl.Name() != "" {
				// Clone the custom template and add it to the main templates
				// This will override any existing template with the same name
				_, err = templates.AddParseTree(tmpl.Name(), tmpl.Tree)
				if err != nil {
					return nil, fmt.Errorf("failed to merge custom template %s: %w", tmpl.Name(), err)
				}
			}
		}
	}

	// Validate required templates exist
	for _, templateID := range requiredTemplates {
		if templates.Lookup(string(templateID)) == nil {
			return nil, fmt.Errorf("required template not found: %s", templateID)
		}
	}

	return &EmailService{
		provider:  provider,
		templates: templates,
		sender:    sender,
	}, nil
}

func (s *EmailService) GetSender() string {
	return s.sender
}

func (s *EmailService) GetDomain() string {
	return s.provider.GetDomain()
}

func (s *EmailService) IsMockProvider() bool {
	return s.provider.GetName() == "mock"
}

func (s *EmailService) SendEmail(ctx context.Context, req EmailRequest) error {
	var htmlBody string
	var err error

	if req.TemplateID != "" {
		htmlBody, err = s.renderTemplate(req.TemplateID, req.TemplateData)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
	} else {
		htmlBody = req.HTMLBody
	}

	from := req.From
	if from == "" {
		from = fmt.Sprintf("noreply@%s", s.provider.GetDomain())
	}

	rawReq := RawEmailRequest{
		To:       req.To,
		From:     from,
		Subject:  req.Subject,
		HTMLBody: htmlBody,
		TextBody: req.TextBody,
	}

	return s.provider.SendRawEmail(ctx, rawReq)
}

func (s *EmailService) renderTemplate(templateID TemplateID, data types.JSONB) (string, error) {
	tmpl := s.templates.Lookup(string(templateID))
	if tmpl == nil {
		return "", fmt.Errorf("template not found: %s", templateID)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", templateID, err)
	}

	return buf.String(), nil
}

type EmailRequest struct {
	To      string
	From    string
	Subject string

	TemplateID   TemplateID  `json:"template_id,omitempty"`
	TemplateData types.JSONB `json:"template_data,omitempty"`

	HTMLBody string `json:"html_body,omitempty"`
	TextBody string `json:"text_body,omitempty"`
}
