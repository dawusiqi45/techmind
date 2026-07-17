package mcp

import (
	"strings"
	"testing"
)

func TestPrometheusRangeQueryMatchesAlert(t *testing.T) {
	tests := []struct {
		alert string
		want  string
	}{
		{"APILatencyHigh", "http_request_duration_seconds_bucket"},
		{"APIHighErrorRate", "http_errors_total"},
		{"RedisStreamBacklogHigh", "redis_stream_pending_total"},
		{"AICallFailureHigh", "ai_call_errors_total"},
	}
	for _, tt := range tests {
		t.Run(tt.alert, func(t *testing.T) {
			_, query := prometheusRangeQueryForAlert(tt.alert)
			if !strings.Contains(query, tt.want) {
				t.Fatalf("query %q does not contain %q", query, tt.want)
			}
		})
	}
}
