package providers

import (
	"context"
	"os"

	"github.com/Gooowan/matchup/modules/core/utils"
	"github.com/Gooowan/matchup/modules/email"
)

type MockProvider struct {
	domain string
}

func NewMockProvider() email.EmailProvider {
	domain := "example.com"
	res, ok := os.LookupEnv("EMAIL_DOMAIN")
	if ok {
		domain = res
	}
	return &MockProvider{domain: domain}
}

func (p *MockProvider) GetName() string {
	return "mock"
}

func (p *MockProvider) GetDomain() string {
	return p.domain
}

func (p *MockProvider) SendRawEmail(ctx context.Context, req email.RawEmailRequest) error {
	utils.DebugPrint("sending mock email")
	utils.DebugPrint("from: %s", req.From)
	utils.DebugPrint("to: %s", req.To)
	utils.DebugPrint("subject: %s", req.Subject)
	utils.DebugPrint("body: %s", req.HTMLBody)
	return nil
}
