package providers

import (
	"context"
	"fmt"
	"os"

	"github.com/Gooowan/matchup/modules/email"
	"github.com/resend/resend-go/v2"
)

type ResendProvider struct {
	client *resend.Client
	domain string
}

func NewResendProvider() (email.EmailProvider, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("RESEND_API_KEY environment variable is required")
	}

	domain := os.Getenv("EMAIL_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("EMAIL_DOMAIN environment variable is required")
	}

	client := resend.NewClient(apiKey)

	return &ResendProvider{
		client: client,
		domain: domain,
	}, nil
}

func (p *ResendProvider) GetName() string {
	return "resend"
}

func (p *ResendProvider) GetDomain() string {
	return p.domain
}

func (p *ResendProvider) SendRawEmail(ctx context.Context, req email.RawEmailRequest) error {
	_, err := p.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    req.From,
		To:      []string{req.To},
		Subject: req.Subject,
		Html:    req.HTMLBody,
	})
	if err != nil {
		return fmt.Errorf("failed to send email via Resend: %w", err)
	}

	return nil
}
