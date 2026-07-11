package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"techmind/internal/pkg/settings"
)

// PrometheusSnapshot 查询与告警类型匹配的聚合指标，控制返回量以适合 LLM 诊断输入。
func PrometheusSnapshot(ctx context.Context, alertName string) Evidence {
	baseURL := strings.TrimRight(settings.Conf.Monitor.PrometheusURL, "/")
	if baseURL == "" {
		return Evidence{"prometheus_status": "not configured"}
	}

	queries := map[string]string{
		"api_request_rate": "sum(rate(http_requests_total{path!=\"/metrics\"}[5m]))",
		"api_error_rate":   "sum(rate(http_errors_total[5m])) / clamp_min(sum(rate(http_requests_total{path!=\"/metrics\"}[5m])), 1)",
	}
	switch alertName {
	case "SearchLatencyHigh":
		queries["search_p95_seconds"] = "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{path=\"/api/v1/search\"}[5m])) by (le))"
	case "RedisStreamBacklogHigh":
		queries["redis_stream_pending"] = "sum(redis_stream_pending_total)"
	case "AICallFailureHigh":
		queries["ai_failure_rate"] = "sum(rate(ai_call_errors_total[5m]))"
	}

	evidence := Evidence{}
	for name, query := range queries {
		result, err := queryPrometheus(ctx, baseURL, query)
		if err != nil {
			evidence["prometheus_"+name+"_error"] = err.Error()
			continue
		}
		evidence["prometheus_"+name] = result
	}
	return evidence
}

func queryPrometheus(ctx context.Context, baseURL, query string) (string, error) {
	endpoint := baseURL + "/api/v1/query?query=" + url.QueryEscape(query)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	resp, err := (&http.Client{Timeout: 5 * time.Second}).Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("prometheus returned %s", resp.Status)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.Status != "success" {
		return "", fmt.Errorf("prometheus query status %q", payload.Status)
	}
	result := string(payload.Data.Result)
	if len(result) > 2000 {
		result = result[:2000] + "..."
	}
	return result, nil
}
