package mcp

import (
	"strings"
	"testing"
)

func TestRedactLog(t *testing.T) {
	input := "normal line\nAuthorization: Bearer abc\napi_key=secret-value\ncookie=session"
	got := redactLog(input)
	if strings.Contains(got, "abc") || strings.Contains(got, "secret-value") || strings.Contains(got, "session") {
		t.Fatalf("sensitive values were not redacted: %q", got)
	}
	if !strings.Contains(got, "normal line") {
		t.Fatalf("non-sensitive log was removed: %q", got)
	}
}

func TestMatchesServiceRejectsUnrelatedPod(t *testing.T) {
	labels := map[string]string{"app.kubernetes.io/component": "worker"}
	if matchesService("techmind-worker-abc", labels, "techmind-server") {
		t.Fatal("unrelated worker pod matched server service")
	}
	if !matchesService("techmind-worker-abc", labels, "techmind-worker") {
		t.Fatal("worker pod did not match worker service")
	}
}
