package worker

import (
	"testing"
	"time"

	"techmind/internal/pkg/settings"
)

func TestDiagnosisWindowUsesAlertAnchor(t *testing.T) {
	previous := settings.Conf.Ops
	settings.Conf.Ops.EvidenceWindowMin = 30
	t.Cleanup(func() { settings.Conf.Ops = previous })

	anchor := time.Now().Add(-time.Hour).Truncate(time.Second)
	start, end := diagnosisWindow(anchor)
	if !start.Equal(anchor.Add(-15*time.Minute)) || !end.Equal(anchor.Add(15*time.Minute)) {
		t.Fatalf("unexpected anchored window: %s..%s", start, end)
	}
}

func TestParseTaskTime(t *testing.T) {
	want := time.Date(2026, 7, 17, 8, 0, 0, 123, time.UTC)
	got := parseTaskTime(want.Format(time.RFC3339Nano))
	if !got.Equal(want) {
		t.Fatalf("parsed time = %s, want %s", got, want)
	}
	if got := parseTaskTime(nil); !got.IsZero() {
		t.Fatalf("nil time should be zero, got %s", got)
	}
}
