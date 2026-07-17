package agent

import (
	"testing"
	"time"

	"techmind/internal/pkg/settings"
)

func TestNormalizeEvidenceWindowDefaultsAndCaps(t *testing.T) {
	previous := settings.Conf.Ops
	settings.Conf.Ops.EvidenceWindowMin = 30
	t.Cleanup(func() { settings.Conf.Ops = previous })

	end := time.Date(2026, 7, 17, 10, 0, 0, 0, time.UTC)
	start, gotEnd := normalizeEvidenceWindow(time.Time{}, end)
	if !gotEnd.Equal(end) || end.Sub(start) != 30*time.Minute {
		t.Fatalf("default window = %s..%s", start, gotEnd)
	}

	start, gotEnd = normalizeEvidenceWindow(end.Add(-2*time.Hour), end)
	if gotEnd.Sub(start) != time.Hour {
		t.Fatalf("window was not capped at one hour: %s", gotEnd.Sub(start))
	}
}
