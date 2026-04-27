package chat

import (
	"regexp"
)

var denyPatterns = []*regexp.Regexp{
	// Phone numbers (common formats)
	regexp.MustCompile(`\b\d{3}[\s.\-]\d{3,4}[\s.\-]\d{4}\b`),
	// Telegram / WhatsApp / Instagram handles solicitation
	regexp.MustCompile(`(?i)@[a-z0-9_]{3,}`),
	// Explicit slurs and harassment (keep list minimal — add as needed)
	regexp.MustCompile(`(?i)\b(kys|kill\s*yourself)\b`),
}

// isMessageAllowed returns false and a reason if content should be blocked.
func isMessageAllowed(content string) (allowed bool, reason string) {
	for _, re := range denyPatterns {
		if re.MatchString(content) {
			return false, "message contains prohibited content"
		}
	}
	return true, ""
}
