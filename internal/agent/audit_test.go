package agent

import (
	"strings"
	"testing"

	"techmind/internal/agent/mcp"
)

func TestSanitizeAuditEvidence(t *testing.T) {
	output := sanitizeAuditEvidence(mcp.Evidence{
		"authorization": "Authorization: Bearer secret-token-value",
		"nested": map[string]interface{}{
			"password": "super-secret",
		},
	})

	if got := output["authorization"].(string); strings.Contains(got, "secret-token-value") {
		t.Fatalf("bearer credential was not redacted: %q", got)
	}
	nested := output["nested"].(map[string]interface{})
	if got := nested["password"].(string); strings.Contains(got, "super-secret") {
		t.Fatalf("named secret was not redacted: %q", got)
	}
}

func TestSanitizeAuditStringTruncates(t *testing.T) {
	value := strings.Repeat("a", maxAuditTextLength+10)
	got := sanitizeAuditString(value)
	if !strings.HasSuffix(got, "... [truncated]") {
		t.Fatalf("expected truncated suffix, got %q", got[len(got)-20:])
	}
}
