package providers

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/Gooowan/matchup/modules/email"
)

type MailgunProvider struct {
	client mailgun.Mailgun
	domain string
}

func NewMailgunProvider() (email.EmailProvider, error) {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MAILGUN_API_KEY environment variable is required")
	}

	host := os.Getenv("MAILGUN_HOST")
	if host == "" {
		host = "https://api.eu.mailgun.net"
	}

	domain := os.Getenv("EMAIL_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("EMAIL_DOMAIN environment variable is required")
	}

	client := mailgun.NewMailgun(apiKey)
	client.SetAPIBase(host)

	return &MailgunProvider{
		client: client,
		domain: domain,
	}, nil
}

func (p *MailgunProvider) GetName() string {
	return "mailgun"
}

func (p *MailgunProvider) GetDomain() string {
	return p.domain
}

func (p *MailgunProvider) SendRawEmail(ctx context.Context, req email.RawEmailRequest) error {
	message := mailgun.NewMessage(req.From, req.Subject, req.TextBody, req.To)
	if req.HTMLBody != "" {
		message.SetHTML(req.HTMLBody)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := p.client.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("failed to send email via Mailgun: %w", err)
	}

	return nil
}
